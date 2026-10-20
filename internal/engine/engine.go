package engine

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
)

//go:embed payload/livekit.exe.gz
var payload embed.FS

func valid(path string) bool {
	f, e := os.Open(path)
	if e != nil {
		return false
	}
	defer f.Close()
	h := sha256.New()
	if _, e = io.Copy(h, f); e != nil {
		return false
	}
	return hex.EncodeToString(h.Sum(nil)) == executableSHA256
}
func Prepare(data string) (string, error) {
	dir := filepath.Join(data, "engine")
	if e := os.MkdirAll(dir, 0700); e != nil {
		return "", e
	}
	target := filepath.Join(dir, "livekit-1.13.10.exe")
	if valid(target) {
		return target, nil
	}
	compressed, e := payload.ReadFile("payload/livekit.exe.gz")
	if e != nil {
		return "", e
	}
	reader, e := gzip.NewReader(bytes.NewReader(compressed))
	if e != nil {
		return "", e
	}
	defer reader.Close()
	temp, e := os.CreateTemp(dir, "engine-*.tmp")
	if e != nil {
		return "", e
	}
	name := temp.Name()
	defer os.Remove(name)
	_, e = io.Copy(temp, io.LimitReader(reader, 150<<20))
	if e == nil {
		e = temp.Sync()
	}
	closeError := temp.Close()
	if e != nil {
		return "", e
	}
	if closeError != nil {
		return "", closeError
	}
	if !valid(name) {
		return "", errors.New("the bundled meeting engine failed its integrity check")
	}
	// Windows will refuse replacement of a running binary. Never stop another instance.
	if _, e = os.Stat(target); e == nil {
		if e = os.Remove(target); e != nil {
			return "", e
		}
	}
	if e = os.Rename(name, target); e != nil {
		return "", e
	}
	return target, nil
}
