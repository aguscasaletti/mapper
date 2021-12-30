package mapper

import (
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/fatih/structtag"
)

type TypeConverterFn func(interface{}) interface{}

var defaultTypeConvertMap = map[string]TypeConverterFn{
	"time.Time": func(value interface{}) interface{} {
		timeValue := value.(time.Time)
		parsedTime, _ := time.Parse(time.RFC3339, timeValue.Format(time.RFC3339))
		return parsedTime
	},
}

func validateParameters(source interface{}, target interface{}) error {
	if target == nil {
		return NewParamErrorNotNil("target")
	}
	if source == nil {
		return NewParamErrorNotNil("source")
	}

	if reflect.ValueOf(target).Kind() != reflect.Ptr {
		return ErrTargetParamNotPointer
	}

	return nil
}

// Map - map values from source to target
func Map(source, target interface{}) error {
	return MapWithConverters(source, target, defaultTypeConvertMap)
}

// MapWithConverters - map values from source to target, and use converter functions passed
// 	when the default behavior is not enough
func MapWithConverters(source, target interface{}, converters map[string]TypeConverterFn) error {
	if err := validateParameters(source, target); err != nil {
		return err
	}

	// merge maps
	converterFnMap := make(map[string]TypeConverterFn, 0)
	for k, v := range defaultTypeConvertMap {
		converterFnMap[k] = v
	}
	for k, v := range converters {
		converterFnMap[k] = v
	}

	targetValue := reflect.Indirect(reflect.ValueOf(target))
	_, err := mapValues(reflect.ValueOf(source), targetValue, &converterFnMap)
	return err
}

// mapValues - recursively map values from one object to another using reflection
func mapValues(sourceValue reflect.Value, targetValue reflect.Value, converters *map[string]TypeConverterFn) (interface{}, error) {
	switch targetValue.Kind() {
	case reflect.Ptr:
		return mapToPointer(sourceValue, targetValue, converters)
	case reflect.Struct:
		return mapToStruct(sourceValue, targetValue, converters)
	case reflect.Slice:
		return mapToSlice(sourceValue, targetValue, converters)
	case reflect.String:
		return mapToString(sourceValue, targetValue)
	case reflect.Invalid:
		log.Println("mapping invalid value", targetValue)
	default:
		if targetValue.CanSet() {
			targetValue.Set(sourceValue)
		}
	}

	return targetValue.Interface(), nil
}

// getSourceFieldValue - Gets the source field value with the following rules:
//  - if a mapper tag exists AND has a fromField property, use that
//  - if a mapper tag exists AND has a fromMethod property, invoke that method and use that
//  - else return the source struct's field value (if any)
//  - if no field is present return a Zero value that will fail an IsValid() check
func getSourceFieldValue(sourceStruct reflect.Value, targetStructField reflect.StructField) reflect.Value {
	tag := targetStructField.Tag
	tags, _ := structtag.Parse(string(tag))

	if mapperTag, _ := tags.Get("mapper"); mapperTag != nil {
		for _, setting := range strings.Split(mapperTag.Value(), ";") {
			switch {
			case strings.HasPrefix(setting, "fromField:"):
				sourceFieldName := strings.Split(setting, ":")[1]
				return sourceStruct.FieldByName(sourceFieldName)
			case strings.HasPrefix(setting, "fromMethod"):
				sourceMethodName := strings.Split(setting, ":")[1]

				// Search struct receiver. E.g: func (s PersonTest) GetFullName() string
				method := sourceStruct.MethodByName(sourceMethodName)
				if !method.IsValid() {
					// Search pointer receiver. E.g: func (s *PersonTest) GetFullName() string
					ptr := reflect.New(sourceStruct.Type())
					ptr.Elem().Set(sourceStruct)
					method = ptr.MethodByName(sourceMethodName)
				}

				if method.IsValid() {
					values := method.Call(make([]reflect.Value, 0))
					if len(values) > 0 {
						return values[0]
					}
				}
			}
		}
	}

	return sourceStruct.FieldByName(targetStructField.Name)
}

