package mapper

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func buildLargeTestStruct() Large {
	return Large{
		Name:   "Test",
		Age:    30,
		Value:  false,
		Score:  394.3,
		Score2: 14.2,
		Map: map[string]string{
			"somekey": "someval",
		},
		Child1: LargeChild{
			Items: []LargeChildItem{
				{
					ID:    "43",
					Value: true,
				},
				{
					ID:    "44",
					Value: false,
				},
			},
		},
		Child2: LargeChild{
			Items: []LargeChildItem{
				{
					ID:    "10",
					Value: false,
				},
				{
					ID:    "44",
					Value: false,
				},
			},
		},
		Child3: LargeChild{
			Items: []LargeChildItem{
				{
					ID:    "11",
					Value: false,
				},
				{
					ID:    "44",
					Value: false,
				},
			},
		},
	}
}
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
	err := Map(source, &target)
	assert.Nil(t, err)

	expected := Target{Name: "John", Age: 23}
	assert.Equal(t, expected, target)
}

func Test_mapStructIgnoreMissingFields(t *testing.T) {
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

func Test_mapStructTargetWithExtraFields(t *testing.T) {
	type Source struct {
		Name string
		Age  int
	}

	type Target struct {
		Name       string
		Age        int
		Profession string
		HasPets    bool
		Parent     *Target
	}

	source := Source{Name: "John", Age: 30}
	target := Target{}
	err := Map(source, &target)
	assert.Nil(t, err)

	expected := Target{Name: "John", Age: 30, Profession: "", HasPets: false, Parent: nil}
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

func Test_mapLargeSliceOfStructs(t *testing.T) {
	type TargetItem struct {
		ID    string
		Value bool
	}

	type TargetChild struct {
		Items []TargetItem
	}

	type Target struct {
		Name   string
		Age    int
		Value  bool
		Score  float64
		Score2 float64
		Map    map[string]string
		Child1 TargetChild
		Child2 TargetChild
		Child3 TargetChild
	}

	source := make([]Large, 0)
	for i := 0; i < 1000; i++ {
		source = append(source, buildLargeTestStruct())
	}

	target := make([]Target, 0)
	err := Map(source, &target)
	assert.Nil(t, err)

	buildTargetChild := func(s LargeChild) TargetChild {
		items := make([]TargetItem, 0)
		for _, v := range s.Items {
			items = append(items, TargetItem{
				ID:    v.ID,
				Value: v.Value,
			})
		}

		return TargetChild{
			Items: items,
		}
	}

	buildTarget := func(source Large) Target {
		target := Target{
			Name:   source.Name,
			Age:    source.Age,
			Value:  source.Value,
			Score:  source.Score,
			Score2: source.Score2,
			Map:    source.Map,
			Child1: buildTargetChild(source.Child1),
			Child2: buildTargetChild(source.Child2),
			Child3: buildTargetChild(source.Child3),
		}

		return target
	}

	expect := make([]Target, 0)
	for _, v := range source {
		expect = append(expect, buildTarget(v))
	}

	assert.Equal(t, expect, target)
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

func Test_mapStructWithTargetCustomStringType(t *testing.T) {
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

	err := MapWithConverters(source, &target, map[string]TypeConverterFn{
		"mapper.JSONStr": func(value interface{}) interface{} {
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

	// Nil target
	source := Person{Name: "John", Age: 23}
	err := Map(source, nil)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrUnexpectedNil)

	// Nil source
	target := Person{Name: "John", Age: 23}
	err = Map(nil, target)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrUnexpectedNil)
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
	assert.ErrorIs(t, err, ErrMustBePointer)

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

func Test_mapStructWithFromFieldTag(t *testing.T) {
	type Source struct {
		ID         int
		Name       string
		FamilyName string
	}

	type Target struct {
		ID        int
		FirstName string `mapper:"fromField:Name"`
		LastName  string `mapper:"fromField:FamilyName"`
	}

	source := Source{ID: 120, Name: "John", FamilyName: "Doe"}
	target := Target{}
	err := Map(source, &target)
	assert.Nil(t, err)

	expected := Target{ID: 120, FirstName: "John", LastName: "Doe"}
	assert.Equal(t, expected, target)
}

type PersonTest struct {
	ID        int
	FirstName string
	LastName  string
	Score     float64
}

func (s *PersonTest) GetFullName() string {
	return fmt.Sprintf("%v %v", s.FirstName, s.LastName)
}

func (s PersonTest) HasPassed() bool {
	return s.Score >= 70
}

func Test_mapStructWithFromMethodTag(t *testing.T) {

	type Target struct {
		ID       int
		FullName string `mapper:"fromMethod:GetFullName"`
		Passed   bool   `mapper:"fromMethod:HasPassed"`
	}

	source := PersonTest{ID: 120, FirstName: "John", LastName: "Doe", Score: 86.5}
	target := Target{}
	err := Map(source, &target)
	assert.Nil(t, err)

	expected := Target{ID: 120, FullName: "John Doe", Passed: true}
	assert.Equal(t, expected, target)
}

func Test_mapStructWithUnexportedFields(t *testing.T) {
	type Source struct {
		name string
		age  int
	}

	type Target struct {
		name string
		age  int
	}

	source := Source{name: "John", age: 23}
	target := Target{}
	err := Map(source, &target)
	assert.Nil(t, err)

	// We effectively ignore unexported source fields without crashing or raising an error value
	expected := Target{name: "", age: 0}
	assert.Equal(t, expected, target)
}
