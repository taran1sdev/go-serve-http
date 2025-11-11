package server

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"

	"boot.taran1s/internal/request"
	"boot.taran1s/internal/response"
)

type Server struct {
	listener net.Listener
	handler  Handler
	closed   bool
}

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

func (h *HandlerError) Write(conn net.Conn) {
	response.WriteStatusLine(conn, h.StatusCode)

	b := []byte(h.Message)

	head := response.GetDefaultHeaders(len(b))

	response.WriteHeaders(conn, head)
	response.WriteBody(conn, b)
}

type Handler func(w io.Writer, req *request.Request) *HandlerError

func runConnection(s *Server, conn io.ReadWriteCloser) {

	response.WriteStatusLine(conn, response.StatusOK)

	b := []byte("Hello World!")

	h := response.GetDefaultHeaders(len(b))

	response.WriteHeaders(conn, h)
	response.WriteBody(conn, b)

	conn.Close()
}

func runServer(s *Server, listener net.Listener) {

	conn, err := listener.Accept()
	if err != nil || s.closed {
		return
	}

	go runConnection(s, conn)
}

func (s *Server) Close() error {
	s.closed = true
	return nil
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.closed {
				return
			}
			log.Printf("Error accepting connection: %v", err)
			continue
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	req, err := request.RequestFromReader(conn)
	if err != nil {
		hErr := &HandlerError{
			StatusCode: response.StatusBadRequest,
			Message:    err.Error(),
		}
		hErr.Write(conn)
	}
	buf := bytes.NewBuffer([]byte{})
	hErr := s.handler(buf, req)
	if hErr != nil {
		hErr.Write(conn)
		return
	}
	b := buf.Bytes()
	response.WriteStatusLine(conn, response.StatusOK)
	headers := response.GetDefaultHeaders(len(b))
	response.WriteHeaders(conn, headers)
	response.WriteBody(conn, b)
	return
}

func Serve(port uint16, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	s := &Server{
		handler:  handler,
		listener: listener,
	}

	go s.listen()
	return s, nil
}
