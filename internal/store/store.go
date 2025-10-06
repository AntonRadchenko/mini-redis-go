package store

import (
	"sync"
	"time"
)

// структура Store — это потокобезопасное и высокопроизводительное in-memory key-value хранилище,
// обеспечивающее параллельный доступ к данным и пригодное для юнит-тестирования.
type Store struct {
	data map[string]string
	mtx  sync.RWMutex
	ttl map[string]time.Time // для каждого ключа храним время, через которое данные по этому ключу должны очиститься
}

// конструктор newStore() создает новый объект Store,
// создавая пустую мапу для хранения. (мютекс инициализируется по дефолту)
func NewStore() *Store {
	return &Store{
		data: make(map[string]string),
		ttl: make(map[string]time.Time),
	}
}

// метод Set - добавляет или обновляет значение по ключу в хранилище.
// Ничего не возвращает — успешность считается гарантированной.
// Ответ клиенту (+OK) формируется на уровне router.go (через WriteSimple).
func (s *Store) Set(key, value string) {
	s.mtx.Lock() // лочим для когкурентной записи 
	defer s.mtx.Unlock()
	s.data[key] = value
}

// мтеод Get - возвращает значение по ключу и флаг наличия.
// Если ключ найден — router.go отправит его клиенту через WriteBulk.
// Если нет — клиенту вернётся nil.
func (s *Store) Get(key string) (string, bool) {
	s.mtx.RLock() // лочим для конкурентного чтения (могут читать параллельно)
	defer s.mtx.RUnlock()
	val, ok := s.data[key]
	return val, ok
}

// метод Del - удаляет из хранилища один или несколько ключей.
// Возвращает количество реально удалённых элементов.
// router.go оборачивает это число в integer-ответ (WriteInteger).
func (s *Store) Del(keys ...string) int {
	count := 0
	s.mtx.Lock()
	defer s.mtx.Unlock()
	for _, key := range keys {
		_, ok := s.data[key]
		if ok {
			delete(s.data, key) // удаляем ключ если он есть
			count++
		}
	}
	return count
}
