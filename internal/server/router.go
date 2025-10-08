package server

import (
	"strconv"
	"strings"

	"github.com/AntonRadchenko/mini-redis-go/internal/store"
)

// Структура Reply — универсальная обёртка для ответа: хранит тип и значение.
// Роутер формирует Reply, а сервер по полю Type выбирает нужный метод записи ответа клиенту.
type Reply struct {
	Type  string      // тип ответа ("simple" | "bulk" | "integer" | "array" | "error")
	Value interface{} // значение ответа (строка, число, массив и т.п.).
}

// структура Router — это обработчик клиентских команд.
// Содержит ссылку на хранилище и решает, какую операцию выполнить (SET, GET, DEL и т.д.).
type Router struct {
	store *store.Store
}

// конструктор New создаёт новый объект Router
// и связывает его с конкретным экземпляром хранилища Store.
func New(store *store.Store) *Router {
	return &Router{store: store}
}

// метод - Handle получает распарсенные аргументы команды,
// определяет, что выполнить, и формирует ответ (Reply) для клиента.
func (r *Router) Handle(args []string) Reply {
	if len(args) == 0 {
		return Reply{"error", "ERR empty command"}
	}

	cmd := strings.ToUpper(args[0]) // приводим строку от клиента к верхнему регистру

	// проверяем введенные данные и сохраняем в структуру тип и значение
	switch cmd {
	case "PING":
		return Reply{Type: "simple", Value: "PONG"}

	case "ECHO":
		if len(args) < 2 {
			return Reply{Type: "error", Value: "ERR wrong num of arguments for 'echo'"}
		}
		// чтобы могли вывести несколько слов, объединяем аргументы после ECHO в одну строку
		msg := strings.Join(args[1:], " ")
		return Reply{Type: "bulk", Value: msg}

	// следующие проверки команд, использующих store/
	case "SET":
		if len(args) != 3 {
			return Reply{Type: "error", Value: "ERR wrong number of arguments for 'set' command"}
		}
		r.store.Set(args[1], args[2])
		return Reply{Type: "simple", Value: "OK"} // просто говорим +OK, типо все записалось хорошо

	case "GET":
		if len(args) != 2 {
			return Reply{Type: "error", Value: "ERR wrong number of arguments for 'get' command"}
		}
		val, ok := r.store.Get(args[1])
		if !ok {
			return Reply{Type: "bulk", Value: nil}
		}
		return Reply{Type: "bulk", Value: val}

	case "DEL":
		if len(args) < 2 {
			return Reply{Type: "error", Value: "ERR wrong number of arguments for 'del' command"}
		}
		count := r.store.Del(args[1:]...)
		return Reply{Type: "integer", Value: count}

	case "MGET":
		if len(args) < 2 {
			return Reply{Type: "error", Value: "ERR wrong number of arguments for 'mget' command"}
		}
		var results []string
		for _, key := range args[1:] {
			val, ok := r.store.Get(key)
			if !ok {
				results = append(results, "") // nil → пустая строка
			} else {
				results = append(results, val)
			}
		}
		return Reply{Type: "array", Value: results}

	case "EXPIRE":
		if len(args) != 3 {
			return Reply{Type: "error", Value: "ERR wrong number of arguments for 'expire' command"}
		}
		seconds, err := strconv.Atoi(args[2]) // превращаем длительность из строкового типа в integer
		if err != nil {
			return Reply{Type: "error", Value: "ERR value is not an integer or out of range"}
		}
		ok := r.store.Expire(args[1], seconds)
		if ok {
			return Reply{Type: "integer", Value: 1}
		}
		return Reply{Type: "integer", Value: 0}

	case "TTL":
		if len(args) != 2 {
			return Reply{Type: "error", Value: "ERR wrong number of arguments for 'ttl' command"}
		}
		ttl := r.store.TTL(args[1])
		return Reply{Type: "integer", Value: ttl}

	default:
		return Reply{"error", "ERR unknown command '" + cmd + "'"}
	}
}
