package store

import (
	"testing"
	"time"
)

func TestStore_TTL(t *testing.T) {
	s := NewStore()
	// 1) Set() - создаем ключ-значение Expire()
	// 2) Expire() - устанавливаем на ключ TTL (возвращает false если ключа нет)
	// 3) Получаем ключ до истечения времени
	// 4) Получаем ключ после истечения времени (если Get вернет false, то все в порядке)
	s.Set("name", "anton")
	ok := s.Expire("name", 1)

	// проверяем до истечения срока
	val, ok := s.Get("name")
	if !ok || val != "anton" {
		t.Fatalf("expected key to exist before TTL expires")
	}

	// ждем истечения
	time.Sleep(1500 * time.Millisecond)
	s.CleanExpiredKeys() // вручную чистим просроченные ключи

	// проверяем после истечения срока
	_, ok = s.Get("name")
	if ok {
		t.Errorf("expected key 'anton' to expire")
	}
}
