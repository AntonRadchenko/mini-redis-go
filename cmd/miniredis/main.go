package main

import (
    "context"
	"os/signal"
	"syscall"
	"github.com/AntonRadchenko/mini-redis-go/internal/config"
	"github.com/AntonRadchenko/mini-redis-go/internal/logx"
	"github.com/AntonRadchenko/mini-redis-go/internal/server"
)

// main — точка входа в приложение.
// Здесь мы загружаем конфигурацию, создаём сервер и запускаем его.
// Сервер внутри сам обрабатывает SIGINT/SIGTERM и завершает работу корректно.
func main() {

    ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer stop()


	cfg := config.Load()
	s := server.NewServer(cfg.Addr)

	if err := s.Run(ctx); err != nil { 
		logx.Error("server stopped with error: %v", err)
	}
}
