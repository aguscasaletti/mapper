package objectmapper

import (
	"fmt"
	"log"
	"reflect"
	"time"
)

type structBuilderFn func(interface{}) interface{}

var defaultTypeTransformFnMap = map[string]structBuilderFn{
	"time.Time": func(value interface{}) interface{} {
		timeValue := value.(time.Time)
		parsedTime, _ := time.Parse(time.RFC3339, timeValue.Format(time.RFC3339))
		return parsedTime
	},
}

func verifyParameters(source interface{}, target interface{}) error {
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
func Map(source interface{}, target interface{}) error {
	return MapWithBuilders(source, target, defaultTypeTransformFnMap)
}

// MapWithBuilders - map values from source to target, and use builder functions passed
func MapWithBuilders(source interface{}, target interface{}, builders map[string]structBuilderFn) error {
	if err := verifyParameters(source, target); err != nil {
		return err
	}

	// shallow copy TODO remove
	var builderMap = defaultTypeTransformFnMap

	// merge maps
	for k, v := range builders {
		builderMap[k] = v
	}

	targetValue := reflect.Indirect(reflect.ValueOf(target))
	_, err := mapValues(reflect.ValueOf(source), targetValue, builderMap)
	return err
}

// mapValues - recursively map values from one object to another using reflection
func mapValues(sourceValue reflect.Value, targetValue reflect.Value, builders map[string]structBuilderFn) (interface{}, error) {
	switch targetValue.Kind() {
	case reflect.Ptr:
		// If source value is a Zero value, there's no value to be copied
		if sourceValue.IsZero() {
			return nil, nil
		}
		// else get the actual source value
		sourceIndirectValue := reflect.Indirect(sourceValue)

		var newValue interface{}
		if fn, ok := defaultTypeTransformFnMap[targetValue.Type().Elem().String()]; ok {
			newValue = fn(sourceIndirectValue.Interface())
		} else {
			// we want to create an artificial target value that
			//  is NOT a pointer AND IS addressable/settable
			// so that we can build a value recursively
			// and after that set a pointer to this new value to the original target
			targetArtificialValue := reflect.New(targetValue.Type().Elem())
			newValue, _ = mapValues(sourceIndirectValue, targetArtificialValue.Elem(), builders)
		}

		// return the actual value (not a pointer, to avoid returning a *interface{} type)
		return newValue, nil

	case reflect.Struct:
		// Copy all field values from source to target
		numFields := targetValue.NumField()
		for i := 0; i < numFields; i++ {
			targetField := targetValue.Type().Field(i)
			targetFieldValue := targetValue.FieldByName(targetField.Name)
			sourceFieldValue := sourceValue.FieldByName(targetField.Name)

			if !sourceValue.IsValid() {
				continue
			}

			var newValue interface{}

			// If we have a function to create a value of the target type, use it
			if fn, ok := defaultTypeTransformFnMap[targetFieldValue.Type().String()]; ok {
				newValue = fn(sourceFieldValue.Interface())
			} else {
				var err error
				newValue, err = mapValues(sourceFieldValue, targetFieldValue, builders)
				if err != nil {
					return nil, NewFieldError(targetField.Name, "invalid field projection", err)
				}
			}

			// if the new value is nil then we don't need to set anything
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

	case reflect.Slice:
		// log.Println(targetValue.Type(), sourceValue.Type(), sourceValue.Len())
		if !sourceValue.IsValid() {
			return nil, nil
		}

		// initialize slice
		sourceValue = reflect.Indirect(sourceValue)
		if sourceValue.Kind() != reflect.Slice {
			return nil, fmt.Errorf("cannot map to a slice from type: %v", sourceValue.Type().String())
		}

		numItems := sourceValue.Len()
		targetSlice := reflect.MakeSlice(targetValue.Type(), numItems, numItems)
		for i := 0; i < numItems; i++ {
			mapValues(sourceValue.Index(i), targetSlice.Index((i)), builders)
		}

		targetValue.Set(reflect.ValueOf(targetSlice.Interface()))

	case reflect.String:
		// attempt conversion to string
		var sourceValueStr string = fmt.Sprintf("%v", sourceValue.Interface())
		if targetValue.CanSet() {
			targetValue.Set(reflect.ValueOf(sourceValueStr))
		}
	case reflect.Invalid:
		log.Println("invalid value", targetValue)
	default:
		if targetValue.CanSet() {
			targetValue.Set(sourceValue)
		}
	}

	return targetValue.Interface(), nil
}
