package logx

import "log"

// logx — это модуль для удобного и единообразного вывода логов,
// чтобы писать важные события (запуск сервера, ошибки, команды клиентов) в консоль в понятном формате,
// вместо беспорядочного использования log.Printf по всему проекту.

func Info(msg string, args ...any) {
	log.Printf("INFO: " + msg, args...)
}

func Error(msg string, args ...any) {
	log.Printf("ERROR: " + msg, args...)
}

// то есть просто добавили уровни логирования