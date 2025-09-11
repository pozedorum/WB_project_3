// logger/logger.go
package logger

// globalLogger глобальный экземпляр логгера с поддержкой слоёв
var globalLogger *LayerLogger

func init() {
	// Инициализируем по умолчанию все слои включенными
	globalLogger = &LayerLogger{
		serverLogs:     true,
		serviceLogs:    true,
		repositoryLogs: true,
	}
}

// LayerLogger структура для хранения настроек логирования по слоям
type LayerLogger struct {
	serviceLogs    bool
	serverLogs     bool
	repositoryLogs bool
}

// Init инициализирует глобальный логгер с настройками слоёв
func Init(serviceLogs, serverLogs, repositoryLogs bool) {
	globalLogger = &LayerLogger{
		serviceLogs:    serviceLogs,
		serverLogs:     serverLogs,
		repositoryLogs: repositoryLogs,
	}
}

// LogService выполняет функцию логирования только если сервисный слой включен
func LogService(logFunc func()) {
	if globalLogger != nil && globalLogger.serviceLogs {
		logFunc()
	}
}

// LogServer выполняет функцию логирования только если серверный слой включен
func LogServer(logFunc func()) {
	if globalLogger != nil && globalLogger.serverLogs {
		logFunc()
	}
}

// LogRepository выполняет функцию логирования только если репозиторный слой включен
func LogRepository(logFunc func()) {
	if globalLogger != nil && globalLogger.repositoryLogs {
		logFunc()
	}
}
