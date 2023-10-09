package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
)

type Calculator int

func (c *Calculator) Add(args *Args, reply *int) error {
	*reply = args.A + args.B
	print(*reply)
	return nil
}

type Args struct {
	A, B int
}

func main() {
	calculator := new(Calculator)
	rpc.Register(calculator)
	listener, err := net.Listen("tcp", ":1234")
	if err != nil {
		log.Fatal("Listen error:", err)
	}
	fmt.Println("Server is listening on port 1234...")
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("Accept error:", err)
			continue
		}
		//建立链接
		go rpc.ServeConn(conn)
	}
}
