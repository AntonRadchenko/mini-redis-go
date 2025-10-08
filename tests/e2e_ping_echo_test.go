package tests

import (
	"bufio"
	"net"
	"strings"
	"testing"
	"time"
)

// sendCommand подключается к mini-redis по TCP, отправляет RESP-команду и возвращает первую строку ответа.
func sendCommand(t *testing.T, cmd string) string {
	conn, err := net.DialTimeout("tcp", "localhost:6379", time.Second)
	if err != nil {
		t.Fatalf("cannot connect to mini-redis: %v", err)
	}
	defer conn.Close()

	_, err = conn.Write([]byte(cmd))
	if err != nil {
		t.Fatalf("failed to send command: %v", err)
	}

	reader := bufio.NewReader(conn)
	resp, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}
	return strings.TrimSpace(resp)
}

// Проверяем команду PING → ожидаем +PONG
func TestPing(t *testing.T) {
	resp := sendCommand(t, "*1\r\n$4\r\nPING\r\n")
	if resp != "+PONG" {
		t.Fatalf("unexpected response: got %q, want %q", resp, "+PONG")
	}
}

// Проверяем команду ECHO <msg> → ожидаем bulk string с сообщением
func TestEcho(t *testing.T) {
	conn, err := net.DialTimeout("tcp", "localhost:6379", time.Second)
	if err != nil {
		t.Fatalf("cannot connect to mini-redis: %v", err)
	}
	defer conn.Close()

	// Отправляем RESP-команду: ECHO "HELLO"
	cmd := "*2\r\n$4\r\nECHO\r\n$5\r\nHELLO\r\n"
	if _, err := conn.Write([]byte(cmd)); err != nil {
		t.Fatalf("failed to send command: %v", err)
	}

	reader := bufio.NewReader(conn)

	// Bulk-ответ состоит из двух строк: "$5" и "HELLO"
	line1, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("failed to read first line: %v", err)
	}
	line2, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("failed to read second line: %v", err)
	}

	if strings.TrimSpace(line1) != "$5" || strings.TrimSpace(line2) != "HELLO" {
		t.Fatalf("unexpected echo response: got [%q, %q]", line1, line2)
	}
}

// Проверяем ECHO без аргументов → ожидаем ошибку (-ERR ...)
func TestEchoWrongArgs(t *testing.T) {
	resp := sendCommand(t, "*1\r\n$4\r\nECHO\r\n")
	if !strings.HasPrefix(resp, "-ERR") {
		t.Fatalf("expected error, got %q", resp)
	}
}
