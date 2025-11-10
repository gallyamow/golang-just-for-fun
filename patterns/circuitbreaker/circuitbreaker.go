package circuitbreaker

import (
	"sync"
	"time"
)

// CircuitBreaker
//
// Для чего:
// Circuit Breaker мониторит вызовы к внешнему сервису и при обнаружении слишком большого количества неудачных попыток
// временно "отключает" вызов, предотвращая тем самым падение всей системы.
//
// Требования:
// TODO: изучить
type CircuitBreaker struct {
	failureThreshold int
	resetTimeout     time.Duration
	state            CircuitState
	failureCount     int
	lastFailureTime  time.Time
	mu               sync.RWMutex
}

type CircuitState int

const (
	CircuitClosed CircuitState = iota
	CircuitOpen
	CircuitHalfOpen
)
