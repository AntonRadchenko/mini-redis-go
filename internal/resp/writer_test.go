package resp

import (
	"bytes"
	"reflect"
	"testing"
)

func TestWriter_WriteSimple(t *testing.T) {
	// создаём буфер (куда Writer будет писать)
    var buf bytes.Buffer
    w := NewWriter(&buf)

	err := w.WriteSimple("HELLO")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "+HELLO\r\n"
	got := buf.String() // результат функции WriteSimple в виде строки
	if expected != got {
		t.Errorf("expected %v, got %v", expected, got)
	}
}

func TestWriter_WriteError(t *testing.T) {
	// создаём буфер (куда Writer будет писать)
    var buf bytes.Buffer
    w := NewWriter(&buf)

	err := w.WriteError("something went wrong")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "-something went wrong\r\n"
	got := buf.String() // результат функции WriteError в виде строки
	if !reflect.DeepEqual(expected, got) {
		t.Errorf("expected %v, got %v", expected, got)
	}
}

func TestWriter_WriteInteger(t *testing.T) {
	// создаём буфер (куда Writer будет писать)
    var buf bytes.Buffer
    w := NewWriter(&buf)

	err := w.WriteInteger(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := ":1\r\n"
	got := buf.String() // результат функции WriteInteger в виде строки
	if !reflect.DeepEqual(expected, got) {
		t.Errorf("expected %v, got %v", expected, got)
	}
}

func TestWriter_WriteBulk(t *testing.T) {
	// создаём буфер (куда Writer будет писать)
    var buf bytes.Buffer
    w := NewWriter(&buf)

	err := w.WriteBulk("world")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "$5\r\nworld\r\n"
	got := buf.String()
	if !reflect.DeepEqual(expected, got) {
		t.Errorf("expected %v, got %v", expected, got)
	}
}

func TestWriter_WriteArray(t *testing.T) {
	// создаём буфер (куда Writer будет писать)
    var buf bytes.Buffer
    w := NewWriter(&buf)

	err := w.WriteArray([]string{"anton", "artem", "margarita"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "*3\r\n$5\r\nanton\r\n$5\r\nartem\r\n$9\r\nmargarita\r\n"
	got := buf.String()
	if !reflect.DeepEqual(expected, got) {
		t.Errorf("expected %v, got %v", expected, got)
	}
}