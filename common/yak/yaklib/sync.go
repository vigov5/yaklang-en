package yaklib

import (
	"sync"

	"github.com/yaklang/yaklang/common/utils"
)

type WaitGroupProxy struct {
	*sync.WaitGroup
}

func (w *WaitGroupProxy) Add(delta ...int) {
	n := 1
	if len(delta) > 0 {
		n = delta[0]
	}
	w.WaitGroup.Add(n)
}

// NewWaitGroup Creates a WaitGroup structure reference, which helps us wait for all tasks to complete before proceeding to the next step when processing multiple concurrent tasks.
// Example:
// ```
// wg = sync.NewWaitGroup()
// for i in 5 {
// wg.Add() // Add a task
// go func(i) {
// defer wg.Done()
// time.Sleep(i)
// printf("Task %d completed\n", i)
// }(i)
// }
// wg.Wait()
// println("All tasks completed")
// ```
func NewWaitGroup() *WaitGroupProxy {
	return &WaitGroupProxy{&sync.WaitGroup{}}
}

// NewSizedWaitGroup creation A SizedWaitGroup structure reference, which helps us wait for all tasks to complete before proceeding to the next step when processing multiple concurrent tasks.
// SizedWaitGroup and WaitGroup The difference is that SizedWaitGroup can limit the number of concurrent tasks.
// Example:
// ```
// wg = sync.NewSizedWaitGroup(5) // The limit size is 5
// for i in 10 {
// wg.Add() // When the number of tasks exceeds 5, it will block until a task is completed
// go func(i) {
// defer wg.Done()
// time.Sleep(i)
// printf("Task %d completed\n", i)
// }(i)
// }
// wg.Wait()
// println("All tasks completed")
// ```
func NewSizedWaitGroup(size int) *utils.SizedWaitGroup {
	swg := utils.NewSizedWaitGroup(size)
	return swg
}

// NewMutex creates a Mutex structure reference to implement a mutex lock, which helps us avoid data competition problems when multiple concurrent tasks access the same resource.
// Example:
// ```
// m = sync.NewMutex()
// newMap = make(map[string]string)
// go func{
// for {
// m.Lock()         // Request lock
// defer m.Unlock() // Release the lock
// newMap["key"] = "value" // Prevents multiple concurrent tasks from modifying newMap at the same time
// }
// }
// for {
// println(newMap["key"])
// }
// ```
func NewMutex() *sync.Mutex {
	return new(sync.Mutex)
}

// NewRWMutex Create a RWMutex structure reference to implement a read-write lock , which helps us avoid data competition problems when multiple concurrent tasks access the same resource.
// Example:
// ```
// m = sync.NewRWMutex()
// newMap = make(map[string]string)
// go func{
// for {
// m.Lock()         // request Write lock
// defer m.Unlock() // Release the write lock
// newMap["key"] = "value" // Prevents multiple concurrent tasks from modifying newMap at the same time
// }
// }
// for {
// m.RLock()         // Request a read lock
// defer m.RUnlock() // Release the read lock
// println(newMap["key"])
// }
// ```
func NewRWMutex() *sync.RWMutex {
	return new(sync.RWMutex)
}

// NewLock Create a Mutex structure reference to implement a mutex lock, which helps us avoid data competition problems when multiple concurrent tasks access the same resource
// It is actually an alias of NewMutex
// Example:
// ```
// m = sync.NewMutex()
// newMap = make(map[string]string)
// go func{
// for {
// m.Lock()         // Request lock
// defer m.Unlock() // Release the lock
// newMap["key"] = "value" // Prevents multiple concurrent tasks from modifying newMap at the same time
// }
// }
// for {
// println(newMap["key"])
// }
// ```
func NewLock() *sync.Mutex {
	return new(sync.Mutex)
}

// NewMap Create a Map structure reference. This Map is concurrency safe
// Example:
// ```
// m = sync.NewMap()
// go func {
// for {
// m.Store("key", "value2")
// }
// }
// for {
// m.Store("key", "value")
// v, ok = m.Load("key")
// if ok {
// println(v)
// }
// }
// ```
func NewMap() *sync.Map {
	return new(sync.Map)
}

// NewOnce Creates a Once structure reference, which helps us ensure that a function will only be executed once
// Example:
// ```
// o = sync.NewOnce()
// for i in 10 {
// o.Do(func() { println("this message will only print once") })
// }
// ```
func NewOnce() *sync.Once {
	return new(sync.Once)
}

// NewPool creates a Pool structure reference, which helps us reuse temporary objects and reduce Number of memory allocations
// Example:
// ```
// p = sync.NewPool(func() {
// return make(map[string]string)
// })
// m = p.Get() // and obtains it from the Pool. If it is not in the Pool, then The first parameter function passed in will be called and a new map[string]string
// m["1"] = "2"
// println(m) // {"1": "2"}
// // Put m back into the Pool
// p.Put(m)
// m2 = p.Get() // is obtained from the Pool. In fact, what we get is the m
// println(m2) // {"1": "2"}
// ```
func NewPool(newFunc ...func() any) *sync.Pool {
	if len(newFunc) > 0 {
		return &sync.Pool{
			New: newFunc[0],
		}
	}
	return new(sync.Pool)
}

// NewCond Create a Cond structure reference, that is, a condition variable. Refer to golang official documentation: https://golang.org/pkg/sync/#Cond
// conditions Variables are a synchronization mechanism used to coordinate multiple concurrent tasks. It allows a task to wait for a certain condition to be true, and allows other tasks to notify the waiting task when the condition is true.
// Example:
// ```
// c = sync.NewCond()
// done = false
// func read(name) {
// c.L.Lock()
// for !done {
// c.Wait()
// }
// println(name, "start reading")
// c.L.Unlock()
// }
//
// func write(name) {
// time.sleep(1)
// println(name, "start writing")
// c.L.Lock()
// done = true
// c.L.Unlock()
// println(name, "wakes all")
// c.Broadcast()
// }
//
// go read("reader1")
// go read("reader2")
// go read("reader3")
// write("writer")
// time.sleep(3)
// ```
func NewCond() *sync.Cond {
	return sync.NewCond(new(sync.Mutex))
}

var SyncExport = map[string]interface{}{
	"NewWaitGroup":      NewWaitGroup,
	"NewSizedWaitGroup": NewSizedWaitGroup,
	"NewMutex":          NewMutex,
	"NewLock":           NewLock,
	"NewMap":            NewMap,
	"NewOnce":           NewOnce,
	"NewRWMutex":        NewRWMutex,
	"NewPool":           NewPool,
	"NewCond":           NewCond,
}
