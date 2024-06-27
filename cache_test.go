package cache

import (
	"log"
	"math/rand"
	"runtime"
	"strconv"
	"sync"
	"testing"
	"time"
	"unsafe"
)

type Person struct {
	Name     string
	Age      int
	Location string
}

func (p *Person) CacheBytes() int {
	return int(unsafe.Sizeof(*p))
}

func printMemStats() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	log.Printf("Alloc = %v KB, TotalAlloc = %v KB, Sys = %v KB,Lookups = %v NumGC = %v\n", m.Alloc/1024, m.TotalAlloc/1024, m.Sys/1024, m.Lookups, m.NumGC)
}

func Min(x, y int64) int64 {
	if x < y {
		return x
	}
	return y
}

func Max(x, y int64) int64 {
	if x > y {
		return x
	}
	return y
}

func Clamp(x, min, max int64) int64 {
	return Min(Max(x, min), max)
}

func EqualWithEps(a, b, eps int64) bool {
	c := a - b
	if c > 0 {
		return c <= +eps
	} else if c < 0 {
		return c >= -eps
	}
	return true
}

func Test_Cache_Simple(t *testing.T) {

	cache, err := New(nil)

	if nil != err {
		t.Fatalf("New cache instance failed! err=%v", err)
	}

	jack := &Person{"Jack", 18, "London"}
	cache.Set("a", jack, 5)
	cache.Set("b", jack, 5)

	if int32(2) != cache.Items() {
		t.Fatalf("cache's total items count should be 2, but %d", cache.Items())
	}

	{
		v, ttl := cache.Get("a")
		log.Println("get a")
		log.Println(v, ttl)

		if v != jack {
			t.Fatalf("get 'a' from cache should be %v, but %v", jack, v)
		}

		if int64(5) != ttl {
			t.Fatalf("the ttl of object which get from key 'a' should be 5, but %d", ttl)
		}
	}

	log.Println("delete a")
	cache.Delete("a")

	if int32(1) != cache.Items() {
		t.Fatalf("count of cache's total items should be 1, but %d", cache.Items())
	}

	{
		v, ttl := cache.Get("a")
		log.Println("get a")
		log.Println(v, ttl)

		if nil != v {
			t.Fatalf("get 'a' from cache should be nil, but %v", v)
		}

		if int64(0) != ttl {
			t.Fatalf("the ttl of object which is not exist should be 0, but %d", ttl)
		}
	}

	{
		v, ttl := cache.Get("b")
		log.Println("get b")
		log.Println(v, ttl)

		if v != jack {
			t.Fatalf("get 'a' from cache should be %v, but %v", jack, v)
		}

		if int64(5) != ttl {
			t.Fatalf("the ttl of object which get from key 'b' should be 5, but %d", ttl)
		}
	}

	log.Println("origin a")
	log.Println(*jack)

	log.Println("To simulate TTLs to timeout, waiting for 10 seconds ...")
	time.Sleep(10 * time.Second) // wait timeout for all ttls

	if int32(0) != cache.Items() {
		t.Fatalf("count of cache's total items should be 0 now, but %d", cache.Items())
	}

	{
		v, ttl := cache.Get("a")
		log.Println("get a")
		log.Println(v, ttl)

		if nil != v {
			t.Fatalf("get 'a' from cache should be nil, but %v", v)
		}

		if int64(0) != ttl {
			t.Fatalf("the ttl of object which is not exist should be 0, but %d", ttl)
		}
	}
	{
		v, ttl := cache.Get("b")
		log.Println("get b")
		log.Println(v, ttl)

		if nil != v {
			t.Fatalf("get 'a' from cache should be nil, but %v", v)
		}

		if int64(0) != ttl {
			t.Fatalf("the ttl of object which is not exist should be 0, but %d", ttl)
		}
	}
}

