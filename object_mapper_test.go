package objectmapper

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_mapStruct(t *testing.T) {
	type Source struct {
		Name string
		Age  int
	}

	type Target struct {
		Name string
		Age  int
	}

	source := Source{Name: "John", Age: 23}
	target := Target{}
	Map(source, &target)

	expected := Target{Name: "John", Age: 23}
	assert.Equal(t, expected, target)
}

func Test_mapStructIgnoreNonPresentFields(t *testing.T) {
	type Source struct {
		Name       string
		Age        int
		Profession string
		HasPets    bool
	}

	type Target struct {
		Name string
		Age  int
	}

	source := Source{Name: "John", Age: 30, Profession: "engineer", HasPets: true}
	target := Target{}
	err := Map(source, &target)
	assert.Nil(t, err)

	expected := Target{Name: "John", Age: 30}
	assert.Equal(t, expected, target)
}

func Test_mapStructWithTime(t *testing.T) {
	type Source struct {
		Created     time.Time
		Name        string
		Description string
	}

	type Target struct {
		Created time.Time
		Name    string
	}

	// TODO CHANGE FOR ACTUAL VALUE
	time, _ := time.Parse(time.RFC3339, time.RFC3339)

	source := Source{Name: "Foo", Description: "Bar", Created: time}
	target := Target{}
	err := Map(source, &target)
	assert.Nil(t, err)

	expected := Target{Name: "Foo", Created: time}
	assert.Equal(t, expected, target)
}

func Test_mapStructWithTimePointer(t *testing.T) {
	type Source struct {
		Deleted     *time.Time
		Name        string
		Description string
	}

	type Target struct {
		Deleted *time.Time
		Name    string
	}

	// Non-nil time
	func() {
		time, _ := time.Parse(time.RFC3339, time.RFC3339)

		source := Source{Name: "Foo", Description: "Bar", Deleted: &time}
		target := Target{}
		err := Map(source, &target)
		assert.Nil(t, err)

		expected := Target{Name: "Foo", Deleted: &time}
		assert.Equal(t, expected, target)
	}()

	// Nil time
	func() {
		source := Source{Name: "Foo", Description: "Bar", Deleted: nil}
		target := Target{}
		err := Map(source, &target)
		assert.Nil(t, err)

		expected := Target{Name: "Foo", Deleted: nil}
		assert.Equal(t, expected, target)
	}()
}

func Test_mapStructWithStringTargetCoercion(t *testing.T) {
	type Source struct {
		ID       int
		IsActive bool
		Score    float64
	}

	type Target struct {
		ID       string
		IsActive string
		Score    string
	}

	source := Source{ID: 1000, IsActive: true, Score: 0.96}
	target := Target{}
	Map(source, &target)

	expected := Target{ID: "1000", IsActive: "true", Score: "0.96"}
	assert.Equal(t, expected, target)
}

func Test_mapStructWithPointers(t *testing.T) {
	type Source struct {
		Value *int
		Tag   *string
		Score float32
	}

	type Target struct {
		Value *int
		Tag   *string
		Score *float32
	}

	sourceValue := 2
	sourceTag := "test"
	source := Source{&sourceValue, &sourceTag, 32.45}

	target := Target{}
	err := Map(source, &target)
	assert.Nil(t, err)

	targetValue := 2
	targetTag := "test"
	var targetScore float32 = 32.45
	expected := Target{&targetValue, &targetTag, &targetScore}
	assert.Equal(t, expected, target)
}

func Test_mapStructWithPointersIgnoreNonPresentFields(t *testing.T) {
	type Source struct {
		Value  *int
		Tag    *string
		Code   *string
		Weight *float64
	}

	type Target struct {
		Value *int
		Tag   *string
	}

	sourceValue := 2
	sourceTag := "test-tag"
	sourceCode := "my-code"
	sourceWeight := 10.5
	source := Source{&sourceValue, &sourceTag, &sourceCode, &sourceWeight}

	target := Target{}
	err := Map(source, &target)
	assert.Nil(t, err)

	targetValue := 2
	targetTag := "test-tag"
	expected := Target{&targetValue, &targetTag}
	assert.Equal(t, expected, target)
}

func Test_mapNestedStruct(t *testing.T) {
	type SourceChild struct {
		Value int
		Tag   string
	}
	type Source struct {
		Name  string
		Child SourceChild
	}

	type TargetChild struct {
		Value int
	}
	type Target struct {
		Name  string
		Child TargetChild
	}

	source := Source{Name: "John", Child: SourceChild{Value: 10, Tag: "mytag"}}
	target := Target{}
	err := Map(source, &target)
	assert.Nil(t, err)

	expected := Target{Name: "John", Child: TargetChild{Value: 10}}
	assert.Equal(t, expected, target)
}

func Test_mapNestedStructWithPointers(t *testing.T) {
	type SourceChild struct {
		Value int
		Tag   string
	}
	type Source struct {
		Name        string
		FirstChild  *SourceChild
		SecondChild *SourceChild
	}

	type TargetChild struct {
		Value int
	}
	type Target struct {
		Name       string
		FirstChild *TargetChild
	}

	source := Source{Name: "John", FirstChild: &SourceChild{Value: 10, Tag: "mytag"}, SecondChild: &SourceChild{Value: 20, Tag: "mytag"}}
	target := Target{}
	err := Map(source, &target)
	assert.Nil(t, err)

	expected := Target{Name: "John", FirstChild: &TargetChild{Value: 10}}
	assert.Equal(t, expected, target)
}

