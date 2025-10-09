package store

import (
	"testing"
)

// проверяет сразу Set, Get и Del
// Set - должен корректно их записать
// Get - должен корректно их вернуть
func TestStore_SetGetDel(t *testing.T) {
	// создаём новое хранилище
	s := NewStore()
	// проверяем SET
	s.Set("name", "anton")

	// проверяем GET
	val, ok := s.Get("name")
	// проверяем что ключ доступен
	if !ok {
		t.Fatalf("expected key to exist")
	}
	if val != "anton" {
		t.Errorf("expected 'anton', got '%s'", val)
	}

	// проверяем DEL
	deleted := s.Del("name")
	if deleted != 1 {
		t.Errorf("expected 1 deleted, got %d", deleted)
	}

	// проверяем что на месте удаленного значения пустота (nil)
	_, ok = s.Get("name")
	if ok {
		t.Errorf("expected key 'name' to be deleted, but still exists")
	}
}
