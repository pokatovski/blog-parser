package main

import (
	"github.com/pokatovski/blog-parser/internal/handler"
	"log"
	"net/http"
)

func main() {
	fs := http.FileServer(http.Dir("./web/static"))
	http.Handle("/web/static/", http.StripPrefix("/web/static/", fs))
	http.HandleFunc("/", handler.Index)
	http.HandleFunc("/parse", handler.Parse)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
