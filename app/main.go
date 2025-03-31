package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
)

var _ = net.Listen
var _ = os.Exit

func main() {
	http.HandleFunc("/", getRoot)

	err := http.ListenAndServe(":4221", nil)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("Server closed\n")
	} else if err != nil {
		fmt.Printf("Error starting server: %s\n", err)
		os.Exit(1)
	}
}

func getRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("got / request\n")
	io.WriteString(w, "HTTP/1.1 200 OK\r\n\r\n")
}