func Test_mapDeepNestedStructWithPointers(t *testing.T) {
	type Person struct {
		ID          int
		Name        string
		Child       *Person
		SecretField *[]byte
	}

	type PersonData struct {
		ID    int
		Name  string
		Child *PersonData
	}

	secret := []byte("my-secret-1")
	source := Person{
		ID:          1,
		Name:        "John",
		SecretField: &secret,
		Child: &Person{
			ID:          2,
			Name:        "Peter",
			SecretField: &secret,
			Child: &Person{
				ID:          3,
				Name:        "Sarah",
				SecretField: nil,
				Child:       nil,
			},
		},
	}
	target := PersonData{}
	err := Map(source, &target)
	assert.Nil(t, err)

	expected := PersonData{
		ID:   1,
		Name: "John",
		Child: &PersonData{
			ID:   2,
			Name: "Peter",
			Child: &PersonData{
				ID:    3,
				Name:  "Sarah",
				Child: nil,
			},
		},
	}
	assert.Equal(t, expected, target)
}

func Test_mapSliceOfStructs(t *testing.T) {
	type Country struct {
		Name         string
		Population   int
		MainLanguage string
	}

	type Regions struct {
		Name string
	}

	countries := []Country{
		{Name: "Argentina", Population: 45, MainLanguage: "Español"},
		{Name: "USA", Population: 330, MainLanguage: "English"},
		{Name: "Deutschland", Population: 83, MainLanguage: "Deutsch"},
	}

	regions := []Regions{}
	err := Map(countries, &regions)
	assert.Nil(t, err)

	expected := []Regions{{"Argentina"}, {"USA"}, {"Deutschland"}}
	assert.Equal(t, expected, regions)
}

func Test_mapPointerToSliceOfStructs(t *testing.T) {
	type Country struct {
		Name         string
		Population   int
		MainLanguage string
	}

	type Regions struct {
		Name string
	}

	countries := []Country{
		{Name: "Argentina", Population: 45, MainLanguage: "Español"},
		{Name: "USA", Population: 330, MainLanguage: "English"},
		{Name: "Deutschland", Population: 83, MainLanguage: "Deutsch"},
	}

	regions := []Regions{}
	err := Map(&countries, &regions)
	assert.Nil(t, err)

	expected := []Regions{{"Argentina"}, {"USA"}, {"Deutschland"}}
	assert.Equal(t, expected, regions)
}

// Custom type for testing
type level string

func (c *level) Points() int {
	if *c == "intermediate" {
		return 50
	} else if *c == "advanced" {
		return 100
	}

	return 10
}

func Test_mapStructWithSourceCustomTypes(t *testing.T) {
	type Source struct {
		Username string
		Level    level
	}

	type Target struct {
		Username string
		Level    string
	}

	source := Source{Username: "demo.username", Level: "intermediate"}
	target := Target{}
	err := Map(source, &target)
	assert.Nil(t, err)

	expected := Target{Username: "demo.username", Level: "intermediate"}
	assert.Equal(t, expected, target)
}

// Custom types
type JSONStr string

func (f *JSONStr) MarshalJSON() ([]byte, error) {
	if f == nil {
		return []byte("[]"), nil
	}
	arr := make([]struct {
		ID    int    `json:"id"`
		Label string `json:"label"`
		Value int    `json:"value"`
	}, 0)
	json.Unmarshal([]byte(string(*f)), &arr)
	return json.Marshal(arr)
}

func Test_mapStructWithTargetCustomString2(t *testing.T) {
	type Source struct {
		Username string
		Items    string
	}

	type Target struct {
		Username string  `json:"username"`
		Items    JSONStr `json:"items"`
	}

	source := Source{Username: "demo.username", Items: "[{ \"id\": 23, \"label\": \"Some label\", \"value\": 100 }]"}
	target := Target{}

	// A converter function must be registered for the target type.
	// If it's not registered we will panic (TODO: error handling)
	assert.Panics(t, func() {
		err := Map(source, &target)
		assert.Nil(t, err)
	})

	err := MapWithBuilders(source, &target, map[string]structBuilderFn{
		"objectmapper.JSONStr": func(value interface{}) interface{} {
			strValue := value.(string)
			return JSONStr(strValue)
		},
	})
	assert.Nil(t, err)

	expected := Target{Username: "demo.username", Items: JSONStr("[{ \"id\": 23, \"label\": \"Some label\", \"value\": 100 }]")}
	assert.Equal(t, expected, target)

	expectedSerialized, err := json.Marshal(&target)
	assert.Equal(t,
		"{\"username\":\"demo.username\",\"items\":[{\"id\":23,\"label\":\"Some label\",\"value\":100}]}",
		string(expectedSerialized),
	)
}

// Error handling
func Test_returnsErrWhenNilParam(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}
	var e *ParametersError

	// Nil target
	source := Person{Name: "John", Age: 23}
	err := Map(source, nil)
	assert.Error(t, err)
	assert.ErrorAs(t, err, &e)

	// Nil source
	target := Person{Name: "John", Age: 23}
	err = Map(nil, target)
	assert.Error(t, err)
	assert.ErrorAs(t, err, &e)
}

func Test_returnsErrWhenTargetNotPointer(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}

	type Employee struct {
		Name string
	}

	source := Person{Name: "John", Age: 23}
	target := Employee{Name: "John"}

	err := Map(source, target)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrTargetParamNotPointer)

	assert.Contains(t, err.Error(), "must be a pointer")
}

func Test_returnsErrWhenMapStructToSlice(t *testing.T) {
	type Country struct {
		Name         string
		Population   int
		MainLanguage string
	}

	type Regions struct {
		Name string
	}

	country := Country{Name: "Argentina", Population: 45, MainLanguage: "Español"}

	regions := []Regions{}
	err := Map(country, &regions)
	assert.Error(t, err)
}
