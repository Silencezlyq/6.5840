package main

import (
	"fmt"
	"log"
	"net/rpc"
)

type Args struct {
	A, B int
}

func main() {
	client, err := rpc.Dial("tcp", "localhost:1234")
	if err != nil {
		log.Fatal("Dial error:", err)
	}
	args := &Args{A: 10, B: 20}
	var reply int
	err = client.Call("Calculator.Add", args, &reply)
	if err != nil {
		log.Fatal("Call error:", err)
	}
	fmt.Println("Result:", reply)
}
