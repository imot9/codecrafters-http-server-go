package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
)

// Ensures gofmt doesn't remove the "net" and "os" imports above (feel free to remove this!)
var _ = net.Listen
var _ = os.Exit

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "127.0.0.1:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go handleConnection(conn)
	}

}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	for {
		line, err := reader.ReadString('\n')
		if err != nil || line == "\r\n" {
			break
		}
	}

	response := "HTTP/1.1 200 OK\r\n\r\n"
	_, err := io.WriteString(conn, response)

	if err != nil {
		fmt.Println("Error sending response:", err.Error())
		return
	}
}
