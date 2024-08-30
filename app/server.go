package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"regexp"
	"strings"
)

func handleEcho(conn net.Conn, toEcho string) {
	fmt.Println(toEcho)
	resp := simpleResponse(200, toEcho)
	conn.Write(resp)
}

func fetchDirectoryArg() string {
	for i, arg := range os.Args {
		if arg == "--directory" && i+1 < len(os.Args) {
			path := os.Args[i+1]
			return path
		}
	}
	return ""
}

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	} else {
		fmt.Println("Now accepting connections on http://0.0.0.0:4221 ")
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleRequest(conn)
	}

}

func handleRequest(conn net.Conn) {
	defer conn.Close()
	req := make([]byte, 1024)
	_, connErr := conn.Read(req)
	if connErr != nil {
		fmt.Println("Error reading request:", connErr)
		return
	}

	stringSplitRegex := regexp.MustCompile("(\r)?\n")
	lines := stringSplitRegex.Split(string(req), -1)
	urlLineParts := strings.Split(lines[0], " ")
	path := urlLineParts[1]
	echoRegex, _ := regexp.Compile("/echo/([^/]+)")
	fileRegex, _ := regexp.Compile("/files/([^/]+)")
	if path == "/" {
		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	} else if matches := echoRegex.FindStringSubmatch(path); len(matches) > 1 && len(matches[1]) > 0 {
		handleEcho(conn, matches[1])
	} else if path == "/user-agent" {
		handleUserAgent(conn, lines)
	} else if fmatches := fileRegex.FindStringSubmatch(path); len(fmatches) > 1 && len(fmatches[1]) > 0 {
		handleFile(conn, fmatches[1])
	} else {
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	}
}

func readFile(fileName string) (string, error) {
	filePath := fetchDirectoryArg() + fileName
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	content, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func handleFile(conn net.Conn, fileName string) {
	fileContent, err := readFile(fileName)
	if err != nil {
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	}
	resp := contentTypeResponse(200, "application/octet-stream", fileContent)
	conn.Write(resp)

}

func simpleResponse(status int, content string) []byte {
	resp := fmt.Sprintf("HTTP/1.1 %d OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%v", status, len(content), content)
	return []byte(resp)
}

func contentTypeResponse(status int, contentType string, content string) []byte {
	resp := fmt.Sprintf("HTTP/1.1 %d OK\r\nContent-Type: %v\r\nContent-Length: %d\r\n\r\n%v", status, contentType, len(content), content)
	return []byte(resp)

}

func handleUserAgent(conn net.Conn, lines []string) {
	var userAgent string
	for _, line := range lines {
		before, after, found := strings.Cut(line, " ")
		if found && strings.ToLower(before) == "user-agent:" {
			userAgent = after
		}
	}
	resp := simpleResponse(200, userAgent)
	conn.Write(resp)
}