func Test_Cache_Expire(t *testing.T) {
	cache, err := New(nil)

	if nil != err {
		t.Fatalf("New cache instance failed! err=%v", err)
	}

	jack := &Person{"Jack", 18, "London"}

	cache.Set("1", jack, 5)
	cache.Set("2", jack, 18)
	cache.Set("3", jack, 23)
	cache.Set("4", jack, -100)
	cache.Set("5", jack, 3000000)
	cache.Set("6", jack, 35)

	if int32(5) != cache.Items() {
		t.Fatalf("count of cache's total items should be 5, but %d", cache.Items())
	}

	count := 0
	for {
		v1, ttl1 := cache.Get("1")
		log.Printf("1==>%v %v", v1, ttl1)
		v2, ttl2 := cache.Get("2")
		log.Printf("2==>%v %v", v2, ttl2)
		v3, ttl3 := cache.Get("3")
		log.Printf("3==>%v %v", v3, ttl3)
		v4, ttl4 := cache.Get("4")
		log.Printf("4==>%v %v", v4, ttl4)
		v5, ttl5 := cache.Get("5")
		log.Printf("5==>%v %v", v5, ttl5)
		v6, ttl6 := cache.Get("6")
		log.Printf("6==>%v %v", v6, ttl6)
		log.Println("total key", cache.Items())
		log.Println("-----------")

		if int(0) == count {
			if int32(5) != cache.Items() {
				t.Fatalf("count of cache's total items should be 5, but %d", cache.Items())
			}

			if v1 != jack {
				t.Fatalf("get '1' from cache should be %v, but %v", jack, v1)
			}
			if v2 != jack {
				t.Fatalf("get '2' from cache should be %v, but %v", jack, v2)
			}
			if v3 != jack {
				t.Fatalf("get '3' from cache should be %v, but %v", jack, v3)
			}
			if v4 != nil {
				t.Fatalf("get '4' from cache should be nil, but %v", v4)
			}
			if v5 != jack {
				t.Fatalf("get '5' from cache should be %v, but %v", jack, v5)
			}
			if v6 != jack {
				t.Fatalf("get '6' from cache should be %v, but %v", jack, v6)
			}

			if int64(5) != ttl1 {
				t.Fatalf("the ttl of object which get from key '1' should be 5, but %d", ttl1)
			}
			if int64(18) != ttl2 {
				t.Fatalf("the ttl of object which get from key '2' should be 18, but %d", ttl2)
			}
			if int64(23) != ttl3 {
				t.Fatalf("the ttl of object which get from key '3' should be 23, but %d", ttl3)
			}
			if int64(0) != ttl4 {
				t.Fatalf("the ttl of object which get from key '4' should be 0, but %d", ttl4)
			}
			if int64(7200) != ttl5 {
				t.Fatalf("the ttl of object which get from key '5' should be 7200, but %d", ttl5)
			}
			if int64(35) != ttl6 {
				t.Fatalf("the ttl of object which get from key '6' should be 35, but %d", ttl6)
			}
		}

		if int(10) == count {
			if int32(4) != cache.Items() {
				t.Fatalf("count of cache's total items should be 4, but %d", cache.Items())
			}

			if v1 != nil {
				t.Fatalf("get '1' from cache should be nil, but %v", v1)
			}
			if v2 != jack {
				t.Fatalf("get '2' from cache should be %v, but %v", jack, v2)
			}
			if v3 != jack {
				t.Fatalf("get '3' from cache should be %v, but %v", jack, v3)
			}
			if v4 != nil {
				t.Fatalf("get '4' from cache should be nil, but %v", v4)
			}
			if v5 != jack {
				t.Fatalf("get '5' from cache should be %v, but %v", jack, v5)
			}
			if v6 != jack {
				t.Fatalf("get '6' from cache should be %v, but %v", jack, v6)
			}

			if int64(0) != ttl1 {
				t.Fatalf("the ttl of object which get from key '1' should be 0, but %d", ttl1)
			}
			if int64(18-10) != ttl2 {
				t.Fatalf("the ttl of object which get from key '2' should be 18-10, but %d", ttl2)
			}
			if int64(23-10) != ttl3 {
				t.Fatalf("the ttl of object which get from key '3' should be 23-10, but %d", ttl3)
			}
			if int64(0) != ttl4 {
				t.Fatalf("the ttl of object which get from key '4' should be 0, but %d", ttl4)
			}
			if int64(7200-10) != ttl5 {
				t.Fatalf("the ttl of object which get from key '5' should be 7200-10, but %d", ttl5)
			}
			if int64(35-10) != ttl6 {
				t.Fatalf("the ttl of object which get from key '6' should be 35-10, but %d", ttl6)
			}
		}

		if int(20) == count {
			if int32(3) != cache.Items() {
				t.Fatalf("count of cache's total items should be 3, but %d", cache.Items())
			}

			if v1 != nil {
				t.Fatalf("get '1' from cache should be nil, but %v", v1)
			}
			if v2 != nil {
				t.Fatalf("get '2' from cache should be nil, but %v", v2)
			}
			if v3 != jack {
				t.Fatalf("get '3' from cache should be %v, but %v", jack, v3)
			}
			if v4 != nil {
				t.Fatalf("get '4' from cache should be nil, but %v", v4)
			}
			if v5 != jack {
				t.Fatalf("get '5' from cache should be %v, but %v", jack, v5)
			}
			if v6 != jack {
				t.Fatalf("get '6' from cache should be %v, but %v", jack, v6)
			}

			if int64(0) != ttl1 {
				t.Fatalf("the ttl of object which get from key '1' should be 0, but %d", ttl1)
			}
			if int64(0) != ttl2 {
				t.Fatalf("the ttl of object which get from key '2' should be 0, but %d", ttl2)
			}
			if int64(23-20) != ttl3 {
				t.Fatalf("the ttl of object which get from key '3' should be 23-20, but %d", ttl3)
			}
			if int64(0) != ttl4 {
				t.Fatalf("the ttl of object which get from key '4' should be 0, but %d", ttl4)
			}
			if int64(7200-20) != ttl5 {
				t.Fatalf("the ttl of object which get from key '5' should be 7200-20, but %d", ttl5)
			}
			if int64(35-20) != ttl6 {
				t.Fatalf("the ttl of object which get from key '6' should be 35-20, but %d", ttl6)
			}
		}

		if int(30) == count {
			if int32(2) != cache.Items() {
				t.Fatalf("count of cache's total items should be 2, but %d", cache.Items())
			}

			if v1 != nil {
				t.Fatalf("get '1' from cache should be nil, but %v", v1)
			}
			if v2 != nil {
				t.Fatalf("get '2' from cache should be nil, but %v", v2)
			}
			if v3 != nil {
				t.Fatalf("get '3' from cache should be nil, but %v", v3)
			}
			if v4 != nil {
				t.Fatalf("get '4' from cache should be nil, but %v", v4)
			}
			if v5 != jack {
				t.Fatalf("get '5' from cache should be %v, but %v", jack, v5)
			}
			if v6 != jack {
				t.Fatalf("get '6' from cache should be %v, but %v", jack, v6)
			}

			if int64(0) != ttl1 {
				t.Fatalf("the ttl of object which get from key '1' should be 0, but %d", ttl1)
			}
			if int64(0) != ttl2 {
				t.Fatalf("the ttl of object which get from key '2' should be 0, but %d", ttl2)
			}
			if int64(0) != ttl3 {
				t.Fatalf("the ttl of object which get from key '3' should be 0, but %d", ttl3)
			}
			if int64(0) != ttl4 {
				t.Fatalf("the ttl of object which get from key '4' should be 0, but %d", ttl4)
			}
			if int64(7200-30) != ttl5 {
				t.Fatalf("the ttl of object which get from key '5' should be 7200-30, but %d", ttl5)
			}
			if int64(35-30) != ttl6 {
				t.Fatalf("the ttl of object which get from key '6' should be 35-30, but %d", ttl6)
			}
		}

		if int(40) == count {
			if int32(1) != cache.Items() {
				t.Fatalf("count of cache's total items should be 2, but %d", cache.Items())
			}

			if v1 != nil {
				t.Fatalf("get '1' from cache should be nil, but %v", v1)
			}
			if v2 != nil {
				t.Fatalf("get '2' from cache should be nil, but %v", v2)
			}
			if v3 != nil {
				t.Fatalf("get '3' from cache should be nil, but %v", v3)
			}
			if v4 != nil {
				t.Fatalf("get '4' from cache should be nil, but %v", v4)
			}
			if v5 != jack {
				t.Fatalf("get '5' from cache should be %v, but %v", jack, v5)
			}
			if v6 != nil {
				t.Fatalf("get '6' from cache should be nil, but %v", v6)
			}

			if int64(0) != ttl1 {
				t.Fatalf("the ttl of object which get from key '1' should be 0, but %d", ttl1)
			}
			if int64(0) != ttl2 {
				t.Fatalf("the ttl of object which get from key '2' should be 0, but %d", ttl2)
			}
			if int64(0) != ttl3 {
				t.Fatalf("the ttl of object which get from key '3' should be 0, but %d", ttl3)
			}
			if int64(0) != ttl4 {
				t.Fatalf("the ttl of object which get from key '4' should be 0, but %d", ttl4)
			}
			if int64(7200-40) != ttl5 {
				t.Fatalf("the ttl of object which get from key '5' should be 7200-40, but %d", ttl5)
			}
			if int64(0) != ttl6 {
				t.Fatalf("the ttl of object which get from key '6' should be 0, but %d", ttl6)
			}
		}

		count++
		if count > 45 {
			return
		}
		time.Sleep(time.Second)
	}
}

