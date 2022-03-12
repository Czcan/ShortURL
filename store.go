package main

import (
	"encoding/gob"
	"io"
	"log"
	"os"
	"sync"
)

type URLStore struct {
	urls map[string]string
	mu   sync.RWMutex
	file *os.File
}

type record struct {
	Key string
	URL string
}

func NewURLStore(filename string) *URLStore {
	s := &URLStore{urls: make(map[string]string)}
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		log.Fatal("Error opening URLStore:", err)
	}

	s.file = f
	if err := s.load(); err != nil {
		log.Fatal("Error loading URLStore:", err)
	}

	return s
}

func (s *URLStore) load() error {
	if _, err := s.file.Seek(0, 0); err != nil {
		return err
	}
	d := gob.NewDecoder(s.file)
	var err error
	for err == nil {
		var r record
		if err = d.Decode(&r); err == nil {
			s.Set(r.Key, r.URL)
		}
	}

	if err == io.EOF {
		return nil
	}
	return err
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
			if err := s.save(key, url); err != nil {
				log.Fatal("Error saving to URLStore: ", err)
			}
			return key
		}
	}
	// shouldn't get shortURL
	return ""
}

func (s *URLStore) save(key, url string) error {
	e := gob.NewEncoder(s.file)
	return e.Encode(record{key, url})
}
