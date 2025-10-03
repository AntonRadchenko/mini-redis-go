package server

import (
	"strings"
)

// Структура Reply — универсальная обёртка для ответа: хранит тип и значение.
// Роутер формирует Reply, а сервер по полю Type выбирает нужный метод записи ответа клиенту.
type Reply struct {
	Type string // тип ответа ("simple" | "bulk" | "integer" | "array" | "error")
	Value interface{} // значение ответа (строка, число, массив и т.п.).
}

// получает распарсенные аргументы команды и решает, что ответить.
// возвращает тип ответа клиенту, и содержимое ответа (в структуре)
func Handle(args []string) Reply {
	if len(args) == 0 {
		return Reply{"error", "ERR empty command"}
	}

	cmd := strings.ToUpper(args[0]) // приводим строку от клиента к верхнему регистру

	// проверяем введенные данные и сохраняем в структуру тип и значение
	switch cmd{
	case "PING":
		return Reply{Type: "simple", Value: "PONG"}
	
	case "ECHO":
		if len(args) < 2 {
			return Reply{Type: "error", Value: "ERR wrong num of arguments for 'echo'"}
		}
		return Reply{Type: "bulk", Value: args[1]} 

	// остальные команды добавим, когда будет store/

	default:
		return Reply{"error", "ERR unknown command '" + cmd + "'"}
	}
}