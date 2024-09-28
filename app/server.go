package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"regexp"
	"strings"
)

func handleEcho(conn net.Conn, headers map[string]string, toEcho string) {
	fmt.Println(toEcho)
	respHeaders := map[string]string{"Content-Type": "text/plain"}
	if headers["accept-encoding"] == "gzip" {
		respHeaders["Content-Encoding"] = headers["accept-encoding"]
	}
	resp := ComposeResponse(200, respHeaders, toEcho)
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

func extractHeaders(requestLines []string) map[string]string {
	headers := make(map[string]string)
	for _, line := range requestLines[1:] {
		if line == "" {
			break
		}
		parts := strings.Split(line, ": ")
		headers[strings.ToLower(parts[0])] = strings.TrimSpace(parts[1])
	}
	return headers
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
	lines := stringSplitRegex.Split(string(bytes.TrimRight(req, "\x00")), -1)
	urlLineParts := strings.Split(lines[0], " ")
	verb := urlLineParts[0]
	path := urlLineParts[1]
	echoRegex, _ := regexp.Compile("/echo/([^/]+)")
	fileRegex, _ := regexp.Compile("/files/([^/]+)")
	headers := extractHeaders(lines)
	body := getBody(lines)
	if path == "/" {
		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	} else if matches := echoRegex.FindStringSubmatch(path); len(matches) > 1 && len(matches[1]) > 0 {
		handleEcho(conn, headers, matches[1])
	} else if path == "/user-agent" {
		handleUserAgent(conn, headers)
	} else if fmatches := fileRegex.FindStringSubmatch(path); len(fmatches) > 1 && len(fmatches[1]) > 0 {
		if verb == "POST" {
			createFile(body, fmatches[1])
			conn.Write([]byte("HTTP/1.1 201 Created\r\n\r\n"))
		} else {
			handleFile(conn, fmatches[1])
		}
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

func getBody(requestLines []string) []string {
	var bodyStartIndex int
	for i, line := range requestLines {
		if line == "" {
			bodyStartIndex = i + 1
		}
	}
	return requestLines[bodyStartIndex:]
}

func createFile(bodyLines []string, fileName string) error {
	filePath := fetchDirectoryArg() + fileName
	body := strings.Join(bodyLines, "\n")
	err := os.WriteFile(filePath, []byte(body), 0644)
	if err != nil {
		return err
	}
	return nil
}

func simpleResponse(status int, content string) []byte {
	resp := fmt.Sprintf("HTTP/1.1 %d OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%v", status, len(content), content)
	return []byte(resp)
}

func contentTypeResponse(status int, contentType string, content string) []byte {
	resp := fmt.Sprintf("HTTP/1.1 %d OK\r\nContent-Type: %v\r\nContent-Length: %d\r\n\r\n%v", status, contentType, len(content), content)
	return []byte(resp)
}

func ComposeResponse(status int, headers map[string]string, content string) []byte {
	codesMap := map[int]string{
		200: "OK",
		201: "Created",
		404: "Not Found",
	}
	resp := fmt.Sprintf("HTTP/1.1 %d %v\r\n", status, codesMap[status])
	var headersSlice = make([]string, 0)
	for key, value := range headers {
		headersSlice = append(headersSlice, fmt.Sprintf("%v: %v\r\n", key, value))
	}
	headersSlice = append(headersSlice, fmt.Sprintf("Content-Length: %d\r\n", len(content)))
	resp += strings.Join(headersSlice, "")
	resp += "\r\n" + content
	return []byte(resp)
}

func handleUserAgent(conn net.Conn, headers map[string]string) {
	userAgent := headers["user-agent"]
	respHeaders := map[string]string{"Content-Type": "text/plain"}
	resp := ComposeResponse(200, respHeaders, userAgent)
	conn.Write(resp)
}
