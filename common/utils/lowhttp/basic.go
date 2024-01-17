package lowhttp

var (
	basicRequest  = []byte("GET / HTTP/1.1\r\nHost: example.com\r\n\r\n")
	basicResponse = []byte("HTTP/1.1 200 OK\r\nContent-Length: 0\r\n\r\n")
)

// BasicResponse Returns a basic HTTP response, for testing, it actually returns a b"HTTP/1.1 200 OK\r\nContent-Length: 0\r\n\r\n"
// Example:
// ```
// poc.BasicResponse() // b"HTTP/1.1 200 OK\r\nContent-Length: 0\r\n\r\n"
// ```
func BasicResponse() []byte {
	return basicResponse
}

// BasicRequest Returns a basic HTTP request, for testing, it actually returns a b"GET / HTTP/1.1\r\nHost: example.com\r\n\r\n"
// Example:
// ```
// poc.BasicRequest() // b"GET / HTTP/1.1\r\nHost: example.com\r\n\r\n"
// ```
func BasicRequest() []byte {
	return basicRequest
}
