package store

import "time"

// метод Expire - задаёт время жизни ключа (в секундах).
// Возвращает true, если TTL успешно установлен, и false, если ключ не существует.
func (s *Store) Expire(key string, seconds int) bool {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	if _, ok := s.data[key]; !ok {
		return false
	}
	s.ttl[key] = time.Now().Add(time.Duration(seconds) * time.Second) // считаем момент истечения значения по ключу
	return true                                                       // показываем что ttl установлен успешно
}

// метод CleanExpiredKeys - удаляет значение по ключу если его время жизни истекло
// то есть проходит по всем s.ttl и чистит просроченные
func (s *Store) CleanExpiredKeys() {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	for key, expireTime := range s.ttl {
		if time.Now().After(expireTime) { // если настал момент истечения (если текущее время позже, чем истечение ключа)
			delete(s.data, key)
			delete(s.ttl, key)
		}
	}
}

// метод StartTTLScanner — запускает фоновую горутину,
// которая через равные интервалы времени вызывает CleanExpiredKeys()
// и удаляет истёкшие ключи из хранилища.
func (s *Store) StartTTLScanner(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval) // создаёт таймер, который через каждые interval вызывает очистку просроченных ключей.
		defer ticker.Stop()                // гарантируем остановку таймера при завершении горутины

		for range ticker.C { // ждём каждый "тик" таймера
			s.CleanExpiredKeys()
		}
	}()
}

// TTL сообщает, сколько секунд осталось до истечения срока жизни ключа.
// Возвращает:
// -2 если ключ не существует,
// -1 если ключ существует, но без TTL,
// N (в секундах), если TTL установлен и активен.
func (s *Store) TTL(key string) int {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	if _, ok := s.data[key]; !ok {
		return -2
	}

	expireTime, ttlOk := s.ttl[key]
	if !ttlOk {
		return -1
	}

	remainingTime := int(time.Until(expireTime).Seconds())
	if remainingTime < 0 { // если время уже ключа уже истекло (то есть также его уже нет в хранилище)
		return -2
	}

	return remainingTime // возвращаем оставшееся время
}