# Object mapper

A tiny library to perform value mappings from a source to a target using reflection.

## Install
```bash
$ go get github.com/agustinaliagac/mapper
```

## How it works
This library copies values from source to target objects by following these rules:
- If a `mapper` struct field tag is present, look for `fromField` or `fromMethod` options.
- Else, use the same field name declared in the target struct.
- All other fields from source that don't exist in the target are ignored.
- All fields in target that have no counterpart in source are left with their zero-value.
- All unexported fields are silently ignored.

If you need anything beyond that, you will probably have to write your own mapping logic, with or without this library's help.

## Usage

```go
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
if err != nil {
    // handle error
}

fmt.Println(target) // {John 30}
```

## Use cases

The most typical use case for this library is to project data from one struct (or slice of structs) into a smaller subset of fields, i.e. to project some values from "source" while ignoring other fields.

One very common example of this is having an Entity/Model with data from some data source, and wanting to project that into a DTO (Data Transfer Object). A DTO allows you to decouple entities (that belong in the Domain Layer) from serialization and other mechanisms which belong to the Application Layer.

### More on DTOs

As an example, Go is very flexible and will let you use the same struct for both database models and JSON serialization if you include the appropiate field tags (in this example, using GORM):
```go

type User struct {
	ID                 int            `json:"id" gorm:"column:user_id"`
	CreationDate       *time.Time     `json:"creationDate" gorm:"autoCreateTime"`
	UpdateDate         *time.Time     `json:"-" gorm:"autoUpdateTime"`
	Name               string         `json:"name"`
	Username           string         `json:"username"`
	Email              *string        `json:"email"`
	Password           *string        `json:"-"`
	Deleted            *time.Time     `json:"-"`
	CreatedBy          *int           `json:"createdBy"`
	Role               string         `json:"role" gorm:"column:identifying_role"`
	Institutions       *[]Institution `json:"institutions" gorm:"many2many:Users_Institutions;"`
}
```
Coupling domain and application/interface layers like this might be a nightmare to maintain in the future.
While this is perfectly valid Go code and you can in fact [ignore some fields with the "-" tag](https://pkg.go.dev/encoding/json#Marshal), I would personally break such implementation in separate structs, each with its own purpose:

```go

type User struct {
	ID                 int            `gorm:"column:user_id"`
	CreationDate       *time.Time     `gorm:"autoCreateTime"`
	UpdateDate         *time.Time     `gorm:"autoUpdateTime"`
	Name               string
	Username           string
	Email              *string
	Password           *string
	Deleted            *time.Time
	CreatedBy          *int
	Role               string         `gorm:"column:identifying_role"`
	Institutions       *[]Institution `gorm:"many2many:Users_Institutions;"`
}

type UserDTO struct {
	ID                 int            `json:"id"`
	CreationDate       *time.Time     `json:"creationDate"`
	Name               string         `json:"name"`
	Username           string         `json:"username"`
	Email              *string        `json:"email"`
	CreatedBy          *int           `json:"createdBy"`
	Role               string         `json:"role"`
	Institutions       *[]Institution `json:"institutions"`
}
```

## What are the benefits of using this library?
- Reduce your project's boilerplate, repetitive code (which can be quite huge for large structs and which you'll have to maintain), which in return will:
- Let you focus on more important things that actually add some value to your project as a whole

## What are the risks/drawbacks of using this library?
### Performance
One of the top disadvantages of using reflection is performance. I've written some not-so-rigurous benchmark tests which you can inspect, just to have a vague idea of how big the performance hit could be when using reflection instead of manually writing your own transformation functions.
```
goos: darwin
goarch: amd64
pkg: github.com/agustinaliagac/mapper
cpu: Intel(R) Core(TM) i7-7660U CPU @ 2.50GHz
BenchmarkMapping/MapSmallStructReflect-4                1000000000               0.0000086 ns/op
BenchmarkMapping/MapSmallStructManual-4                 1000000000               0.0000009 ns/op
BenchmarkMapping/MapSmallSliceOfStructsReflect-4        1000000000               0.0000163 ns/op
BenchmarkMapping/MapSmallSliceOfStructsManual-4         1000000000               0.0000020 ns/op
BenchmarkMapping/MapLargeStructReflect-4                1000000000               0.0000302 ns/op
BenchmarkMapping/MapLargeStructManual-4                 1000000000               0.0000030 ns/op
BenchmarkMapping/MapLargeSliceOfStructsReflect-4        1000000000               0.009009 ns/op
BenchmarkMapping/MapLargeSliceOfStructsManual-4         1000000000               0.0006908 ns/op
```

As you can see, code that uses reflection can be roughly 8-15 times slower, but don't take this as a definitive statement. You can always try it out for yourself and measure how big the impact is in your codebase.

Depending on your requirements and how big the objects you're mapping are, the overall performance hit may or may not outweight the productivity gains of using this library.

### Compile time errors

One of Go's most appealing properties to me is its static-typing. When you're writing your own transformation functions, you get compile-time errors when doing something wrong. However, *note that having compile-time errors will not protect you from any type of issues: e.g: if you make a mistake by forgetting to set one field to the target struct*.

When you're using this library, you're letting type-conversion be a run-time operation, and as such, you should now be prepared to handle errors at run-time.
You can do this just like you handle any other error in the Go language:

```go
err := Map(source, &target)
if err != nil {
    // Do something
}
```