func Test_Cache_SetAndRemove(t *testing.T) {
	jack := &Person{"Jack", 18, "America"}
	cache, err := New(nil)

	if nil != err {
		t.Fatalf("New cache instance failed! err=%v", err)
	}

	log.Println("start")
	printMemStats()

	for i := 0; i < 20; i++ {
		//set
		for j := 0; j < 10000; j++ {
			cache.Set(strconv.Itoa(j), jack, 1)
		}

		if int32(10000) != cache.Items() {
			t.Fatalf("count of cache's total items should be 10000, but %d", cache.Items())
		}

		if int32(10000*jack.CacheBytes()) != cache.Bytes() {
			t.Fatalf("count of cache's total items should be 10000*%d, but %d", jack.CacheBytes(), cache.Items())
		}

		log.Println("round:", i)
		log.Println("finish set")
		printMemStats()

		time.Sleep(2 * time.Second)
	}

	log.Println("finish")
	printMemStats()
}

func Test_Cache_FastRecycling(t *testing.T) {
	jack := &Person{"Jack", 18, "America"}

	config := CacheConfig{
		CacheBytesLimit: 1000 * 10000 * 40,
	}

	cache, err := New(&config)

	if nil != err {
		t.Fatalf("New cache instance failed! err=%v", err)
	}

	for j := 0; j < 1000*10000; j++ {
		cache.Set(strconv.Itoa(j), jack, int64(60))
	}

	for i := 0; i < 50; i++ {
		log.Println("total items:", cache.Items())
		log.Println("total bytes:", cache.Bytes())
		time.Sleep(1 * time.Second)
	}

}
func Test_Cache_BigAmountKey(t *testing.T) {
	jack := &Person{"Jack", 18, "America"}

	config := CacheConfig{
		CacheBytesLimit: 1024 * 1024 * 50 * 4,
	}

	cache, err := New(&config)

	if nil != err {
		t.Fatalf("New cache instance failed! err=%v", err)
	}

	log.Println("start")
	printMemStats()

	// go func() {
	// 	log.Println(http.ListenAndServe("0.0.0.0:10000", nil))
	// }()

	for i := 0; i < 30; i++ {
		log.Println("----------")
		log.Println("round", i)
		log.Println("mem start set")
		printMemStats()

		for j := 0; j < 100*10000; j++ {
			cache.Set(strconv.Itoa(j), jack, int64(rand.Intn(10)+10))
		}

		//time.Sleep(2 * time.Second) // waiting for skiplist to update ttl

		if cache.Items()*int32(jack.CacheBytes()) != cache.Bytes() {
			t.Fatalf("total bytes expect %d*%d, but %d", jack.CacheBytes(), cache.Items(), cache.Bytes())
		}

		if int32(100*10000) != cache.Items() {
			t.Fatalf("expect total items: 100*10000, but %d", cache.Items())
		}

		if 100*10000*jack.CacheBytes() != int(cache.Bytes()) {
			t.Fatalf("expect total bytes: 100*10000*%d, but %d", jack.CacheBytes(), cache.Bytes())
		}

		log.Println("mem after set")
		log.Printf("unix time now: %d\n", time.Now().Unix())
		printMemStats()
		time.Sleep(time.Second)
	}

	log.Println("~~~~~~")
	log.Println("finish set")
	printMemStats()

	log.Println("do GC")

	runtime.GC()
	log.Println("after GC")
	printMemStats()

	count := 0
	for {
		time.Sleep(1 * time.Second)
		log.Println("---job finished---")
		printMemStats()
		count++
		if count > 45 {
			return
		}
	}
	//time.Sleep(1*time.Hour)
}

