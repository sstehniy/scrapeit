package scraper

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

type Cookie struct {
	Domain   string `json:"domain"`
	Expiry   int64  `json:"expiry"`
	HttpOnly bool   `json:"httpOnly"`
	Name     string `json:"name"`
	Path     string `json:"path"`
	SameSite string `json:"sameSite"`
	Secure   bool   `json:"secure"`
	Value    string `json:"value"`
}

type UserAgentWithCookies struct {
	Cookie    []Cookie
	UserAgent string
}

type CookieStore struct {
	mu      sync.Mutex
	Cookies map[string]UserAgentWithCookies `json:"cookies"`
}

var cookieFile = "cookies.json"

func LoadCookieStore() (*CookieStore, error) {
	store := &CookieStore{Cookies: make(map[string]UserAgentWithCookies)}
	if _, err := os.Stat(cookieFile); err == nil {
		data, err := os.ReadFile(cookieFile)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(data, store); err != nil {
			return nil, err
		}
	}
	return store, nil
}

func saveCookieStore(store *CookieStore) error {
	store.mu.Lock()
	defer store.mu.Unlock()
	data, err := json.Marshal(store)
	if err != nil {
		return err
	}
	return os.WriteFile(cookieFile, data, 0644)
}

func GetValidCookies(store *CookieStore, url string) (UserAgentWithCookies, bool) {
	store.mu.Lock()
	defer store.mu.Unlock()
	cookies, exists := store.Cookies[url]
	if !exists {
		return UserAgentWithCookies{}, false
	}
	for _, c := range cookies.Cookie {
		if c.Expiry > time.Now().Unix() {
			return cookies, true
		}
	}
	return UserAgentWithCookies{}, false
}

func SetCookies(store *CookieStore, url string, input UserAgentWithCookies) {
	store.mu.Lock()
	store.Cookies[url] = input
	store.mu.Unlock()
	saveCookieStore(store)
}
