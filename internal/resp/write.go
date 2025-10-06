package resp

import (
	"bufio"
	"fmt"
	"io"
)

// структура Writer это обертка над bufio.Writer,
// предназначенная для записи ответов сервера клиенту в формате RESP
type Writer struct {
	w *bufio.Writer
}

// Конструктор NewWriter создаёт новый объект Writer,
// оборачивая переданный io.Writer в bufio.Writer для удобной буферизованной записи RESP-ответов.
func NewWriter(wr io.Writer) *Writer {
	return &Writer{bufio.NewWriter(wr)}
}


// методы для записи различных данных в RESP формате
func (w *Writer) WriteSimple(s string) error {
	_, err := w.w.WriteString("+" + s + "\r\n") 
	if err != nil {
		return err
	}
	return w.w.Flush() // Flush отправляет все записанные данные клиенту.
}

func (w *Writer) WriteError(s string) error {
	_, err := w.w.WriteString("-" + s + "\r\n")
	if err != nil {
		return err
	}
	return w.w.Flush()
}

func (w *Writer) WriteInteger(i int) error {
	_, err := fmt.Fprintf(w.w, ":%d\r\n", i) // Fprintf не просто форматирует строку, а сразу пишет ее в bufio.Writer
	if err != nil {
		return err
	}
	return w.w.Flush()
}

func (w *Writer) WriteBulk(s string) error {
	if s == "" { // если nil значение
		_, err := w.w.WriteString("$-1\r\n")
		if err != nil {
			return err
		}
		return w.w.Flush()
	}
	// обычная строка (длину строки, затем саму строку)
	_, err := fmt.Fprintf(w.w, "$%d\r\n%s\r\n", len(s), s)
	if err != nil {
		return err
	}
	return w.w.Flush()
}

func (w *Writer) WriteArray(values []string) error {
	// заголовок
	_, err := fmt.Fprintf(w.w, "*%d\r\n", len(values))
	if err != nil {
		return err
	}
	// затем сами строки (по принципу WriteBulk)
	for _, v := range values {
		err := w.WriteBulk(v)
		if err != nil {
			return err
		}
	}
	
	return nil
}