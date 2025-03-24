package jsondiffhelpers

import (
	"errors"
	"strings"
)

const (
	separator    = '/'
	escapeSlash  = "~1"
	escapeTilde  = "~0"
	emptyPointer = ""
)

var (
	// rfc6901Escaper is a replacer that escapes a JSON Pointer string
	// in compliance with the JavaScript Object Notation Pointer syntax.
	// https://tools.ietf.org/html/rfc6901
	//rfc6901Escaper = strings.NewReplacer("~", escapeTilde, "/", escapeSlash)

	// rfc6901Unescaper is a replacer that unescape a JSON Pointer string.
	rfc6901Unescaper = strings.NewReplacer(escapeTilde, "~", escapeSlash, "/")
)

var (
	errLeadingSlash             = errors.New("no leading slash")
	errIncompleteEscapeSequence = errors.New("incomplete escape sequence")
	errInvalidEscapeSequence    = errors.New("invalid escape sequence")
)

func parsePointer(s string) ([]string, error) {
	if s == "" {
		return nil, nil
	}
	a := []rune(s)

	if len(a) > 0 && a[0] != '/' {
		return nil, errLeadingSlash
	}
	var tokens []string

	ls := 0
	for i, r := range a {
		switch {
		case r == '/':
			if i != 0 {
				tokens = append(tokens, string(a[ls+1:i]))
			}
			if i == len(a)-1 {
				// Last char is a '/', next fragment is an empty string.
				tokens = append(tokens, "")
				break
			}
			ls = i
		case r == '~':
			if i == len(a)-1 {
				return nil, errIncompleteEscapeSequence
			}
			if a[i+1] != '0' && a[i+1] != '1' {
				return nil, errInvalidEscapeSequence
			}
		case i == len(a)-1:
			// End of string, accumulate from last separator.
			tokens = append(tokens, string(a[ls+1:]))
		}
	}
	return tokens, nil
}
