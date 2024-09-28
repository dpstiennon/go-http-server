package main

import (
	"strings"
	"testing"
)

func TestComposeResponse(t *testing.T) {
	tests := []struct {
		status   int
		headers  map[string]string
		content  string
		expected string
	}{
		{
			status:   200,
			headers:  map[string]string{"Content-Type": "text/plain"},
			content:  "Hello, World!",
			expected: "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: 13\r\n\r\nHello, World!",
		},
		{
			status:   201,
			headers:  map[string]string{"Content-Type": "application/json"},
			content:  `{"message": "created"}`,
			expected: "HTTP/1.1 201 Created\r\nContent-Type: application/json\r\nContent-Length: 22\r\n\r\n{\"message\": \"created\"}",
		},
		{
			status:   404,
			headers:  map[string]string{"Content-Type": "text/html"},
			content:  "<h1>Not Found</h1>",
			expected: "HTTP/1.1 404 Not Found\r\nContent-Type: text/html\r\nContent-Length: 18\r\n\r\n<h1>Not Found</h1>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := ComposeResponse(tt.status, tt.headers, tt.content)
			if strings.TrimSpace(string(result)) != strings.TrimSpace(tt.expected) {
				t.Errorf("\nexpected %v, \ngot\n %v", tt.expected, string(result))
			}
		})
	}
}
