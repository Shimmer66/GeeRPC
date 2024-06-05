package main

import (
	"fmt"
	"geerpc"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

type Foo int

func startServer(addr chan string) {
	var foo Foo
	l, _ := net.Listen("tcp", ":9999")
	_ = GeeRPC.Register(&foo)
	GeeRPC.HandleHTTP()

	addr <- l.Addr().String()
	_ = http.Serve(l, nil)
}

func main() {
	log.SetFlags(0)
	addr := make(chan string)
	go startServer(addr)
	client, _ := GeeRPC.Dial("tcp", <-addr)

	defer func() { _ = client.Close() }()

	time.Sleep(time.Second)

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			args := fmt.Sprintf("geerpc req %d", i)
			var reply string

			if err := client.Call("Foo.Sum", args, &reply); err != nil {
				log.Fatal("call Foo.Sum error", err)
			}
			log.Println("reply:", reply)
		}(i)
		wg.Wait()
	}
}
