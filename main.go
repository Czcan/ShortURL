package main

import (
	"flag"
	"fmt"
	"net/http"
)

const AddForm = `
<form method="POST" action="/add">
URL: <input type="text" name="url">
<input type="submit" value="Add">
</form>
`

var store *URLStore

func main() {
	flag.Parse()
	store = NewURLStore(*dataFile)
	http.HandleFunc("/add", Add)
	http.HandleFunc("/", Redirect)
	http.ListenAndServe(*listenAddr, nil)
}

// 从表单读取长URL
// PUT存储 -> shortURL
// 发送 shortURL 给用户
func Add(w http.ResponseWriter, r *http.Request) {
	url := r.FormValue("url")
	if url == "" {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, AddForm)
		return
	}

	key := store.Put(url)
	fmt.Fprintf(w, "http://%s:%s", *hostName, key)
}

// 输入shortURL -> 重定向到 longURL
func Redirect(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Path[1:]
	url := store.Get(key)
	if url == "" {
		http.NotFound(w, r)
		return
	}

	http.Redirect(w, r, url, http.StatusFound)
}
