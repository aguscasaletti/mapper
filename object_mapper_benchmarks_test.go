package mapper

import (
	"fmt"
	"testing"
)

type testCase struct {
	Name        string
	ManualImpl  func(b *testing.B) interface{}
	ReflectImpl func(b *testing.B) interface{}
}

var testCases = []testCase{
	{
		"MapSmallStruct",
		mapSmallStructManual,
		mapSmallStructReflect,
	},
	{
		"MapSmallSliceOfStructs",
		mapSmallSliceOfStructsManual,
		mapSmallSliceOfStructsReflect,
	},
	{
		"MapLargeStruct",
		mapLargeStructManual,
		mapLargeStructReflect,
	},
	{
		"MapLargeSliceOfStructs",
		mapLargeSliceOfStructsManual,
		mapLargeSliceOfStructsReflect,
	},
}

func mapSmallStructReflect(b *testing.B) interface{} {
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
	if err != nil {
		b.Fail()
	}

	return target
}

func mapSmallStructManual(b *testing.B) interface{} {
	type Source struct {
		Name string
		Age  int
	}

	type Target struct {
		Name string
		Age  int
	}

	source := Source{Name: "John", Age: 23}
	target := Target{
		Name: source.Name,
		Age:  source.Age,
	}

	return target
}

func mapSmallSliceOfStructsReflect(b *testing.B) interface{} {
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
	if err != nil {
		b.Fail()
	}

	return regions
}

func mapSmallSliceOfStructsManual(b *testing.B) interface{} {
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

	regions := make([]Regions, 0)
	for _, v := range countries {
		regions = append(regions, Regions{
			Name: v.Name,
		})
	}
	return regions
}

type LargeChildItem struct {
	ID    string
	Value bool
}

type LargeChild struct {
	Items []LargeChildItem
}

type Large struct {
	Name   string
	Age    int
	Value  bool
	Score  float64
	Score2 float64
	Map    map[string]string
	Child1 LargeChild
	Child2 LargeChild
	Child3 LargeChild
}

func mapLargeStructReflect(b *testing.B) interface{} {
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

	source := buildLargeTestStruct()

	target := Target{}
	err := Map(source, &target)
	if err != nil {
		b.Fail()
	}

	return target
}

func mapLargeStructManual(b *testing.B) interface{} {
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

	source := buildLargeTestStruct()

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

func mapLargeSliceOfStructsReflect(b *testing.B) interface{} {
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
	if err != nil {
		b.Fail()
	}

	return target
}

func mapLargeSliceOfStructsManual(b *testing.B) interface{} {
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

	target := make([]Target, 0)
	for _, v := range source {
		target = append(target, buildTarget(v))
	}
	return target
}

func BenchmarkMapping(b *testing.B) {
	for _, item := range testCases {
		b.Run(fmt.Sprintf("%vReflect", item.Name), func(b *testing.B) {
			item.ReflectImpl(b)
		})
		b.Run(fmt.Sprintf("%vManual", item.Name), func(b *testing.B) {
			item.ManualImpl(b)
		})
	}
}