func Test_Cache_RandomSet(t *testing.T) {
	cache, _ := New(nil)
	jack := &Person{"Jack", 18, "America"}

	cache.Set("a", jack, 15)
	cache.Set("b", jack, 19)
	cache.Set("c", jack, 60)
	cache.Set("d", jack, 63)
	cache.Set("e", jack, 65)

	log.Println("before big amount set")
	v, ttl := cache.Get("a")
	log.Printf("a==>%v %v", v, ttl)
	v, ttl = cache.Get("b")
	log.Printf("b==>%v %v", v, ttl)
	v, ttl = cache.Get("c")
	log.Printf("c==>%v %v", v, ttl)
	v, ttl = cache.Get("d")
	log.Printf("d==>%v %v", v, ttl)
	v, ttl = cache.Get("e")
	log.Printf("e==>%v %v", v, ttl)

	log.Println("start amount set")
	for i := 0; i < 200; i++ {
		for j := 0; j < 10000; j++ {
			num := rand.Intn(9999999999999)
			key := strconv.Itoa(num)
			cache.Set(key, jack, int64(rand.Intn(30)+20))
		}

		if cache.Items()*int32(jack.CacheBytes()) != cache.Bytes() {
			t.Fatalf("total bytes expect %d*%d, but %d", jack.CacheBytes(), cache.Items(), cache.Bytes())
		}
	}

	if cache.Items()*int32(jack.CacheBytes()) != cache.Bytes() {
		t.Fatalf("total bytes expect %d*%d, but %d", jack.CacheBytes(), cache.Items(), cache.Bytes())
	}

	for i := 0; i < 70; i++ {
		time.Sleep(time.Second)
		log.Println("--------------")
		v, ttl = cache.Get("a")
		log.Printf("a==>%v %v", v, ttl)
		v, ttl = cache.Get("b")
		log.Printf("b==>%v %v", v, ttl)
		v, ttl = cache.Get("c")
		log.Printf("c==>%v %v", v, ttl)
		v, ttl = cache.Get("d")
		log.Printf("d==>%v %v", v, ttl)
		v, ttl = cache.Get("e")
		log.Printf("e==>%v %v", v, ttl)
		log.Println("total key", cache.Items())

		if cache.Items()*int32(jack.CacheBytes()) != cache.Bytes() {
			t.Fatalf("total bytes expect %d*%d, but %d", jack.CacheBytes(), cache.Items(), cache.Bytes())
		}
	}
}

