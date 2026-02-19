package main

import (
	"context"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"log"
	"net/http"
	"time"
)

var metricRequestsTotal, metricRequestsLatency = initPrometheus()

var port = 8080

// Prometheus отвечает за метрики и агрегацию, OpenTelemetry — за трассировку и контекст выполнения. Вместе они дают полноценную observability.
//
// Prometheus:
// - metrics backend
// - метрики покажут что плохо, но не покажут где именно
//
// OpenTelemetry (OTel) — это открытый стандарт и набор инструментов (API, SDK, агенты) для сбора, генерации и
// экспорта данных телеметрии (трассировок, метрик и логов) из приложений.
// В трассировке есть понятие span, это аналог одного лога, в консоль. У span есть:
//   - Название, обычно это название метода который выполнялся
//   - Название сервиса, в котором был сгенерирован span
//   - Собственный уникальный ID
//   - Какая-то мета информация в виде key/value, которую залогировали в него. Например, параметры метода или закончился метод ошибкой или нет
//   - Время начала и конца выполнения этого span
//   - ID родительского span
//
// Как это всё работает ВМЕСТЕ:
// Client -> HTTP request -> [otelhttp middleware] (создаёт span, добавляет trace-id в context) -> Business logic (использует span) -> Response.
//
// Проблема "high cardinality labels":
// Prometheus хранит тайм-серии, Одна уникальная комбинация labels = одна тайм-серия. Используя стандартный mux, придется готовить label. В случа chi - все готово.
//
// Рекомендации к настройке мониторинга:
// 1) Error Budget Policy
// Если slo = 99.9%, => Error Budget = 0.1%: если ошибок столько еще не было, можно релизить, иначе стабилизация.
// 2) Burn Rate Alerting = надо алертить не факт прожига error budget, а скорости его прожига. Это касается и других конечных величин.
// Выявляют с помощью 2 окон:
// - Быстрое (5–15 мин) - "Прямо сейчас всё плохо?"  = резкие аварии
// - Медленное (1–6 ч) - "Проблема устойчивая?" = деградации
// Значения показателей подбирают.
// Алерт срабатывает только если оба “да”.
// 5. Golden Signals
// 4 метрики, которые реально важны
// - Latency — как долго ждут
// - Traffic — сколько запросов
// - Errors — сколько неудач
// - Saturation — насколько упёрлись в лимиты
func main() {
	ctx := context.Background()
	shutdownTracer := initJaeger()
	defer func() {
		err := shutdownTracer(ctx)
		log.Printf("failed close: %v", err)
	}()

	router := chi.NewRouter()
	router.Use(prometheusMiddleware)

	// Обычный route
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprintf(w, time.Now().Format(time.RFC3339))
		if err != nil {
			log.Printf("Error: %v", err)
		}
	})

	// Route с трассировкой:
	router.Get("/traced", func(w http.ResponseWriter, r *http.Request) {
		// namespace
		tracer := otel.Tracer("tracer-name")

		// Используем контекст запроса, создаем дочерний от HTTP span.
		// Этим span будет помечен запрос.
		_, span := tracer.Start(r.Context(), "span-name")
		defer span.End()

		_, err := fmt.Fprintf(w, time.Now().Format(time.RFC3339))
		if err != nil {
			log.Printf("Error: %v", err)
		}
	})

	// promhttp - реализует HTTP handler, собирает все зарегистрированные метрики, сериализует их
	// в prometheus exposition format и отдает их
	// без него придется вручную:
	// - читать registry
	// - форматировать текст
	// - обрабатывать "scrape concurrency" - это когда prometheus может делать 2 и более запроса на сбор данных (scrape request),
	// в итоге два запроса одновременно оба читают метрики
	router.Handle("/metrics", promhttp.Handler())
	log.Println("routes:")
	log.Printf("http://localhost:%d\n", port)
	log.Printf("http://localhost:%d/traced\n", port)
	log.Printf("http://localhost:%d/metrics\n", port)

	err := http.ListenAndServe(fmt.Sprintf(":%d", port), router)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
}

// Автоматически публикует метрики при выполнении запросов.
func prometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// rr - нужен, потому что стандартный http.ResponseWriter не содержит status после выполнения
		rr := &responseRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rr, r)

		duration := time.Since(start).Seconds()
		path := r.URL.Path

		// увеличиваем метрики
		metricRequestsTotal.WithLabelValues(
			r.Method,
			path,
			http.StatusText(rr.status),
		).Inc()

		metricRequestsLatency.WithLabelValues(
			r.Method,
			path,
		).Observe(duration)
	})
}

// responseRecorder - потому что стандартный интерфейс не предоставляет способ узнать код ответа после обработки запроса.
type responseRecorder struct {
	http.ResponseWriter
	status int
}

// WriteHeader переопределили для записи статусов
func (r *responseRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func initPrometheus() (*prometheus.CounterVec, *prometheus.HistogramVec) {
	// Метрика: Счетчик для кол-ва запросов
	// Counter - просто счетчик.
	// CounterVec - набор счетчиков разбитых по labels.
	// (названия hardcoded в grafana)
	requestsTotal := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Общее количество HTTP запросов",
	}, []string{"method", "path", "status"})

	// Метрика: Гистограмма для latency
	requestsLatency := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "http_request_duration_seconds",
		Help: "Латентность",
		// Стандартные сегменты {.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10}, подобранные для типовых latency веб-сервисов (в секундах).
		// .005 = ответ <= 5мс, 10 = ответ за <= 10 сек
		// Для других сценариев можно определять свои.
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path"})

	prometheus.MustRegister(requestsTotal, requestsLatency)

	return requestsTotal, requestsLatency
}

func initJaeger() func(context.Context) error {
	exp, err := jaeger.New(
		jaeger.WithCollectorEndpoint(
			jaeger.WithEndpoint("http://localhost:14268/api/traces"),
		),
	)
	if err != nil {
		log.Fatal(err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("prometheus-tracing"),
		)),
	)

	otel.SetTracerProvider(tp)
	return tp.Shutdown
}
