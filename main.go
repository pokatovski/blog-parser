package main

import (
	_ "github.com/heroku/x/hmetrics/onload"
	"github.com/pokatovski/blog-parser/internal/handler"
	"github.com/pokatovski/blog-parser/internal/service"
	"log"
	"net/http"
	"os"
)

type Server struct {
	httpServer *http.Server
}

func main() {

	port := os.Getenv("PORT")

	if port == "" {
		log.Println("$PORT is empty, set default to :8080")
		port = "8080"
	}
	services := service.NewService()
	handlers := handler.NewHandler(services)

	srv := new(Server)
	err := srv.Run(port, handlers.InitRoutes())
	if err != nil {
		log.Fatal(err)
	}
}

func (s *Server) Run(port string, handler http.Handler) error {
	s.httpServer = &http.Server{
		Addr:    ":" + port,
		Handler: handler,
	}

	return s.httpServer.ListenAndServe()
}