func Test_Cache_KeepTTL(t *testing.T) {
	cache, _ := New(nil)
	mayun := &Person{"Ma Yun", 58, "China"}
	jack := &Person{"Jack Ma", 18, "America"}

	cache.Set("a", mayun, 30)
	cache.Set("b", mayun, 40)
	cache.Set("c", mayun, 50)

	//log
	{
		v, ttl := cache.Get("a")
		log.Printf("a==>%v %v", v, ttl)

		if v != mayun {
			t.Fatalf("item of key 'a' expect %v, but %v", mayun, v)
		}
		if ttl != 30 {
			t.Fatalf("ttl of key 'a' expect %d, but %d", 30, ttl)
		}
	}

	{
		v, ttl := cache.Get("b")
		log.Printf("b==>%v %v", v, ttl)

		if v != mayun {
			t.Fatalf("item of key 'b' expect %v, but %v", mayun, v)
		}
		if ttl != 40 {
			t.Fatalf("ttl of key 'b' expect %d, but %d", 40, ttl)
		}
	}

	{
		v, ttl := cache.Get("c")
		log.Printf("c==>%v %v", v, ttl)

		if v != mayun {
			t.Fatalf("item of key 'c' expect %v, but %v", mayun, v)
		}
		if ttl != 50 {
			t.Fatalf("ttl of key 'c' expect %d, but %d", 50, ttl)
		}
	}

	time.Sleep(5 * time.Second)

	cache.Set("a", jack, 300) // update ttl to 300 for key 'a'
	cache.Set("b", jack, 0)   // do nothing

	{
		v, ttl := cache.Get("a")
		log.Printf("a==>%v %v", v, ttl)

		if v != jack {
			t.Fatalf("item of key 'a' expect %v, but %v", jack, v)
		}
		if ttl != 300 {
			t.Fatalf("ttl of key 'a' expect %d, but %d", 300, ttl)
		}
	}

	{
		v, ttl := cache.Get("b")
		log.Printf("b==>%v %v", v, ttl)

		if v != jack {
			t.Fatalf("item of key 'b' expect %v, but %v", jack, v)
		}
		if !EqualWithEps(40-5, ttl, 1) {
			t.Fatalf("ttl of key 'b' expect %d-5, but %d", 40, ttl)
		}
	}

	{
		v, ttl := cache.Get("c")
		log.Printf("c==>%v %v", v, ttl)

		if v != mayun {
			t.Fatalf("item of key 'c' expect %v, but %v", mayun, v)
		}
		if !EqualWithEps(50-5, ttl, 1) {
			t.Fatalf("ttl of key 'c' expect %d-5, but %d", 50, ttl)
		}
	}

	//log
	for i := 0; i < 10; i++ {
		log.Println("-----------")
		v, ttl := cache.Get("a")
		log.Printf("a==>%v %v", v, ttl)
		v, ttl = cache.Get("b")
		log.Printf("b==>%v %v", v, ttl)
		v, ttl = cache.Get("c")
		log.Printf("c==>%v %v", v, ttl)
		time.Sleep(time.Second)
	}

}

