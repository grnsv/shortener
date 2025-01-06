package util

import (
	"crypto/md5"
	"encoding/base64"
)

const shortURLLength = 8

func GenerateShortURL(url []byte) string {
	h := md5.Sum(url)
	return base64.URLEncoding.EncodeToString(h[:])[:shortURLLength]
}
