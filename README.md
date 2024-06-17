# reference

```high-speed```
```thread-safe```
```key-value```
```all data in memory```
```not-persistent```
```auto recycling ```

## Description
```
reference is a reference system , it is not a cache system
so the value can only be reference type 
deep copy won't happen in set process
```

## support Type
value type can only be Pointer/Slice/Map

## usage

```go
//import
import (
    "github.com/coreservice-io/reference"
)
```

### example

```go
package main

import (
	"log"
	"time"

	"github.com/coreservice-io/reference"
)

type Person struct {
	Name     string
	Age      int32
	Location string
}

func main() {

	lf := reference.New()
	lf.SetMaxRecords(10000)

	//set ""
	v := "nothing value"
	err := lf.Set("", &v, 300) //only support Pointer Slice and Map
	if err != nil {
		log.Fatalln("reference set error:", err)
	}
	//get ""
	valuen, ttl := lf.Get("")
	if valuen != nil {
		log.Println("key:nothing value:", valuen.(*string), "ttl:", ttl)
	}

	//set slice
	err = lf.Set("slice", []int{1, 2, 3}, 300)
	if err != nil {
		log.Fatalln("reference set error:", err)
	}

	//set struct pointer
	err = lf.Set("struct*", &Person{"Jack", 18, "London"}, 300)
	if err != nil {
		log.Fatalln("reference set error:", err)
	}

	//set map
	err = lf.Set("map", map[string]int{"a": 1, "b": 2}, 100)
	if err != nil {
		log.Fatalln("reference set error:", err)
	}

	//get
	log.Println("---get---")
	log.Println(lf.Get("slice"))
	log.Println(lf.Get("struct*"))
	log.Println(lf.Get("map"))

	//overwrite
	log.Println("---set overwrite---")
	log.Println(lf.Get("struct*"))
	err = lf.Set("struct*", &Person{"Tom", 38, "London"}, 10)
	if err != nil {
		log.Fatalln("reference set error:", err)
	}
	log.Println(lf.Get("struct*"))

	//test ttl
	go func() {
		for {
			time.Sleep(2 * time.Second)
			log.Println(lf.Get("struct*"))
		}
	}()

	time.Sleep(20 * time.Second)

	//if not a pointer cause error
	err = lf.Set("int", 10, 10)
	if err != nil {
		log.Fatalln("reference set error:", err)
	}
}

```

### default config

```
MaxRecords(*)         = 5000000
MinRecords            = 10000
MaxTTLSecs            = 7200
RecycleIntervalSecs   = 5
RecycleOverLimitRatio = 0.15
(* : configurable)
```

### auto recycling

RecycleOverLimitRatio of records will be recycled automatically
if MaxRecords is reached.

### custom config

```go
//new instance
lf,err := reference.New()
if err != nil {
    panic(err.Error())
}
lf.SetMaxRecords(10000) //custom the max key-value pairs that can be kept in memory
```

## Benchmark

### set

```
cpu: Intel(R) Core(TM) i7-7700HQ CPU @ 2.80GHz
BenchmarkLocalReference_SetPointer
BenchmarkLocalReference_SetPointer-8   	 1000000	      1495 ns/op	     347 B/op	      9 allocs/op
PASS
```

### get

```
cpu: Intel(R) Core(TM) i7-7700HQ CPU @ 2.80GHz
BenchmarkLocalReference_GetPointer
BenchmarkLocalReference_GetPointer-8   	 9931429	       28.48 ns/op	       0 B/op	       0 allocs/op
PASS
```
