package server

import (
	"context"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/AntonRadchenko/mini-redis-go/internal/logx"
	"github.com/AntonRadchenko/mini-redis-go/internal/resp"
	"github.com/AntonRadchenko/mini-redis-go/internal/store"
)

// Структура Server - это место для таких зависимостей как адрес порта, логи, хранилище
type Server struct {
	addr  string // адрес порта
	store *store.Store
	r     *Router
	maxClients int // max число клиентов, которые могут подключиться одновременно
}

// Конструктор NewServer создает новый объект Server, то есть создает сервер для пользователя
func NewServer(a string) *Server {
	s := store.NewStore()
	s.StartTTLScanner(1 * time.Second) // запускаем фоновой сканер истёкших ключей
	r := New(s)                        // создаём роутер, связанный с этим хранилищем
	maxClients := 100 // задаем максимальное кол-во клиентов

	return &Server{
		addr:  a,
		store: s,
		r:     r,
		maxClients: maxClients,
	}
}

// метод Run - поднимает TCP-листенер и мы принимаем соединения
func (s *Server) Run(ctx context.Context) error {
	// запускаем прослушивание указаного адреса и порта
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}

	defer listener.Close() // <- вызовется при выходе из функции

	// показываем что сервер начал работу
	logx.Info("Server started on %s", s.addr) 

	sem := make(chan struct{}, s.maxClients) // семафор для ограничения клиентов
	var wg sync.WaitGroup // для ожидания завершения всех соединений (только потом сможем выйти)

	// бесконечный цикл для приема соединений
	for {
		// select нужен, чтобы завершить цикл применив Ctrl+C (ctx.Done())
		// — закрываем listener, ждём активные соединения с wg.Wait() и выходим
		select {
		case <-ctx.Done():
			// graceful shutdown
			logx.Info("Shutdown signal received, closing listener...")
			listener.Close() // <- вызывается вручную при Ctrl+C
			logx.Info("Listener closed, waiting for active clients...")
			wg.Wait() // дождёмся завершения активных соединений
			return nil		

		default:
			// Ставим дедлайн, чтобы Accept() не зависал навсегда.
			// Accept() — блокирующий вызов: пока никто не подключился (redis-cli), сервер "спит".
			// Дедлайн заставляет Accept() возвращать ошибку timeout каждые 500 мс,
			// чтобы можно было проверить, не нажали ли Ctrl+C (ctx.Done()).
			if tcpLn, ok := listener.(*net.TCPListener); ok {
				_ = tcpLn.SetDeadline(time.Now().Add(500 * time.Millisecond))
			}

			conn, err := listener.Accept() // ждём подключения клиента (может блокировать выполнение)
			if err != nil {
				// Если ошибка — это таймаут, просто проверяем контекст и продолжаем цикл
				if ne, ok := err.(net.Error); ok && ne.Timeout() {
					select {
					// Если во время ожидания клиентского подключения нажали Ctrl+C —
					// выходим из сервера (graceful shutdown)
					case <-ctx.Done():
						logx.Info("Shutdown signal received while waiting on Accept")
						wg.Wait() // дождёмся завершения активных соединений
						return nil
					default:
						// Если сигнала нет — продолжаем слушать новых клиентов
						continue
					}
				}

				// Если контекст уже отменён (например, listener закрыт) — выходим
				if ctx.Err() != nil {
					logx.Info("Listener stopped by context cancel")
					wg.Wait()
					return nil
				}

				// прочие ошибки Accept — логируем и продолжаем
				log.Printf("Accept error: %v", err)
				continue
			}

			// Параллельная обработка нового клиента.
			// sem — ограничивает количество клиентов (maxClients);
			// wg — ждёт, пока все активные соединения завершатся при shutdown.
			sem <- struct{}{}
			wg.Add(1)
			go func(c net.Conn) {
				defer wg.Done()
				defer func() { <-sem }()
				s.handleConn(c)
			}(conn)
		}
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
				logx.Error("Client disconnected")
			} else {
				logx.Info("Server started on %s", s.addr)
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
