package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/xlander-io/cache"
)

type Person struct {
	Name     string
	Age      int32
	Location string
}

func (p *Person) CacheBytes() int {
	return 100
}

func main() {

	local_cache, _ := cache.New(nil)

	p1 := &Person{Name: "jack", Age: 12, Location: "x"}
	//p2 := &Person{Name: "rose", Age: 12, Location: "xxasdfasdfadfxxasdfasdfadf"}

	for i := 0; i < 500000; i++ {
		local_cache.Set(strconv.Itoa(i), p1, 30)
	}

	// for i := 50000; i < 100000; i++ {
	// 	local_cache.Set(strconv.Itoa(i), p1, 30)
	// }

	fmt.Println(local_cache.TotalBytes())

	item, _ := local_cache.Get("jack")
	fmt.Println(item)
	//
	jack, _ := local_cache.Get("0")
	//rose, _ := local_cache.Get("rose")

	fmt.Println(jack)
	//fmt.Println(rose.CacheBytes())

	fmt.Println(local_cache.TotalItems())

	//system.Sleep(15 * time.Second)
	time.Sleep(15 * time.Second)
	fmt.Println(local_cache.TotalItems())
	//system.Sleep(15 * time.Second)
	time.Sleep(15 * time.Second)
	fmt.Println(local_cache.TotalItems())

	// //set ""
	// v := "nothing value"
	// err := lf.Set("", &v, 300) //only support Pointer Slice and Map
	// if err != nil {
	// 	log.Fatalln("reference set error:", err)
	// }
	// //get ""
	// valuen, ttl := lf.Get("")
	// if valuen != nil {
	// 	log.Println("key:nothing value:", valuen.(*string), "ttl:", ttl)
	// }

	// //set slice
	// err = lf.Set("slice", []int{1, 2, 3}, 300)
	// if err != nil {
	// 	log.Fatalln("reference set error:", err)
	// }

	// //set struct pointer
	// err = lf.Set("struct*", &Person{"Jack", 18, "London"}, 300)
	// if err != nil {
	// 	log.Fatalln("reference set error:", err)
	// }

	// //set map
	// err = lf.Set("map", map[string]int{"a": 1, "b": 2}, 100)
	// if err != nil {
	// 	log.Fatalln("reference set error:", err)
	// }

	// //get
	// log.Println("---get---")
	// log.Println(lf.Get("slice"))
	// log.Println(lf.Get("struct*"))
	// log.Println(lf.Get("map"))

	// //overwrite
	// log.Println("---set overwrite---")
	// log.Println(lf.Get("struct*"))
	// err = lf.Set("struct*", &Person{"Tom", 38, "London"}, 10)
	// if err != nil {
	// 	log.Fatalln("reference set error:", err)
	// }
	// log.Println(lf.Get("struct*"))

	// for i := 0; i < 10000; i++ {
	// 	a := i
	// 	go func() {
	// 		for {
	// 			err := lf.Set(strconv.Itoa(a), &Person{"Tom", 38, "London"}, 10)
	// 			if err != nil {
	// 				log.Println("err:", err)
	// 			}
	// 			err = lf.Set(strconv.Itoa(a)+"b", &Person{"Tom777", 38, "London777"}, 10)
	// 			if err != nil {
	// 				log.Println("err:", err)
	// 			}
	// 		}
	// 	}()
	// }

	// for i := 0; i < 10000; i++ {
	// 	a := i
	// 	go func() {
	// 		for {
	// 			lf.Get(strconv.Itoa(a))
	// 			//log.Println("value:", v, "ttl:", ttl)
	// 			lf.Get(strconv.Itoa(a) + "b")
	// 			//log.Println("value:", v, "ttl:", ttl)
	// 			lf.Delete(strconv.Itoa(a))
	// 			lf.Delete(strconv.Itoa(a) + "b")
	// 		}
	// 	}()
	// }

	// for {
	// 	log.Println("running")
	// 	time.Sleep(5 * time.Second)
	// }

	// //test ttl
	// go func() {
	// 	for {
	// 		time.Sleep(2 * time.Second)
	// 		log.Println(lf.Get("struct*"))
	// 	}
	// }()

	// time.Sleep(20 * time.Second)

	// //if not a pointer cause error
	// err = lf.Set("int", 10, 10)
	// if err != nil {
	// 	log.Fatalln("reference set error:", err)
	// }
}
