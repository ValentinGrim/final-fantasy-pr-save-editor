package file

import (
	"bytes"
	"compress/flate"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kiamev/ffpr-save-cypher/rijndael"
	"pixel-remastered-save-editor/global"
	"pixel-remastered-save-editor/save"
)

// syntheticBestiary builds a bestiary document shaped like the game's, with n
// defeated monsters.
func syntheticBestiary(n int) []byte {
	flags := make([]string, 1024)
	for i := range flags {
		flags[i] = "0"
		if i%7 == 0 {
			flags[i] = "1"
		}
	}
	keys := make([]string, n)
	values := make([]string, n)
	total := 0
	for i := 0; i < n; i++ {
		keys[i] = fmt.Sprint((i*37)%n + 1)
		values[i] = fmt.Sprint(i%50 + 1)
		total += i%50 + 1
	}
	return []byte(fmt.Sprintf(`{"scenarioFlags":"{\"target\":[%s]}","monsterDefeats":"{\"keys\":[%s],\"values\":[%s]}","totalSubjugationCount":%d}`,
		strings.Join(flags, ","), strings.Join(keys, ","), strings.Join(values, ","), total))
}

func deflateSyncFlushed(t *testing.T, data []byte) []byte {
	t.Helper()
	var b bytes.Buffer
	zw, err := flate.NewWriter(&b, 6)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = zw.Write(data); err != nil {
		t.Fatal(err)
	}
	// Flush without Close: the stream ends with a sync marker, no final block.
	if err = zw.Flush(); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

func TestRoundTrip(t *testing.T) {
	for _, prefix := range [][]byte{nil, utf8BOM} {
		for _, n := range []int{0, 1, 75, 128, 300} {
			payload := syntheticBestiary(n)
			enc, err := encode(payload, prefix)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.HasPrefix(enc, prefix) {
				t.Fatalf("prefix not kept")
			}
			// The game terminates its files with CRLF.
			out, gotPrefix, err := decode(append(enc, '\r', '\n'))
			if err != nil {
				t.Fatalf("n=%d bom=%v: %v", n, prefix != nil, err)
			}
			if !bytes.Equal(out, payload) || !bytes.Equal(gotPrefix, prefix) {
				t.Fatalf("n=%d bom=%v: round trip mismatch", n, prefix != nil)
			}
		}
	}
}

// A deflate stream that ends with zero bytes lost them to the zero padding
// removal of Decrypt, and inflating failed with io.ErrUnexpectedEOF.
func TestDecodeStreamEndingWithZeroByte(t *testing.T) {
	for name, payload := range map[string][]byte{
		"json": syntheticBestiary(128),
		// Not JSON: must decode without the missing-final-block fallback.
		"text": bytes.Repeat([]byte("not json "), 100),
	} {
		// Sync flush then an empty final fixed-Huffman block (bits 1, 01,
		// then the 7-bit end-of-block code 0000000): 0x03 0x00, as zlib
		// writes it.
		stream := append(deflateSyncFlushed(t, payload), 0x03, 0x00)
		if out, err := inflate(stream); err != nil || !bytes.Equal(out, payload) {
			t.Fatalf("%s: test stream is not valid: %v", name, err)
		}
		enc, err := seal(stream, utf8BOM)
		if err != nil {
			t.Fatal(err)
		}

		// The failure mode of the old pipeline.
		cipher, _ := base64.StdEncoding.DecodeString(string(enc[len(utf8BOM):]))
		plain, _ := rijndael.New().Decrypt(cipher)
		if _, err = inflate(plain); !errors.Is(err, io.ErrUnexpectedEOF) {
			t.Fatalf("%s: expected the stripped stream to fail with unexpected EOF, got %v", name, err)
		}

		out, _, err := decode(enc)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if !bytes.Equal(out, payload) {
			t.Fatalf("%s: payload mismatch", name)
		}
	}
}

func TestDecodeStreamWithoutFinalBlock(t *testing.T) {
	var padded, unpadded bool
	for n := 1; n < 200 && !(padded && unpadded); n++ {
		payload := syntheticBestiary(n)
		stream := deflateSyncFlushed(t, payload)
		if len(stream)%cipherBlockSize == 0 {
			unpadded = true
		} else {
			padded = true
		}
		enc, err := seal(stream, nil)
		if err != nil {
			t.Fatal(err)
		}
		out, _, err := decode(enc)
		if err != nil {
			t.Fatalf("n=%d (stream %d bytes): %v", n, len(stream), err)
		}
		if !bytes.Equal(out, payload) {
			t.Fatalf("n=%d: payload mismatch", n)
		}
	}
	if !padded || !unpadded {
		t.Fatalf("did not cover both padded and unpadded streams (padded=%v unpadded=%v)", padded, unpadded)
	}
}

func TestDecodeErrors(t *testing.T) {
	good, err := encode(syntheticBestiary(10), utf8BOM)
	if err != nil {
		t.Fatal(err)
	}
	truncated, err := seal(deflateSyncFlushed(t, syntheticBestiary(10))[:40], nil)
	if err != nil {
		t.Fatal(err)
	}
	for name, b := range map[string][]byte{
		"empty":           {},
		"not base64":      []byte("not base64 at all!"),
		"partial block":   []byte(base64.StdEncoding.EncodeToString(make([]byte, 40))),
		"cut ciphertext":  good[:len(good)-44],
		"truncated input": truncated,
	} {
		if _, _, err := decode(b); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}

const syntheticSlot = `{"id":1,"pictureData":"","userData":"{}","configData":"{}","dataStorage":"{}","mapData":"{}","timeStamp":"t","playTime":1.5,"clearFlag":0,"isCompleteFlag":0}`

func writeEncoded(t *testing.T, path string, payload []byte) []byte {
	t.Helper()
	b, err := encode(payload, utf8BOM)
	if err != nil {
		t.Fatal(err)
	}
	b = append(b, '\r', '\n')
	if err = os.WriteFile(path, b, 0o644); err != nil {
		t.Fatal(err)
	}
	return b
}

func TestLoadSaveBestiary(t *testing.T) {
	t.Setenv("PR_PRINT", "false")
	const slot = "slot"

	t.Run("missing", func(t *testing.T) {
		dir := t.TempDir()
		writeEncoded(t, filepath.Join(dir, slot), []byte(syntheticSlot))
		data, err := LoadSave(global.One, dir, slot, global.PC)
		if err != nil {
			t.Fatal(err)
		}
		if b, _ := data.Bestiary(); b != nil {
			t.Fatal("no bestiary expected")
		}
		if err = SaveSave(global.One, data, 1, dir, slot, global.PC); err != nil {
			t.Fatal(err)
		}
		if _, err = os.Stat(filepath.Join(dir, bestiaryFile)); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("bestiary file must not be created: %v", err)
		}
	})

	t.Run("unreadable", func(t *testing.T) {
		dir := t.TempDir()
		writeEncoded(t, filepath.Join(dir, slot), []byte(syntheticSlot))
		garbage := []byte("\xEF\xBB\xBFnot-a-bestiary\r\n")
		if err := os.WriteFile(filepath.Join(dir, bestiaryFile), garbage, 0o644); err != nil {
			t.Fatal(err)
		}
		data, err := LoadSave(global.One, dir, slot, global.PC)
		if err != nil {
			t.Fatal(err)
		}
		if b, _ := data.Bestiary(); b != nil {
			t.Fatal("no bestiary expected")
		}
		if err = data.SetBestiary(&save.Bestiary{MonsterDefeatsInternal: `{"keys":[1],"values":[1]}`, TotalSubjugationCount: 1}); err == nil {
			t.Fatal("setting a bestiary that was not loaded must fail")
		}
		if err = SaveSave(global.One, data, 1, dir, slot, global.PC); err != nil {
			t.Fatal(err)
		}
		if b, _ := os.ReadFile(filepath.Join(dir, bestiaryFile)); !bytes.Equal(b, garbage) {
			t.Fatal("unreadable bestiary was overwritten")
		}
	})

	t.Run("unchanged and changed", func(t *testing.T) {
		dir := t.TempDir()
		writeEncoded(t, filepath.Join(dir, slot), []byte(syntheticSlot))
		payload := syntheticBestiary(20)
		orig := writeEncoded(t, filepath.Join(dir, bestiaryFile), payload)

		data, err := LoadSave(global.Two, dir, slot, global.PC)
		if err != nil {
			t.Fatal(err)
		}
		b, err := data.Bestiary()
		if err != nil || b == nil {
			t.Fatalf("bestiary not loaded: %v", err)
		}
		// Re-setting the same content must not count as a change.
		if err = data.SetBestiary(b); err != nil {
			t.Fatal(err)
		}
		if err = SaveSave(global.Two, data, 1, dir, slot, global.PC); err != nil {
			t.Fatal(err)
		}
		if got, _ := os.ReadFile(filepath.Join(dir, bestiaryFile)); !bytes.Equal(got, orig) {
			t.Fatal("unchanged bestiary was rewritten")
		}

		b.TotalSubjugationCount++
		if err = data.SetBestiary(b); err != nil {
			t.Fatal(err)
		}
		if err = SaveSave(global.Two, data, 1, dir, slot, global.PC); err != nil {
			t.Fatal(err)
		}
		raw, _ := os.ReadFile(filepath.Join(dir, bestiaryFile))
		if !bytes.HasPrefix(raw, utf8BOM) {
			t.Fatal("BOM not kept")
		}
		got, _, err := decode(raw)
		if err != nil {
			t.Fatal(err)
		}
		if bytes.Equal(got, payload) || !bytes.Equal(got, data.BestiaryDataInternal) {
			t.Fatal("changed bestiary not written")
		}
		reloaded, err := LoadSave(global.Two, dir, slot, global.PC)
		if err != nil {
			t.Fatal(err)
		}
		if rb, _ := reloaded.Bestiary(); rb == nil || *rb != *b {
			t.Fatal("reloaded bestiary differs")
		}
	})
}
