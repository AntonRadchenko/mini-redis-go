package tests

// e2e-тесты проверяют работу сервера "снаружи":
// вместо ручного ввода команд через redis-cli,
// они автоматически подключаются к TCP-порту mini-redis,
// отправляют реальные RESP-запросы и сверяют ответы сервера с ожидаемыми.

// главное отличие:
// ручная проверка = ты вводишь команды сам,
// e2e-тест = программа делает то же самое автоматически.

// В e2e-тесте мы выступаем клиентом, а не самим сервером.
// Поэтому мы не трогаем внутренние функции — мы просто шлём команду в сокет.

// То есть мы не вызываем Set, Get, WriteSimple вручную,
// потому что сервер делает всё это сам —
// а мы просто проверяем, что его внешний ответ соответствует ожидаемому.

import (
	"bufio"
	"net"
	"strings"
	"testing"
	"time"
)

// sendResp — вспомогательная функция: подключается, отправляет RESP-команду и возвращает все строки ответа.
func sendResp(t *testing.T, cmd string) []string {
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

// Тест SET key value → ожидаем +OK
func TestSetGetDel(t *testing.T) {
	// 1. SET key value
	resp := sendResp(t, "*3\r\n$3\r\nSET\r\n$3\r\nfoo\r\n$3\r\nbar\r\n")
	if len(resp) == 0 || resp[0] != "+OK" {
		t.Fatalf("SET failed: got %v", resp)
	}

	// 2. GET key → "$3\r\nbar\r\n"
	conn, err := net.DialTimeout("tcp", "localhost:6379", time.Second)
	if err != nil {
		t.Fatalf("cannot connect: %v", err)
	}
	defer conn.Close()

	conn.Write([]byte("*2\r\n$3\r\nGET\r\n$3\r\nfoo\r\n"))
	reader := bufio.NewReader(conn)

	line1, _ := reader.ReadString('\n')
	line2, _ := reader.ReadString('\n')

	if strings.TrimSpace(line1) != "$3" || strings.TrimSpace(line2) != "bar" {
		t.Fatalf("GET failed: got [%q, %q]", line1, line2)
	}

	// 3. DEL key → ":1"
	resp = sendResp(t, "*2\r\n$3\r\nDEL\r\n$3\r\nfoo\r\n")
	if len(resp) == 0 || resp[0] != ":1" {
		t.Fatalf("DEL failed: got %v", resp)
	}

	// 4. GET после удаления → "$-1"
	resp = sendResp(t, "*2\r\n$3\r\nGET\r\n$3\r\nfoo\r\n")
	if len(resp) == 0 || resp[0] != "$-1" {
		t.Fatalf("expected nil after DEL, got %v", resp)
	}
}

// Проверяем DEL несуществующего ключа → ":0"
func TestDelNonexistentKey(t *testing.T) {
	resp := sendResp(t, "*2\r\n$3\r\nDEL\r\n$5\r\nnoKey\r\n")
	if len(resp) == 0 || resp[0] != ":0" {
		t.Fatalf("DEL nonexistent key failed: got %v", resp)
	}
}
