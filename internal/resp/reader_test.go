package resp

import (
	"strings"
	"testing"
)

func TestReader_ReadArray(t *testing.T) {
	// входные данные в RESP: массив из 2 строк ["PING", "TEST"]
	input := "*2\r\n$4\r\nPING\r\n$4\r\nTEST\r\n"

	// создаём Reader на основе строки (имитация TCP потока)
	r := NewReader(strings.NewReader(input))

	// вызываем метод
	result, err := r.ReadArray()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// ожидаем результат
	expected := []string{"PING", "TEST"}
	if len(result) != len(expected) {
		t.Fatalf("expected len=%d, got %d", len(expected), len(result))
	}
	for i := range expected {
		if result[i] != expected[i] {
			t.Errorf("expected %q, got %q", expected[i], result[i])
		}
	}
}
