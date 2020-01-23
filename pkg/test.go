package main

import (
	"fmt"
	"github.com/ds-kit/pkg/zookeeper"
	dsSync "github.com/ds-kit/pkg/zookeeper/sync"
	"github.com/ds-kit/pkg/zookeeper/sync/driver"
	"github.com/samuel/go-zookeeper/zk"
	"time"
)

func main() {

	conn, _, err := zk.Connect([]string{"192.168.205.10:2181"}, time.Second * 1000)
	if err != nil {
		panic(err)
	}

	//var count = 0
	//var wg sync.WaitGroup

	mu, err := dsSync.NewDSMutex(zookeeper.NewFramework(conn, zk.WorldACL(zk.PermAll)), new(driver.Standard), "/lock")
	if err != nil {
		panic(err)
	}

	if err = mu.Acquire(); err != nil {
		panic(err)
	}

	if err = mu.Acquire(); err != nil {
		panic(err)
	}

	if err = mu.Acquire(); err != nil {
		panic(err)
	}

	if err = mu.Acquire(); err != nil {
		panic(err)
	}

	fmt.Println("error....")

	if err := mu.Release(); err != nil {
		panic(err)
	}

	if err := mu.Release(); err != nil {
		panic(err)
	}

	if err := mu.Release(); err != nil {
		panic(err)
	}

	if err := mu.Release(); err != nil {
		panic(err)
	}

	if err = mu.Acquire(); err != nil {
		panic(err)
	}

	//for i := 0; i < 10; i++ {
	//	wg.Add(1)
	//	go func() {
	//
	//		mu := dsSync.NewDSMutex(zookeeper.NewFramework(conn, zk.WorldACL(zk.PermAll)), new(driver.standard), "/lock")
	//		err := mu.Acquire()
	//		if err != nil {
	//			panic(err)
	//		}
	//
	//		err = mu.Acquire()
	//		if err != nil {
	//			panic(err)
	//		}
	//		count+=10
	//		if err := mu.Release(); err != nil {
	//			panic(err)
	//		}
	//		wg.Done()
	//	}()
	//}
	//
	//wg.Wait()
	//fmt.Println(count)
}

