package main

import (
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"boot.taran1s/internal/request"
	"boot.taran1s/internal/response"
	"boot.taran1s/internal/server"
)

const port = 8888

func get200() []byte {
	return []byte(`
	<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`)
}

func get400() []byte {
	return []byte(`
	<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`)
}

func get500() []byte {
	return []byte(`
	<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`)
}

func proxyRequest(w *response.Writer, endpoint string) {
	h := response.GetDefaultHeaders(0)

	resp, err := http.Get("https://httpbin.org" + endpoint)

	if err != nil {
		w.WriteStatusLine(response.StatusInternalServerError)
		body := get500()
		h.Replace("Content-Type", "text/html")
		h.Replace("Content-Length", fmt.Sprintf("%d", len(body)))
		w.WriteHeaders(h)
		w.WriteBody(body)
	}
	tr := response.GetDefaultTrailers()

	w.WriteStatusLine(response.StatusOK)
	h.Delete("Content-Length")
	h.Set("Transfer-Encoding", "chunked")
	h.Replace("Content-Type", "text/plain")
	w.WriteHeaders(h)

	fullBody := []byte{}
	bodyLen := 0
	for {
		data := make([]byte, 1024)
		n, err := resp.Body.Read(data)
		if err != nil {
			break
		}
		fullBody = append(fullBody, data[:n]...)
		l, err := w.WriteChunkedBody(data)
		if err != nil {
			// handle error
			break
		}
		bodyLen += l
	}
	l, err := w.WriteChunkedBodyDone()
	if err != nil {
		// handle error
	}

	bodyLen += l

	tr.Replace("X-Content-SHA256", fmt.Sprintf("%x", sha256.Sum256(fullBody)))
	tr.Replace("X-Content-Length", fmt.Sprintf("%d", len(fullBody)))

	w.WriteTrailers(tr)
	return
}

func handleRequest(w *response.Writer, req *request.Request) {
	// This is just for testing but probably a better way to implement this?
	h := response.GetDefaultHeaders(0)
	body := get200()
	status := response.StatusOK

	switch req.RequestLine.RequestTarget {
	case "/yourproblem":
		body = get400()
		status = response.StatusBadRequest
	case "/myproblem":
		body = get500()
		status = response.StatusInternalServerError
	}

	// To test chunked encoding we will proxy requests to httpbin.org
	if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin") {
		endpoint := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin")
		fmt.Println("Proxying request...")
		proxyRequest(w, endpoint)
		return
	}

	w.WriteStatusLine(status)
	h.Replace("Content-Length", fmt.Sprintf("%d", len(body)))
	h.Replace("Content-Type", "text/html")
	w.WriteHeaders(h)
	w.WriteBody(body)
}

func main() {
	server, err := server.Serve(port, handleRequest)
	if err != nil {
		log.Fatal("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server stopped")
}
