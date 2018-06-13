package yaml

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/renstrom/dedent"
	"github.com/stretchr/testify/require"
)

func TestDecoder(t *testing.T) {
	doc := dedent.Dedent(`
		# a comment
		---
		document: 1
		... # document suffix
		---
		["document", "2"]
		---
		document 3
		---
		4  # just an integer`)

	dec := NewDecoder(bytes.NewReader([]byte(doc)))
	var decoded interface{}

	err := dec.Decode(&decoded)
	if err != nil {
		t.Errorf("%s", err)
		return
	}
	require.Equal(t, fmt.Sprintf("%#v", decoded), "map[string]interface {}{\"document\":1}")

	err = dec.Decode(&decoded)
	if err != nil {
		t.Errorf("%s", err)
		return
	}
	require.Equal(t, fmt.Sprintf("%#v", decoded), "[]interface {}{\"document\", \"2\"}")

	err = dec.Decode(&decoded)
	if err != nil {
		t.Errorf("%s", err)
		return
	}
	require.Equal(t, fmt.Sprintf("%#v", decoded), "\"document 3\"")

	err = dec.Decode(&decoded)
	if err != nil {
		t.Errorf("%s", err)
		return
	}
	require.Equal(t, fmt.Sprintf("%#v", decoded), "4")

	err = dec.Decode(&decoded)
	if err == nil {
		t.Errorf("expected io.EOF, got %#v", decoded)
		return
	}
	if err != io.EOF {
		t.Errorf("%s", err)
		return
	}
}
