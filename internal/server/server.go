package server

import (
	"io"
	"log"
	"net"
	"time"

	"github.com/AntonRadchenko/mini-redis-go/internal/resp"
	"github.com/AntonRadchenko/mini-redis-go/internal/store"
)

// Структура Server - это место для таких зависимостей как адрес порта, логи, хранилище
type Server struct {
	addr string // адрес порта
	store *store.Store
	r *Router
}

// Конструктор NewServer создает новый объект Server, то есть создает сервер для пользователя
func NewServer(a string) *Server {
	s := store.NewStore()
	s.StartTTLScanner(1 * time.Second) // запускаем фоновой сканер истёкших ключей
	r := New(s) // создаём роутер, связанный с этим хранилищем

	return &Server{
		addr: a,
		store: s,
		r: r,
	}
}

// метод Run - поднимает TCP-листенер и мы принимаем соединения
func (s *Server) Run() error {
	// запускаем прослушивание указаного адреса и порта
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	defer listener.Close()

	log.Printf("Server started on %s", s.addr)

	// бесконечный цикл для приема соединений
	for {
		// принимаем новое входящее соединение
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Accept error: %v", err)
			continue // продолжаем слушать следующие соединения
		}
		// обработка соединения (в горутине)
		go s.handleConn(conn)
	}
}

// метод handleConn - обрабатывает соединение
func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()

	// у нас открытое TCP-соединение с клиентом;
	// оборачиваем наш conn в reader и writer, которые мы реализовали в /resp;

	// в интерфейсе net.Conn есть методы чтения и записи, поэтому он реализует методы
	// интерфейсов io.Reader и io.Writer => conn можно передавать в аргументы NewReader и NewWriter

	rd := resp.NewReader(conn) // оборачиваем conn в Reader
	wr := resp.NewWriter(conn) // оборачиваем conn в Writer

	// цикл общения с клиентом
	for {
		args, err := rd.ReadArray() // читаем данные от клиента
		if err != nil {
			if err == io.EOF { // если достигнут конец потока данных
				log.Println("Client Disconnected")
			} else {
				log.Printf("Read error: %v", err)
			}
			return
		}

		// обрабатываем в router данные и получаем в структуре тип команды и само значение которое нужно отдать клиенту (write)
		reply := s.r.Handle(args)
		
		// в зависимости от типа команды, выбираем как записать ответ клиенту
		switch reply.Type {
		case "simple":
			err := wr.WriteSimple(reply.Value.(string)) // достаем из интерфейса Value определенный тип
			if err != nil {
				log.Printf("Write error: %v", err)
			}
		
		case "bulk":
			if reply.Value == nil {
				_ = wr.WriteBulk("") // передадим "", чтобы сработала ветка "$-1\r\n"
				break
			}
			err := wr.WriteBulk(reply.Value.(string))
			if err != nil {
				log.Printf("Write error: %v", err)
			}

		case "integer":
			err := wr.WriteInteger(reply.Value.(int))
			if err != nil {
				log.Printf("Write error: %v", err)
			}
		
		case "array":
			values, ok := reply.Value.([]string)
			if !ok {
				_ = wr.WriteError("ERR internal: array value type mismatch")
				return
			}
			_ = wr.WriteArray(values)

		case "error":
			err := wr.WriteError(reply.Value.(string))
			if err != nil {
				log.Printf("Write error: %v", err)
			}
		
		default:
			// на всякий случай
			err := wr.WriteError("ERR internal: unsupported reply type")
			if err != nil {
				log.Printf("Write Error: %v", err)
			}
			return 
		}
	}
}

// Потом, когда появятся другие части проекта:
// logx заменит log.Printf → будет единый логгер, аккуратнее.