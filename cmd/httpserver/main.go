package main

import (
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"boot.taran1s/internal/request"
	"boot.taran1s/internal/response"
	"boot.taran1s/internal/server"
)

const port = 8888

func handleRequest(w io.Writer, req *request.Request) *server.HandlerError {
	hErr := &server.HandlerError{}
	switch req.RequestLine.RequestTarget {
	case "/yourproblem":
		hErr.StatusCode = response.StatusBadRequest
		hErr.Message = "Your problem not my problem\n"
	case "/myproblem":
		hErr.StatusCode = response.StatusInternalServerError
		hErr.Message = "Whoopsie, my bad\n"
	default:
		_, err := w.Write([]byte("All good\n"))
		if err != nil {
			hErr.StatusCode = response.StatusInternalServerError
			hErr.Message = "Whoopsie, my bad\n"
		} else {
			hErr = nil
		}
	}
	return hErr
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