func mapToStruct(sourceValue, targetValue reflect.Value, converters *map[string]TypeConverterFn) (interface{}, error) {
	numFields := targetValue.NumField()

	// Indirect the source value in case it's a pointer to a struct, and not a struct
	sourceValue = reflect.Indirect(sourceValue)

	for i := 0; i < numFields; i++ {
		targetField := targetValue.Type().Field(i)
		targetFieldValue := targetValue.FieldByName(targetField.Name)
		sourceFieldValue := getSourceFieldValue(sourceValue, targetField)

		// E.g: the field does not exist or is not exported
		// check CanInterface to see if sourceFieldValue is exported or not
		// we IGNORE unexported source fields
		if !sourceFieldValue.IsValid() || !sourceFieldValue.CanInterface() {
			continue
		}

		var newValue interface{}
		// If we have a function to create a value of the target type, use it
		if fn, ok := (*converters)[targetFieldValue.Type().String()]; ok {
			newValue = fn(sourceFieldValue.Interface())
		} else {
			var err error
			newValue, err = mapValues(sourceFieldValue, targetFieldValue, converters)
			if err != nil {
				return nil, NewFieldError(targetField.Name, "invalid field projection", err)
			}
		}

		// if the new value is nil then we don't need to set anything and thus we move on
		if newValue == nil {
			continue
		}

		// if the target field is a pointer, but mapValues only returns actual values (not pointers)
		// then we should wrap this new value into a pointer to be set into targetFieldValue
		if targetFieldValue.Kind() == reflect.Ptr {
			wrapper := reflect.New(reflect.TypeOf(newValue))
			wrapper.Elem().Set(reflect.ValueOf(newValue))
			targetFieldValue.Set(wrapper)
		} else {
			targetFieldValue.Set(reflect.ValueOf(newValue))
		}
	}

	return targetValue.Interface(), nil
}

func mapToPointer(sourceValue, targetValue reflect.Value, converters *map[string]TypeConverterFn) (interface{}, error) {
	// If source value is a Zero value, there's no value to be copied
	if sourceValue.IsZero() {
		return nil, nil
	}

	// Indirect the source value in case it's a pointer to a struct, and not a struct
	sourceIndirectValue := reflect.Indirect(sourceValue)

	var newValue interface{}
	if fn, ok := (*converters)[targetValue.Type().Elem().String()]; ok {
		newValue = fn(sourceIndirectValue.Interface())
	} else {
		// we want to create an artificial target value that
		//  is NOT a pointer AND IS addressable/settable
		// so that we can build a value recursively
		// and after that set a pointer to this new value to the original target
		targetArtificialValue := reflect.New(targetValue.Type().Elem())
		newValue, _ = mapValues(sourceIndirectValue, targetArtificialValue.Elem(), converters)
	}

	// return the actual value (not a pointer, to avoid returning a *interface{} type)
	return newValue, nil
}

func mapToString(sourceValue, targetValue reflect.Value) (interface{}, error) {
	// attempt conversion to string
	var sourceValueStr string = fmt.Sprintf("%v", sourceValue.Interface())
	if targetValue.CanSet() {
		targetValue.Set(reflect.ValueOf(sourceValueStr))
	}

	return targetValue.Interface(), nil
}

func mapToSlice(sourceValue, targetValue reflect.Value, converters *map[string]TypeConverterFn) (interface{}, error) {
	if !sourceValue.IsValid() {
		return nil, nil
	}

	sourceValue = reflect.Indirect(sourceValue)
	if sourceValue.Kind() != reflect.Slice {
		return nil, fmt.Errorf("cannot map to a slice from type: %v", sourceValue.Type().String())
	}

	numItems := sourceValue.Len()
	targetSlice := reflect.MakeSlice(targetValue.Type(), numItems, numItems)
	for i := 0; i < numItems; i++ {
		mapValues(sourceValue.Index(i), targetSlice.Index((i)), converters)
	}

	targetValue.Set(reflect.ValueOf(targetSlice.Interface()))
	return targetValue.Interface(), nil
}
