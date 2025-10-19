package server

import (
	"net/http"

	"github.com/Ko4etov/go-metrics/internal/server/router"
)

type Server struct {
	serverAddress  string
}

func New(serverAddress string) *Server {
	
	return &Server{
		serverAddress: serverAddress,
	}
}

func (s *Server) Run() {
	serverRouter := router.New()

	// Запускаем сервер
	err := http.ListenAndServe(s.serverAddress, serverRouter)
	if err != nil {
		panic(err)
	}
}