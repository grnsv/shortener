package service

import (
	"crypto/md5"
	"encoding/base64"
	"sync"
)

const shortURLLength = 8

var (
	urlMap = make(map[string]string)
	mu     sync.RWMutex
)

func generateShortURL(url string) string {
	h := md5.Sum([]byte(url))
	return base64.URLEncoding.EncodeToString(h[:])[:shortURLLength]
}

func ShortenURL(url string) string {
	shortURL := generateShortURL(url)

	mu.Lock()
	urlMap[shortURL] = url
	mu.Unlock()

	return shortURL
}

func ExpandURL(shortURL string) (string, bool) {
	mu.RLock()
	url, exists := urlMap[shortURL]
	mu.RUnlock()

	return url, exists
}
