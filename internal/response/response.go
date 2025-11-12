package response

import (
	"fmt"
	"io"
	"net"
	"strconv"

	"boot.taran1s/internal/headers"
)

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

var statusName = map[StatusCode]string{
	StatusOK:                  "OK",
	StatusBadRequest:          "Bad Request",
	StatusInternalServerError: "Internal Server Error",
}

func (sc StatusCode) String() string {
	return statusName[sc]
}

type responseState int

const (
	stateStatus   responseState = 0
	stateHeaders  responseState = 1
	stateBody     responseState = 2
	stateTrailers responseState = 3
	stateDone     responseState = 4
)

type Writer struct {
	writerState responseState
	writer      io.Writer
}

// do we pass the connection to write here?
func NewWriter(conn net.Conn) *Writer {
	return &Writer{
		writerState: stateStatus,
		writer:      conn,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.writerState != stateStatus {
		return fmt.Errorf("Status line already written")
	}
	status := string(fmt.Sprintf("HTTP/1.1 %d %s \r\n", statusCode, statusCode))
	_, err := w.writer.Write([]byte(status))
	if err != nil {
		return err
	}

	w.writerState = stateHeaders
	return nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()

	h.Set("Content-Length", strconv.Itoa(contentLen))
	h.Set("Connection", "close")
	h.Set("Content-type", "text/plain")
	return *h
}

func GetDefaultTrailers() headers.Headers {
	h := headers.NewHeaders()

	h.Set("X-Content-SHA256", "")
	h.Set("X-Content-Length", "")
	return *h
}

func (w *Writer) WriteHeaders(h headers.Headers) error {
	if w.writerState != stateHeaders {
		if w.writerState == stateStatus {
			return fmt.Errorf("Missing Status Line")
		}
		return fmt.Errorf("Headers already written")
	}

	headers := h.GetAll()

	for k, v := range headers {
		_, err := w.writer.Write([]byte(fmt.Sprintf("%s: %s\r\n", k, v)))
		if err != nil {
			return err
		}
	}

	_, err := w.writer.Write([]byte("\r\n"))
	if err != nil {
		return err
	}
	w.writerState = stateBody
	return nil
}

func (w *Writer) WriteBody(b []byte) (int, error) {
	if w.writerState != stateBody {
		return 0, fmt.Errorf("Must write status line and headers before the body")
	}
	bytes, err := w.writer.Write(b)
	if err != nil {
		return 0, err
	}

	w.writerState = stateDone
	return bytes, nil
}

// This works like write body except it doesn't change the writer state
func (w *Writer) WriteChunkedBody(b []byte) (int, error) {
	if w.writerState != stateBody {
		return 0, fmt.Errorf("Must write status line and headers before the body")
	}
	n, lErr := w.writer.Write([]byte(fmt.Sprintf("%x\r\n", len(b))))
	if lErr != nil {
		return 0, fmt.Errorf("Failed to write chunk length to body")
	}

	bytes, dErr := w.writer.Write(append(b, []byte("\r\n")...))
	if dErr != nil {
		return n, fmt.Errorf("Failed to write chunked data to body")
	}

	return n + bytes, nil
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	if w.writerState != stateBody {
		return 0, fmt.Errorf("This should never happen...")
	}

	bytes, err := w.writer.Write([]byte("0\r\n\r\n"))
	if err != nil {
		return 0, fmt.Errorf("Failed to write end of chunked data to body")
	}

	w.writerState = stateTrailers
	return bytes, nil
}

func (w *Writer) WriteTrailers(h headers.Headers) error {
	for k, v := range h.GetAll() {
		_, err := w.writer.Write([]byte(fmt.Sprintf("%s: %s\r\n", k, v)))
		if err != nil {
			return fmt.Errorf("Failed to write trailers")
		}
	}
	_, err := w.writer.Write([]byte("\r\n"))
	if err != nil {
		return fmt.Errorf("Failed to write end of request")
	}
	w.writerState = stateDone
	return nil
}
