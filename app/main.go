package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

type Request struct {
	Method   string
	Path     string
	Protocol string
	Header   map[string]string
}

type Response struct {
	StatusLine string
	Header     map[string]string
	Body       string
}

var router = NewRouter()

// Ensures gofmt doesn't remove the "net" and "os" imports above (feel free to remove this!)
var _ = net.Listen
var _ = os.Exit

func main() {
	flag.StringVar(&FileDirectory, "directory", "/", "File directory")
	flag.Parse()

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
	request, err := readRequest(reader)

	response, err := createResponse(request)
	if err != nil {
		fmt.Println("Error creating response:", err.Error())
		return
	}

	formattedResponse, err := formatResponse(response)
	if err != nil {
		fmt.Println("Error formatting response:", err.Error())
		return
	}

	_, err = io.WriteString(conn, formattedResponse)

	if err != nil {
		fmt.Println("Error sending response:", err.Error())
		return
	}
}

func readRequest(reader *bufio.Reader) (*Request, error) {
	requestLine, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	requestLineParts := strings.Split(strings.TrimSpace(requestLine), " ")
	if len(requestLineParts) != 3 {
		return nil, err
	}

	request := &Request{
		Method:   requestLineParts[0],
		Path:     requestLineParts[1],
		Protocol: requestLineParts[2],
		Header:   make(map[string]string),
	}

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		line = strings.TrimSpace(line)
		if line == "" {
			break
		}

		key, value, found := strings.Cut(line, ":")
		if !found {
			break
		}

		request.Header[key] = strings.TrimSpace(value)
	}

	return request, nil
}

func createResponse(request *Request) (*Response, error) {
	return router.HandleRequest(request)
}

func formatResponse(response *Response) (string, error) {
	var sb strings.Builder
	sep := "\r\n"

	/* Status line */
	sb.Write([]byte(response.StatusLine))
	sb.Write([]byte(sep))

	/* Header */
	if len(response.Header) != 0 {
		for key, value := range response.Header {
			sb.Write([]byte(key))
			sb.Write([]byte(": "))
			sb.Write([]byte(value))
			sb.Write([]byte(sep))
		}
	}
	sb.Write([]byte(sep))

	/* Body */
	sb.Write([]byte(response.Body))

	return sb.String(), nil
}
