package main

import (
	"fmt"
	"strings"

	"github.com/AntonRadchenko/mini-redis-go/internal/resp"
)

func main() {
	input := "*2\r\n$4\r\nPING\r\n$4\r\nTEST\r\n"

	r := resp.NewReader(strings.NewReader(input))
	result, err := r.ReadArray()
	if err != nil {
		panic(err)
	}

	fmt.Println(result) // [PING TEST]
}
