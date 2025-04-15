package main

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var FileDirectory string

type HandlerFunc func(req *Request) (*Response, error)
type PrefixRoute struct {
	Prefix  string
	Handler HandlerFunc
}

type Router struct {
	routes       map[string]HandlerFunc
	prefixRoutes []PrefixRoute
}

func NewRouter() *Router {
	return &Router{
		routes: map[string]HandlerFunc{
			"/":           handleRoot,
			"/user-agent": handleUserAgent,
		},
		prefixRoutes: []PrefixRoute{
			{"/echo/", handleEcho},
			{"/files/", handleFiles},
		},
	}
}

func handleRoot(request *Request) (*Response, error) {
	return &Response{
		StatusLine: "HTTP/1.1 200 OK",
		Header:     make(map[string]string),
		Body:       "",
	}, nil
}

func handleUserAgent(request *Request) (*Response, error) {
	body := request.Header["User-Agent"]
	return &Response{
		StatusLine: "HTTP/1.1 200 OK",
		Header: map[string]string{
			"Content-Type":   "text/plain",
			"Content-Length": strconv.Itoa(len(body)),
		},
		Body: body,
	}, nil
}

func handleEcho(request *Request) (*Response, error) {
	body, _ := strings.CutPrefix(request.Path, "/echo/")

	return &Response{
		StatusLine: "HTTP/1.1 200 OK",
		Header: map[string]string{
			"Content-Type":   "text/plain",
			"Content-Length": strconv.Itoa(len(body)),
		},
		Body: body,
	}, nil
}

func handleFiles(request *Request) (*Response, error) {
	body, _ := strings.CutPrefix(request.Path, "/files/")

	if strings.EqualFold(request.Method, "POST") {
		file, _ := os.Create(filepath.Join(FileDirectory, body))
		file.Write([]byte(request.Body))
		return &Response{
			StatusLine: "HTTP/1.1 201 Created",
			Header:     make(map[string]string),
			Body:       "",
		}, nil
	}

	content, err := os.ReadFile(filepath.Join(FileDirectory, body))
	if err != nil {
		return &Response{
			StatusLine: "HTTP/1.1 404 Not Found",
			Header:     make(map[string]string),
			Body:       "",
		}, nil
	}

	return &Response{
		StatusLine: "HTTP/1.1 200 OK",
		Header: map[string]string{
			"Content-Type":   "application/octet-stream",
			"Content-Length": strconv.Itoa(len(content)),
		},
		Body: string(content),
	}, nil
}

func (r *Router) HandleRequest(request *Request) (*Response, error) {
	if handler, ok := r.routes[request.Path]; ok {
		return handler(request)
	}

	for _, route := range r.prefixRoutes {
		if strings.HasPrefix(request.Path, route.Prefix) {
			return route.Handler(request)
		}
	}

	return &Response{
		StatusLine: "HTTP/1.1 404 Not Found",
		Header:     make(map[string]string),
		Body:       "",
	}, nil
}
