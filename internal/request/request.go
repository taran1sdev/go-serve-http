package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
)

type RequestLine struct {
	Method        string
	RequestTarget string
	HttpVersion   string
}

func (r *RequestLine) ValidHTTP() bool {
	return r.HttpVersion == "1.1"
}

type Request struct {
	RequestLine RequestLine
	Headers     map[string]string
	Body        []byte
}

var ERROR_MALFORMED_REQUEST_LINE = fmt.Errorf("Malformed request line")

var SEPARATOR = []byte("\r\n")

func parseRequestLine(b []byte) (*RequestLine, []byte, error) {
	idx := bytes.Index(b, SEPARATOR)
	if idx == -1 {
		return nil, b, nil
	}

	startLine := b[:idx]
	restOfMsg := b[idx+len(SEPARATOR):]

	parts := bytes.Split(startLine, []byte(" "))
	if len(parts) != 3 {
		return nil, restOfMsg, ERROR_MALFORMED_REQUEST_LINE
	}

	httpParts := bytes.Split(parts[2], []byte("/"))
	if len(httpParts) != 2 {
		return nil, restOfMsg, ERROR_MALFORMED_REQUEST_LINE
	}

	rl := &RequestLine{
		Method:        string(parts[0]),
		RequestTarget: string(parts[1]),
		HttpVersion:   string(httpParts[1]),
	}

	if !rl.ValidHTTP() {
		return nil, restOfMsg, ERROR_MALFORMED_REQUEST_LINE
	}
	return rl, restOfMsg, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, errors.Join(
			fmt.Errorf("Unable to io.ReadAll:"),
			err,
		)
	}

	rl, _, err := parseRequestLine(data)
	if err != nil {
		return nil, ERROR_MALFORMED_REQUEST_LINE
	}

	return &Request{
		RequestLine: *rl,
	}, nil

}
