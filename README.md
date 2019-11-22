# BloomFilter
A bloom filter using mem hash!
## Usage
### Basic
```go
package main

import (
    "github.com/quanee/BloomFilter"
)

func main() {
    // create a bloom filter
    bf := bloomfilter.New()
    // new element
    ele := []byte("bloom filter")
    // add element to bloom filter
    bf.Add(ele)
    // Determine if the element exists
    bf.Check(ele) // true
}
```

### Struct
```go
type Person struct {
    Name string `json:"name"`
    Age  int    `json:"age"`
}

func main() {
    // create a bloom filter
    bf := bloomfilter.New()
    // new person
    person := Person{
    		  Name: "bloom",
    		  Age:  100,
    	  }
    // serialize
    ele, _ := json.Marshal(person)
    // add element to bloom filter
    bf.Add(ele)
    // judge element
    bf.Check(ele)
}
```
