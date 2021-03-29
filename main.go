package main

import (
	"log"
	"net/http"
	"os"

	_ "github.com/heroku/x/hmetrics/onload"
	"github.com/pokatovski/blog-parser/internal/handler"
)

func main() {

	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}
	fs := http.FileServer(http.Dir("./web/static"))
	http.Handle("/web/static/", http.StripPrefix("/web/static/", fs))
	http.HandleFunc("/", handler.Index)
	http.HandleFunc("/parse", handler.Parse)

	http.ListenAndServe(":"+port, nil)
}
