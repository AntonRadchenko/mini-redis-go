package resp

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Структура Reader — это обёртка над bufio.Reader,
// предназначенная для чтения и парсинга RESP-запросов из источника байтов (io.Reader)
type Reader struct {
	r *bufio.Reader
}

// Конструктор NewReader создаёт новый объект Reader,
// оборачивая переданный io.Reader в bufio.Reader для удобного построчного чтения.
func NewReader(rd io.Reader) *Reader {
	return &Reader{r: bufio.NewReader(rd)}
}

// Метод ReadArray() читает RESP-массив из входящего потока (по типу *1\r\n$4\r\nPING\r\n)
// и преобразует его в срез строк (например ["SET", "key", "value"])
// для дальнейшей обработки уже на уровне Сервера
func (r *Reader) ReadArray() ([]string, error) {
	// читаем первый байт (проверяем что массив действительно начинается с '*')
	bt, err := r.r.ReadByte() // обращаемся к структуре Reader и потом уже к его полю, поэтому r.r. двойной
	if err != nil {
		return nil, err
	}

	if bt != '*' {
		return nil, fmt.Errorf("expected '*', got %q", bt)
	}

	// читаем кол-во элементов массива (по кол-ву '\n')
	line, err := r.r.ReadString('\n')
	if err != nil {
		return nil, err
	}

	// ищем число-количество элементов n будущего массива
	// (убираем лишние символы и превращаем в число)
	n, err := strconv.Atoi(strings.TrimSpace(line))
	if err != nil {
		return nil, err
	}
	// проверка на отрицательное количество элементов массива
	if n < 0 {
		return nil, fmt.Errorf("invalid array length %d", n)
	}

	// читаем n элементов
	result := make([]string, 0, n)
	for i := 0; i < n; i++ {
		// также ищем кол-во символов строки, которую поместим в result
		bt, err := r.r.ReadByte() // автоматически читается со следующих позиций, а не сначала
		if err != nil {
			return nil, err
		}
		if bt != '$' {
			return nil, fmt.Errorf("expected '$', got %q", bt)
		}

		line, err := r.r.ReadString('\n')
		if err != nil {
			return nil, err
		}

		// кол-во элементов строки i массива result
		length, err := strconv.Atoi(strings.TrimSpace(line))
		if err != nil {
			return nil, err
		}
		// проверка на отрицательную длину строки
		if length < 0 {
			return nil, fmt.Errorf("invalid bulk length %d", length)
		}

		// читаем из r.r кол-во (length) элементов, кладем в buf и превращаем в строку
		buf := make([]byte, length)
		if _, err := io.ReadFull(r.r, buf); err != nil {
			return nil, err
		}

		str := string(buf)
		// кладем строку в массив
		result = append(result, str)

		// пропустим символы после строки (\r\n)
		if _, err := r.r.Discard(2); err != nil {
			return nil, err
		}
	}
	return result, nil
}
