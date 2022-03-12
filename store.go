package main

import "sync"

type URLStore struct {
	urls map[string]string
	mu   sync.RWMutex
}

func NewURLStore() *URLStore {
	return &URLStore{urls: make(map[string]string)}
}

func (s *URLStore) Get(key string) string { // get shortURL from LongURL
	s.mu.RLock() // 防止脏读
	defer s.mu.RUnlock()
	return s.urls[key]
}

func (s *URLStore) Set(key string, url string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, present := s.urls[key]
	if present {
		return false
	}

	s.urls[key] = url
	return true
}

func (s *URLStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.urls)
}

func (s *URLStore) Put(url string) string { // longURL -> get shortURL
	for {
		key := genKey(s.Count())
		if s.Set(key, url) {
			return key
		}
	}
	// shouldn't get shortURL
	return ""
}
