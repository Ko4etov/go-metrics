package server

import (
	"net/http"

	"github.com/Ko4etov/go-metrics/internal/server/router"
	"github.com/Ko4etov/go-metrics/internal/server/service/logger"
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

	if err := logger.Initialize("info"); err != nil {
        panic(err)
    }

	// Запускаем сервер
	err := http.ListenAndServe(s.serverAddress, serverRouter)
	if err != nil {
		panic(err)
	}
}