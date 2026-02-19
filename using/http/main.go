package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	// HandleFunc - глобальный, очень простой роутинг из стандартной библиотеки.
	// Минусы:
	// - глобальное состояние
	// - нет параметров /user/{id} (у http.ServeMux до go 1.22 тоже нет)
	// - слабая поддержка http-методов
	// - middleware - сложно
	//
	// Альтернатива: встроенный mux либо third-party gorilla/mux - вроде deprecated почти, либо  go-chi/chi - легковеснее (1000 LOC).
	// Либо уже фреймворк: gin, echo
	//
	// Плюсы go-chi/chi:
	// - параметры (спорно, теперь в стандартном есть)
	// - middlewares и готовые версии
	// - без reflection (как и другие), маленький, активно развивается
	// - RouteContext =  каждый запрос получает контекст маршрута — структуру, в которой хранится: шаблон, параметры, middlewares.
	//   Стандартный просто выбирает handler по пути путь.
	//   Это позволяет проще интегрировать prometheus (на такие labels /user/{id}, а не парсить url и выяснять label).
	//   (избегаем high cardinality labels - вместо уникальных URI типа /orders/123, используйте шаблоны /orders/{id}, отбрасываем все лишнее из url)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// @idiomatic: using Fprintf to write string to io.Writer
		_, err := fmt.Fprintf(w, time.Now().Format(time.RFC3339))
		if err != nil {
			log.Printf("Error: %v", err)
		}
	})

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
}
