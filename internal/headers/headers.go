package headers

import (
	"bytes"
	"fmt"
)

type Headers map[string]string

func NewHeaders() Headers {
	return make(Headers)
}

var MALFORMED_HEADERS = fmt.Errorf("Malformed HTTP Headers")

var SEPARATOR = []byte("\r\n")

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	// No CLRF means we are awaiting data
	idx := bytes.Index(data, SEPARATOR)

	if idx == -1 {
		return 0, false, nil
	}

	// CLRF at start means no headers to parse
	if idx == 0 {
		return 0, true, nil
	}

	line := data[:idx]

	f, v, found := bytes.Cut(line, []byte(":"))
	// If we have more that one : whitespace after field header is invalid
	if !found || bytes.HasSuffix(f, []byte(" ")) {
		return 0, false, MALFORMED_HEADERS
	}

	// Remove any whitespace
	f = bytes.TrimSpace(f)
	v = bytes.TrimSpace(v)

	// Add to our headers map
	h[string(f)] = string(v)

	return idx + len(SEPARATOR), false, nil

}
