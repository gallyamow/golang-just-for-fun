package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

func main() {
	go tcpEcho(8081)
	go udpEcho(8082)

	// @idiomatic: block forever
	select {} // бесконечное ожидание
}

// Usage telnet localhost 8081
func tcpEcho(port int) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal(err)
	}
	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			log.Printf("failed to close listener: %v", err)
		}
	}(listener)

	handleConnection := func(conn net.Conn) {
		defer func(conn net.Conn) {
			err := conn.Close()
			if err != nil {
				log.Printf("failed to close connection: %v", err)
			}
		}(conn)

		// Используя сnn.Read() можно читать в []byte, но это не удобно.
		//
		// 1) Если читать строки, rune, slice - можно использовать bufio.Reader.
		// rdr := bufio.NewReader(conn)
		// rdr.ReadString('\n') // именно rune
		//
		// 2) Если надо читать потоковые данные, то лучше conn.Read + проверка err == io.EOF в цикле.
		//
		// 3) Если надо прочитать все - то можно использовать ReadAll. Но оно завершится только тогда когда соединение
		// будет закрыто.
		//data, err := io.ReadAll(conn)

		rdr := bufio.NewReader(conn)
		data, err := rdr.ReadString('\n')
		if err != nil {
			log.Println(err)
		}

		_, err = conn.Write([]byte(data))
		if err != nil {
			log.Println(err)
		}
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		go handleConnection(conn)
	}
}

// Usage nc -u 127.0.0.1 8082
func udpEcho(port int) {
	addr := net.UDPAddr{
		Port: port,
		IP:   net.ParseIP("127.0.0.1"),
	}

	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		log.Fatal(err)
	}
	defer func(conn *net.UDPConn) {
		err := conn.Close()
		if err != nil {
			log.Printf("failed to close connection: %v", err)
		}
	}(conn)

	buffer := make([]byte, 1024)
	for {
		n, clientAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Println(err)
			continue
		}

		_, err = conn.WriteToUDP(buffer[:n], clientAddr)
		if err != nil {
			log.Println(err)
		}
	}
}
