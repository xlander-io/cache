package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/xlander-io/cache"
)

type Person struct {
	Name     string
	Age      int32
	Location string
}

// only support reference type with CacheBytes interface support
func (p *Person) CacheBytes() int {
	return 100
}

func main() {

	// modify duplication of the default config is convenience
	config := cache.DupDefaultConfig()
	// config.CacheBytesLimit = 1024 * 1024 * 50 * 4

	local_cache, _ := cache.New(&config)

	jack := &Person{Name: "jack", Age: 12, Location: "x"}
	//p2 := &Person{Name: "rose", Age: 12, Location: "xxasdfasdfadfxxasdfasdfadf"}

	for i := 0; i < 500000; i++ {
		local_cache.Set(strconv.Itoa(i), jack, 30)
	}

	for i := 50000; i < 100000; i++ {
		local_cache.Set(strconv.Itoa(i), jack, 30)
	}

	fmt.Println(local_cache.TotalBytes())

	item, _ := local_cache.Get("jack")
	fmt.Println(item)
	//
	value0, _ := local_cache.Get("0")
	//rose, _ := local_cache.Get("rose")

	fmt.Println(value0)
	//fmt.Println(rose.CacheBytes())

	fmt.Println(local_cache.TotalItems())

	//system.Sleep(15 * time.Second)
	time.Sleep(15 * time.Second)
	fmt.Println(local_cache.TotalItems())
	//system.Sleep(15 * time.Second)
	time.Sleep(15 * time.Second)
	fmt.Println(local_cache.TotalItems())

	// get
	log.Println("---get---")
	log.Println(local_cache.Get("slice"))
	log.Println(local_cache.Get("struct*"))
	log.Println(local_cache.Get("map"))

	// overwrite
	log.Println("---set overwrite---")
	log.Println(local_cache.Get("struct*"))
	err := local_cache.Set("struct*", &Person{"Tom", 38, "London"}, 10)
	if err != nil {
		log.Fatalln("reference set error:", err)
	}
	log.Println(local_cache.Get("struct*"))

	for i := 0; i < 10000; i++ {
		a := i
		go func() {
			for {
				err := local_cache.Set(strconv.Itoa(a), &Person{"Tom", 38, "London"}, 10)
				if err != nil {
					log.Println("err: ", err)
				}
				err = local_cache.Set(strconv.Itoa(a)+"b", &Person{"Tom777", 38, "London777"}, 10)
				if err != nil {
					log.Println("err: ", err)
				}
			}
		}()
	}

	for i := 0; i < 10000; i++ {
		a := i
		go func() {
			for {
				local_cache.Get(strconv.Itoa(a))
				//log.Println("value:", v, "ttl:", ttl)
				local_cache.Get(strconv.Itoa(a) + "b")
				//log.Println("value:", v, "ttl:", ttl)
				local_cache.Delete(strconv.Itoa(a))
				local_cache.Delete(strconv.Itoa(a) + "b")
			}
		}()
	}

	for {
		log.Println("running")
		time.Sleep(5 * time.Second)
	}

	// test ttl
	// go func() {
	// 	for {
	// 		time.Sleep(2 * time.Second)
	// 		log.Println(local_cache.Get("struct*"))
	// 	}
	// }()

	//time.Sleep(20 * time.Second)

	// if not a pointer cause error
	// err = local_cache.Set("int", 10, 10)
	// if err != nil {
	// 	log.Fatalln("reference set error:", err)
	// }
}
