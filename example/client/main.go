package main

import (
	"fmt"
	"net"
	"time"

	"github.com/wuqifei/server_lib/libio"

	"github.com/wuqifei/server_lib/signal"
)

func main() {
	server := "127.0.0.1:6868"
	tcpAddr, _ := net.ResolveTCPAddr("tcp4", server)
	conn, _ := net.DialTCP("tcp", nil, tcpAddr)
	go recv(conn)
	go send(conn)
	signal.InitSignal()
}

func send(conn *net.TCPConn) {
	for {

		lengthb := make([]byte, 2)
		libio.PutUint16BE(lengthb, 7)
		b := make([]byte, 2)
		libio.PutUint16BE(b, 100)
		lengthb = append(lengthb, b...)
		lengthb = append(lengthb, ([]byte)("abcde")...)
		conn.Write(lengthb)
		fmt.Printf("send msg :%v\n", lengthb)
		time.Sleep(time.Duration(20) * time.Second)
	}
}

func recv(conn *net.TCPConn) {
	for {
		buf := make([]byte, 1024)
		length, _ := conn.Read(buf)
		data := buf[:length]

		fmt.Printf("受到消息:%v \n", data)
	}
}
