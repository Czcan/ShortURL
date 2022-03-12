package main

import (
	"encoding/json"
	"flag"
	"io"
	"log"
	"os"
	"sync"
)

const saveQueueLength = 1000

var (
	listenAddr = flag.String("http", ":8099", "http listen address")
	dataFile   = flag.String("file", "store.json", "data store file name")
	hostName   = flag.String("host", "localhost:8099", "host name and port")
)

type URLStore struct {
	urls map[string]string
	mu   sync.RWMutex
	save chan record
}

type record struct {
	Key string
	URL string
}

func NewURLStore(filename string) *URLStore {
	s := &URLStore{
		urls: make(map[string]string),
		save: make(chan record, saveQueueLength),
	}

	if err := s.load(filename); err != nil {
		log.Println("Error loading URLStore:", err)
	}
	go s.saveLoop(filename)
	return s
}

func (s *URLStore) load(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		log.Println("Error opening URLStore: ", err)
	}
	defer f.Close()

	d := json.NewDecoder(f)
	for err == nil {
		var r record
		if err = d.Decode(&r); err == nil {
			s.Set(r.Key, r.URL)
		}
	}

	if err == io.EOF {
		return nil
	}
	log.Println("Error decoding URLStore: ", err)
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
			s.save <- record{key, url}
			return key
		}
	}
	// shouldn't get shortURL
	panic("shouldn't get here")
}

func (s *URLStore) saveLoop(filename string) {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		log.Fatal("Error opening URLStore: ", err)
	}
	defer f.Close()

	e := json.NewEncoder(f)
	for {
		r := <-s.save
		if err = e.Encode(&r); err != nil {
			log.Println("Error saving to URLStore: ", err)
		}
	}
}
