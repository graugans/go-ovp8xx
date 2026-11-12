package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// writeHumanReadable writes a JSON document as indented list of keys and
// values. The order of the keys is preserved:
//
//	device: OVP810
//	version:
//	  firmware: 1.2.3
//	ports:
//	  - port0
//	  - port1
func writeHumanReadable(w io.Writer, document []byte) error {
	dec := json.NewDecoder(bytes.NewReader(document))
	// Keep the numbers as they are sent by the device
	dec.UseNumber()
	p := &humanPrinter{w: w, dec: dec}
	tok, err := dec.Token()
	if err != nil {
		return fmt.Errorf("malformed json: %w", err)
	}
	switch tok {
	case json.Delim('{'):
		p.object(0)
	case json.Delim('['):
		p.array(0)
	default:
		p.printf("%s\n", scalar(tok))
	}
	return p.err
}

// humanPrinter keeps the first error, so the callers do not need to check
// each single write.
type humanPrinter struct {
	w   io.Writer
	dec *json.Decoder
	err error
}

func (p *humanPrinter) printf(format string, args ...any) {
	if p.err == nil {
		_, p.err = fmt.Fprintf(p.w, format, args...)
	}
}

// token returns the next token, the second result is false at the end of the
// current object or array and in case of an error.
func (p *humanPrinter) token() (json.Token, bool) {
	if p.err != nil {
		return nil, false
	}
	tok, err := p.dec.Token()
	if err != nil {
		p.err = fmt.Errorf("malformed json: %w", err)
		return nil, false
	}
	if tok == json.Delim('}') || tok == json.Delim(']') {
		return nil, false
	}
	return tok, true
}

// object prints the members of an object, the opening brace is consumed.
func (p *humanPrinter) object(depth int) {
	indent := strings.Repeat("  ", depth)
	for {
		key, ok := p.token()
		if !ok {
			return
		}
		p.printf("%s%s:", indent, key)
		p.value(depth)
	}
}

// array prints the elements of an array, the opening bracket is consumed.
func (p *humanPrinter) array(depth int) {
	indent := strings.Repeat("  ", depth)
	for p.err == nil {
		if !p.dec.More() {
			p.token() // consume the closing bracket
			return
		}
		p.printf("%s-", indent)
		p.value(depth)
	}
}

// value prints the value following a key or a list marker.
func (p *humanPrinter) value(depth int) {
	tok, ok := p.token()
	if !ok {
		return
	}
	switch tok {
	case json.Delim('{'):
		if !p.dec.More() {
			p.printf(" {}")
		}
		p.printf("\n")
		p.object(depth + 1)
	case json.Delim('['):
		if !p.dec.More() {
			p.printf(" []")
		}
		p.printf("\n")
		p.array(depth + 1)
	default:
		p.printf(" %s\n", scalar(tok))
	}
}

// scalar returns the text of a string, number, boolean or null token.
func scalar(tok json.Token) string {
	if tok == nil {
		return "null"
	}
	return fmt.Sprint(tok)
}
