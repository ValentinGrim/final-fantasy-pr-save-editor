package file

import (
	"bytes"
	"compress/flate"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/kiamev/ffpr-save-cypher/rijndael"
)

// PC files (slots and the bestiary) are: optional UTF-8 BOM, base64 text
// (the game ends it with CRLF), Rijndael-256 CBC with zero padding, raw
// DEFLATE, JSON.

const cipherBlockSize = 32

var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

// decode turns the content of a PC file into its JSON payload. prefix is the
// BOM found in front of the base64 text (nil if there was none); encode puts
// it back.
func decode(b []byte) (out []byte, prefix []byte, err error) {
	if bytes.HasPrefix(b, utf8BOM) {
		prefix = append([]byte(nil), utf8BOM...)
		b = b[len(utf8BOM):]
	}
	b = bytes.TrimRight(b, " \t\r\n")
	if r := len(b) % 4; r != 0 {
		b = append(b[:len(b):len(b)], bytes.Repeat([]byte{'='}, 4-r)...)
	}

	var cipher []byte
	if cipher, err = base64.StdEncoding.DecodeString(string(b)); err != nil {
		return nil, nil, fmt.Errorf("base64: %w", err)
	}
	if len(cipher) == 0 || len(cipher)%cipherBlockSize != 0 {
		return nil, nil, fmt.Errorf("encrypted data is %d bytes, not a non-zero multiple of %d", len(cipher), cipherBlockSize)
	}

	var plain []byte
	if plain, err = rijndael.New().Decrypt(cipher); err != nil {
		return nil, nil, fmt.Errorf("decrypt: %w", err)
	}
	// Decrypt removes the zero padding by dropping every trailing zero byte
	// of the last block, including zero bytes that belong to the deflate
	// stream itself (a stream often ends with 0x00, e.g. an end-of-block
	// code of a fixed Huffman block). Inflating that truncated stream fails
	// with io.ErrUnexpectedEOF. The dropped bytes were all zeros, so put them
	// back: the inflater stops at the final block and ignores the padding,
	// like the game does.
	stripped := len(plain)
	plain = append(plain[:stripped:stripped], make([]byte, len(cipher)-stripped)...)

	if out, err = inflate(plain); err == nil {
		return out, prefix, nil
	}
	// A stream that was only sync-flushed (no final block) runs into the
	// padding (corrupt data) or the end of the input. Without the padding it
	// ends on a block boundary: accept it if everything up to there is a
	// complete JSON document.
	if alt, altErr := inflate(plain[:stripped]); errors.Is(altErr, io.ErrUnexpectedEOF) && len(alt) > 0 && json.Valid(alt) {
		return alt, prefix, nil
	}
	return nil, nil, fmt.Errorf("inflate: %w", err)
}

func inflate(b []byte) ([]byte, error) {
	zr := flate.NewReader(bytes.NewReader(b))
	defer func() { _ = zr.Close() }()
	return io.ReadAll(zr)
}

// encode is the inverse of decode. The deflate stream is terminated with a
// final block.
func encode(data []byte, prefix []byte) ([]byte, error) {
	var b bytes.Buffer
	zw, err := flate.NewWriter(&b, 6)
	if err != nil {
		return nil, err
	}
	if _, err = zw.Write(data); err != nil {
		return nil, err
	}
	if err = zw.Close(); err != nil {
		return nil, err
	}
	return seal(b.Bytes(), prefix)
}

// seal encrypts an already compressed stream and formats it as a PC file.
func seal(compressed []byte, prefix []byte) ([]byte, error) {
	// Encrypt pads in place; never let it write into the caller's array.
	cipher, err := rijndael.New().Encrypt(compressed[:len(compressed):len(compressed)])
	if err != nil {
		return nil, fmt.Errorf("encrypt: %w", err)
	}
	s := base64.StdEncoding.EncodeToString(cipher)
	out := make([]byte, 0, len(prefix)+len(s))
	out = append(out, prefix...)
	return append(out, s...), nil
}