func Test_Cache_SetTTL(t *testing.T) {
	cache, _ := New(nil)
	mayun := &Person{"Ma Yun", 58, "China"}

	TTLs := []int64{1, 20000, 0, -100, 200, 45, 346547457457457, -20000, 434, 9}
	for i := 0; i < 10; i++ {
		key := strconv.Itoa(i)
		cache.Set(key, mayun, TTLs[i])
	}

	for i := 0; i < 10; i++ {
		log.Println("-----------")
		for j := 0; j < 10; j++ {
			key := strconv.Itoa(j)
			v, ttl := cache.Get(key)
			log.Printf("%s==>%v %v", key, v, ttl)

			// if v != mayun {
			// 	t.Fatalf("expect %v, but %v", mayun, v)
			// }

			TTL := Max(Clamp(TTLs[j], 0, 7200)-int64(i), 0)
			if !EqualWithEps(TTL, ttl, 1) {
				t.Fatalf("input %d, expect %d, but %v", TTLs[j], TTL, ttl)
			}
		}
		log.Println("total key", cache.Items())
		time.Sleep(time.Second)
	}
}

// func Test_SyncMap(t *testing.T) {
// 	printMemStats()

// 	go func() {
// 		log.Println(http.ListenAndServe("0.0.0.0:10000", nil))
// 	}()

// 	type Person struct {
// 		Name     string
// 		Age      int
// 		Location string
// 	}
// 	a := Person{"Jack", 18, "America"}
// 	type Element struct {
// 		Member string
// 		Score  int64
// 		Value  interface{}
// 	}

// 	var myMap sync.Map
// 	for i := 0; i < 1000000; i++ {
// 		key := strconv.Itoa(i)
// 		b := &Element{
// 			Member: key,
// 			Score:  10,
// 			Value:  a,
// 		}
// 		myMap.Store(key, b)
// 	}

// 	printMemStats()

// 	for i := 0; i < 1000000; i++ {
// 		myMap.Delete(strconv.Itoa(i))
// 	}
// 	runtime.GC()

// 	for {
// 		printMemStats()
// 		time.Sleep(time.Second)
// 	}
// }

func BenchmarkLocalReference_SetPointer(b *testing.B) {
	cache, _ := New(nil)
	jack := &Person{"Jack", 18, "America"}

	keyArray := []string{}
	for i := 0; i < b.N; i++ {
		keyArray = append(keyArray, strconv.Itoa(i))
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set(keyArray[i], jack, 300)
	}
}

func BenchmarkLocalReference_GetPointer(b *testing.B) {
	cache, _ := New(nil)
	jack := &Person{"Jack", 18, "America"}
	cache.Set("1", jack, 300)
	var e *Person
	log.Println(e)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		it, _ := cache.Get("1")
		e = it.(*Person)
	}
}

func Benchmark_syncMap(b *testing.B) {
	var m sync.Map
	jack1 := &Person{"Jack", 18, "America"}
	for i := 0; i < 100; i++ {
		m.Store(i, jack1)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p, _ := m.Load(1)
		jack2 := &Person{"Jack", 18, "America"}
		m.Store(i, jack2)
		_ = p.(*Person)
	}

}

func Benchmark_map(b *testing.B) {
	m := map[int]int{}
	for i := 0; i < 100; i++ {
		m[i] = i
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m[i]
	}

}

func Benchmark_time(b *testing.B) {

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		time.Now().Unix()
	}

}
