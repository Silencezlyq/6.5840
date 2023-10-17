package main

import (
	"fmt"
	"net"
	"time"
)

func main() {
	listen, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("Error listening:", err)
		return
	}
	defer listen.Close()

	for {
		conn, err := listen.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	// 设置连接超时时间为10秒
	conn.SetDeadline(time.Now().Add(10 * time.Second))

	buf := make([]byte, 1024)
	_, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading from connection:", err)
		return
	}
	// 处理接收到的数据
	fmt.Printf("Received data: %s\n", string(buf))
}

