package main

import (
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"
)

func handleEcho(conn net.Conn, toEcho string) {
	fmt.Println(toEcho)
	resp := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%v", len(toEcho), toEcho)
	conn.Write([]byte(resp))
}

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	// Uncomment this block to pass the first stage
	//
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}

	req := make([]byte, 1024)
	conn.Read(req)
	if err != nil {
		fmt.Println("Error reading request:", err)
		return
	}

	lines := strings.Split(string(req), "\n")
	urlLineParts := strings.Split(lines[0], " ")
	path := urlLineParts[1]
	echoRegex, _ := regexp.Compile("/echo/([^/]+)")
	if path == "/" {
		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	} else if matches := echoRegex.FindStringSubmatch(path); len(matches) > 1 && len(matches[1]) > 0 {
		handleEcho(conn, matches[1])
	} else {
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	}
	conn.Close()

}
