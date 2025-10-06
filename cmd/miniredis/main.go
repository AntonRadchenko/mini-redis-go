package main

import (
	"github.com/AntonRadchenko/mini-redis-go/internal/config"
	"github.com/AntonRadchenko/mini-redis-go/internal/logx"
	"github.com/AntonRadchenko/mini-redis-go/internal/server"
)

func main() {
    cfg := config.Load()
    s := server.NewServer(cfg.Addr)

    if err := s.Run(); err != nil {
        logx.Error("server stopped with error: %v", err)
    }
}
