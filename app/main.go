package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

type Request struct {
	Method   string
	Path     string
	Protocol string
	Header   map[string]string
	Body     string
}

type Response struct {
	StatusLine string
	Header     map[string]string
	Body       string
}

const CONN_MAX_TIMEOUT = 5 * time.Second

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

	for {
		if err := conn.SetReadDeadline(time.Now().Add(CONN_MAX_TIMEOUT)); err != nil {
			fmt.Println("Error setting read deadline:", err.Error())
			return
		}

		request, err := readRequest(reader)
		if err != nil {
			if err == io.EOF || errors.Is(err, net.ErrClosed) {
				return
			}
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				return
			}
			fmt.Println("Error reading request:", err.Error())
			return
		}

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
		fmt.Println(formattedResponse)
		_, err = io.WriteString(conn, formattedResponse)

		if err != nil {
			fmt.Println("Error sending response:", err.Error())
			return
		}

		if request.Header["Connection"] == "close" {
			conn.Close()
		}
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
		Body:     "",
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

	if contentLengthStr, ok := request.Header["Content-Length"]; ok {
		contentLength, err := strconv.Atoi(contentLengthStr)

		if err == nil && contentLength > 0 {
			body := make([]byte, contentLength)
			_, err = io.ReadFull(reader, body)
			if err == nil {
				request.Body = string(body)
			}
		}
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
