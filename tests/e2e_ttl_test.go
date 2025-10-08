package tests

import (
	"bufio"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"
)

// helper: подключается к серверу и возвращает все строки ответа
func sendTTLCommand(t *testing.T, cmd string) []string {
	conn, err := net.DialTimeout("tcp", "localhost:6379", time.Second)
	if err != nil {
		t.Fatalf("cannot connect to mini-redis: %v", err)
	}
	defer conn.Close()

	if _, err := conn.Write([]byte(cmd)); err != nil {
		t.Fatalf("failed to send command: %v", err)
	}

	reader := bufio.NewReader(conn)
	var lines []string
	
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		lines = append(lines, strings.TrimSpace(line))
		if reader.Buffered() == 0 { // если в буфере больше нет данных — выходим
			break
		}
	}
	return lines
}

// Проверяем EXPIRE + TTL + автоматическое истечение ключа
func TestExpireAndTTL(t *testing.T) {
	// 1. SET foo bar → +OK
	resp := sendTTLCommand(t, "*3\r\n$3\r\nSET\r\n$3\r\nfoo\r\n$3\r\nbar\r\n")
	if len(resp) == 0 || resp[0] != "+OK" {
		t.Fatalf("SET failed: got %v", resp)
	}

	// 2. EXPIRE foo 2 → :1
	resp = sendTTLCommand(t, "*3\r\n$6\r\nEXPIRE\r\n$3\r\nfoo\r\n$1\r\n2\r\n")
	if len(resp) == 0 || resp[0] != ":1" {
		t.Fatalf("EXPIRE failed: got %v", resp)
	}

	// 3. TTL foo → положительное число
	resp = sendTTLCommand(t, "*2\r\n$3\r\nTTL\r\n$3\r\nfoo\r\n")
	if len(resp) == 0 || !strings.HasPrefix(resp[0], ":") {
		t.Fatalf("TTL failed: got %v", resp)
	}
	ttlVal, err := strconv.Atoi(strings.TrimPrefix(resp[0], ":"))
	if err != nil || ttlVal <= 0 {
		t.Fatalf("TTL should be positive, got %v", resp)
	}

	// 4. ждём 2.5 секунды (ключ должен истечь)
	time.Sleep(2500 * time.Millisecond)

	// 5. GET foo → $-1
	resp = sendTTLCommand(t, "*2\r\n$3\r\nGET\r\n$3\r\nfoo\r\n")
	if len(resp) == 0 || resp[0] != "$-1" {
		t.Fatalf("expected expired key, got %v", resp)
	}

	// 6. TTL foo → -2 (ключ не существует)
	resp = sendTTLCommand(t, "*2\r\n$3\r\nTTL\r\n$3\r\nfoo\r\n")
	if len(resp) == 0 || resp[0] != ":-2" {
		t.Fatalf("expected TTL=-2 after expire, got %v", resp)
	}
}

// Проверяем TTL для несуществующего ключа
func TestTTLNonexistentKey(t *testing.T) {
	resp := sendTTLCommand(t, "*2\r\n$3\r\nTTL\r\n$5\r\nnoKey\r\n")
	if len(resp) == 0 || resp[0] != ":-2" {
		t.Fatalf("expected TTL=-2 for nonexistent key, got %v", resp)
	}
}
