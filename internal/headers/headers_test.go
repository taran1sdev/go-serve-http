package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeaderParse(t *testing.T) {
	// Test: Valid single header
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers.Get("Host"))
	assert.Equal(t, 25, n)
	assert.True(t, done)

	// Test: Valid multi unique header
	headers = NewHeaders()
	data = []byte("HoSt: localhost:42069\r\nContent-Type: application/json\r\nContent-Length: 42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers.Get("Host"))
	assert.Equal(t, "application/json", headers.Get("Content-Type"))
	assert.Equal(t, "42069", headers.Get("Content-Length"))
	assert.Equal(t, 80, n)
	assert.True(t, done)

	// Test: Valid multi duplicate header
	headers = NewHeaders()
	data = []byte("Host: localhost:42069\r\nsome-header: firstOne\r\nSome-header: secondOne\r\nSome-HeAdEr: thirdOne\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers.Get("Host"))
	assert.Equal(t, "firstOne,secondOne,thirdOne", headers.Get("Some-Header"))
	assert.Equal(t, 95, n)
	assert.True(t, done)

	// Test: Invalid spacing
	headers = NewHeaders()
	data = []byte("	  	 Host : localhost:24069	   	     \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Invalid - field name not a token
	headers = NewHeaders()
	data = []byte("	  	 Ho@st: localhost:24069	   	     \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

}
