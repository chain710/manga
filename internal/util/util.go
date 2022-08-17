package util

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"github.com/jxskiss/base62"
	"hash/fnv"
	"os"
	"time"
)

func FileHash(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}

	in := struct {
		Path    string    `json:"path"`
		Size    int64     `json:"size"`
		ModTime time.Time `json:"modtime"`
	}{
		Path:    path,
		Size:    info.Size(),
		ModTime: info.ModTime(),
	}
	data, err := json.Marshal(in)
	if err != nil {
		panic(err)
	}
	h64 := fnv.New64()
	_, _ = h64.Write(data)
	return base62.EncodeToString(h64.Sum(nil)), nil
}

func ImageHash(data []byte) string {
	h := md5.New()
	_, _ = h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}
