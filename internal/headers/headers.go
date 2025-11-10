package headers

import (
	"bytes"
	"fmt"
	"strings"
)

func isToken(str []byte) bool {
	for _, ch := range str {
		found := false
		if ch >= 'A' && ch <= 'Z' ||
			ch >= 'a' && ch <= 'z' ||
			ch >= '0' && ch <= '9' {
			found = true
		}
		switch ch {
		case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~':
			found = true
		}

		if !found {
			return false
		}
	}

	return true
}

type Headers struct {
	headers map[string]string
}

func NewHeaders() *Headers {
	return &Headers{
		headers: map[string]string{},
	}
}

var MALFORMED_FIELD_LINE = fmt.Errorf("Malformed Field Line")
var MALFORMED_FIELD_NAME = fmt.Errorf("Malformed Field Name")

func parseHeader(fieldLine []byte) (string, string, error) {
	parts := bytes.SplitN(fieldLine, []byte(":"), 2)
	if len(parts) != 2 {
		return "", "", MALFORMED_FIELD_LINE
	}

	if bytes.HasSuffix(parts[0], []byte(" ")) {
		return "", "", MALFORMED_FIELD_NAME
	}

	name := bytes.TrimSpace(parts[0])
	value := bytes.TrimSpace(parts[1])

	return string(name), string(value), nil
}

var SEPARATOR = []byte("\r\n")

func (h *Headers) Get(name string) string {
	return h.headers[strings.ToLower(name)]
}

func (h *Headers) Set(name, value string) {
	name = strings.ToLower(name)

	if v, ok := h.headers[name]; ok {
		h.headers[name] = fmt.Sprintf("%s,%s", v, value)
	} else {
		h.headers[name] = value
	}
}

func (h *Headers) Parse(data []byte) (int, bool, error) {
	// No CLRF means we are awaiting data
	read := 0
	done := false

	for {
		idx := bytes.Index(data[read:], SEPARATOR)

		if idx == -1 {
			break
		}

		if idx == 0 {
			done = true
			read += len(SEPARATOR)
			break
		}

		name, value, err := parseHeader(data[read : read+idx])
		if err != nil {
			return 0, done, err
		}

		if !isToken([]byte(name)) {
			return 0, false, MALFORMED_FIELD_NAME
		}

		read += idx + len(SEPARATOR)
		h.Set(name, value)
	}

	return read, done, nil
}
