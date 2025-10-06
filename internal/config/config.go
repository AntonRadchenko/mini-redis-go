package config

import "time"

// структура Config — это набор всех настроек приложения (например, адрес сервера, уровень логов и т.д.),
// которые загружаются при запуске, чтобы управлять поведением программы без изменения кода.
type Config struct {
	Addr string // адрес, на котором слушает сервер
	LogLevel string // уровень логирования
	ReadTimeout time.Duration // таймаут на чтение запросов
	WriteTimeout time.Duration // таймаут на запись ответов
}

// метод Load — конструктор, который возвращает структуру Config  
// с дефолтными значениями основных параметров для запуска сервера.
func Load() *Config {
	cfg := &Config{
		Addr: ":6379", 
		LogLevel: "info",
		ReadTimeout: 5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
	return cfg
}

