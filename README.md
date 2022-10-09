# Allocate but evil

Allocate provides functions for allocating golang structs so that pointer fields are pointers to zero'd values instead of `nil`.

Also provides two evil functions `ZeroNested`/`SetNested` for only allocating or setting value for nested field.

## Brief Example

```go
package main

import (
    "fmt"

    "github.com/Thearas/allocate"
)

type TopLevel struct {
    Inner *Embedded
}

type Embedded struct {
    Inner *Leaf
}

type Leaf struct {
    SomeInt int
}

func main() {
    top := new(TopLevel)
    fmt.Printf("befor: %v\n", top)

    allocate.MustSetNested(&top, ".Inner.Inner", &Leaf{SomeInt: 1})
    fmt.Printf("after: top.Inner.Inner.SomeInt == %d\n", top.Inner.Inner.SomeInt)
}
```

```bash
# OUTPUT
before: &{<nil>}
after: top.Inner.Inner.SomeInt == 1
```

### Use Cases

* Initializing structures that contain any type of pointer fields, including recursive struct fields
* Preventing panics by ensuring that all fields of a struct are initialized
* Initializing [golang protobuf struct](https://github.com/golang/protobuf) (the golang protobuf makes heavy use of pointers to embedded structs that contain pointers to embedded structs, ad infinitum)
* Initializing structs for black box testing (see also https://golang.org/pkg/testing/quick/)
