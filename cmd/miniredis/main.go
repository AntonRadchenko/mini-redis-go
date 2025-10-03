package main

import (
    "log"

    "github.com/AntonRadchenko/mini-redis-go/internal/server"
)

func main() {
    s := server.NewServer(":6379")
    if err := s.Run(); err != nil {
        log.Fatal(err)
    }
}
