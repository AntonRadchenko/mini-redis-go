package main

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/AntonRadchenko/mini-redis-go/internal/resp"
)

func main() {
	var buf bytes.Buffer
	w := resp.NewWriter(&buf)

	// Simple String
	buf.Reset()
	w.WriteSimple("PONG")
	fmt.Println("WriteSimple:", strconv.Quote(buf.String()))

	// Error
	buf.Reset()
	w.WriteError("ERR something went wrong")
	fmt.Println("WriteError:", strconv.Quote(buf.String()))

	// Integer
	buf.Reset()
	w.WriteInteger(42)
	fmt.Println("WriteInteger:", strconv.Quote(buf.String()))

	// Bulk String
	buf.Reset()
	w.WriteBulk([]byte("hello"))
	fmt.Println("WriteBulk (normal):", strconv.Quote(buf.String()))

	// Nil Bulk String
	buf.Reset()
	w.WriteBulk(nil)
	fmt.Println("WriteBulk (nil):", strconv.Quote(buf.String()))

	// Array
	buf.Reset()
	w.WriteArray([]string{"foo", "bar"})
	fmt.Println("WriteArray:", strconv.Quote(buf.String()))
}
