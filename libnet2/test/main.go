package main

import (
	"fmt"
	"net"
	"time"

	"github.com/wuqifei/server_lib/libnet2/test/tcp_server"
	"github.com/wuqifei/server_lib/signal"
)

func main() {
	tcp_server.NewServer()

	time.Sleep(time.Duration(1) * time.Second)
	runclient()
	signal.InitSignal()
}

func runclient() {
	conn, err := net.Dial("tcp", "127.0.0.1:10001")
	if err != nil {
		fmt.Println("连接服务端失败:", err.Error())
		return
	}
	fmt.Println("已连接服务器")
	defer conn.Close()

	go sendData(conn)
	readData(conn)
}

func readData(conn net.Conn) {
	for {
		buf := make([]byte, 128)
		c, err := conn.Read(buf)
		if err != nil {
			fmt.Println("读取服务器数据异常:", err.Error())
		}
		fmt.Printf("[%v]:client recv:[%s] \n", time.Now(), string(buf[0:c]))
	}
}

func sendData(conn net.Conn) {
	for {
		time.Sleep(time.Duration(2) * time.Second)
		_, err := conn.Write([]byte("hi you"))
		if err != nil {
			fmt.Println("发送服务器数据异常:", err.Error())
		}
	}
}
