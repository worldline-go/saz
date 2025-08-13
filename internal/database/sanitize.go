package database

import (
	"bytes"
	"io"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

const (
	EncodingISO88591 = "ISO 8859-1"
)

func SanitizeString(s string) string {
	if utf8.ValidString(s) {
		return s
	}

	var b strings.Builder

	for i, r := range s {
		if r == utf8.RuneError {
			_, size := utf8.DecodeRuneInString(s[i:])
			if size == 1 {
				// skip invalid byte
				b.WriteRune('�')
				continue
			}
		}
		b.WriteRune(r)
	}

	return b.String()
}

func ConvertISO88591ToUTF8(data []byte) (string, error) {
	decoder := charmap.ISO8859_1.NewDecoder()
	reader := transform.NewReader(bytes.NewReader(data), decoder)

	utf8Bytes, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}

	return string(utf8Bytes), nil
}
