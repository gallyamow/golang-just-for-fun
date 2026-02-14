package main

import (
	"context"
	"fmt"
	"github.com/IBM/sarama"
	"log"
	"os"
	"os/signal"
	"time"
)

// Как используется:
// Это очень быстрый (и распределенный) лог сообщений (не очередь). Используется обычно как:
// - streaming событий (а не для управления задачами)
// - event-driven системы
//
// Почему быстрый? (*)
// - append only запись на диск
// - zero-copy (данные идут из диска в сеть без копирования, использует системный вызов sendfile())
// - партиции обрабатываются параллельно
//
// Из чего состоит:
// 1) producer, consumer - пишет, читает, broker - сервер kafka
// 2) topic - логический канал сообщений
// 3) partition - топик разбит на партиции для масштабирования и ПОРЯДОК гарантируется только внутри одной партиции.
// Хотите порядок — используйте ключ сообщения, чтобы все события одного объекта попадали в одну партицию.
// 4) consumer group - группа consumer, где одну партицию читает только один consumer
// 5) offset - позиция сообщения в партиции, его хранит сам consumer. Поэтому он может перемещаться и перечитывать сообщения.
// Это ключевое отличие от RabbitMQ.
//
// Репликация:
// У каждой партиции есть Leader и Followers (реплики), если лидер падает => выбирается новый (raft). Новые записи
// в партицию идут только в лидера. Реплики - получают записи от лидера асинхронно.
// Фактор репликации (replication.factor) ставят обычно 3 => есть 3 реплики одной партиции на 3 разных брокерах.
// Таким образом партиция выдерживает 2 падения лидеров (на практике одно - потому что записи не будет в таком случае).
// Оно не может быть больше кол-ва брокеров.
// min.insync.replicas = 2 - кол-во минимальное живых реплик чтобы запись была активной. Живых = успевающих за лидером.
//
// Гарантии доставки (*):
// 1) at most once = максимум один раз (минимум - ни разу) => может потеряться => настраивается через acks=0
// 2) at least once = как минимум один раз (может и два, и больше) => может повторяться => настраивается через acks=all
// 3) exactly once - ровно один раз = сложно сделать на практике => требуется:
// idempotent producer (enable.idempotence=true) + transactional writes (transactional.id) + управление offset через транзакцию (для консюмера)
//
// acks=N означает кол-во in-sync replicas (ISR) живых реплик которые подтвердили получения и после этого возвращается успех producer.
// acks=0 - не ждет => возможна потеря сообщений
// acks=1 - только лидер подтвердил => возможна потеря при падении лидера
// acks=all - все isr подтвердили
//
// На практике чаще всего используют at least once + идемпотентность.
// Пишется все в лидера - потому что так просто сохранять порядок и замену лидера.
//
// Подробнее:
//   - Consumer groups — это ключевой механизм для масштабирования и управления обработкой сообщений в Kafka.
//     При добавлении либо удалении (падении) consumer происходит перебалансировка. Kafka стремится равномерно распределить
//     все партиции среди всех consumer.
//     Топик: 4 партиции, Consumer Group: 2 consumer => Распределение: C1 → партиции 0,1 и C2 → партиции 2,3.
//     Разные consumer groups читают один топик независимо у них свой offset.
//   - offset - Каждое сообщение в партиции получает уникальный, возрастающий номер. Чтобы прочитать следующее сообщение,
//     consumer смотрит свой последний offset + 1.
//     offset хранится для группы.
//
// Чем отличается ConsumerGroup и Consumer без группы:
// - consumer = low-level consumer - за offset придется следить самому и сохранять его самому где-то
// - не будет участвовать в балансировке, будет читать со всех заданных партиции
// - его использую в небольших утилитах и скриптах, в тестах и логировании где не нужно масштабирование

const (
	broker = "localhost:9092"
	topic  = "some_topic"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go producer(ctx)
	go groupOfConsumers(ctx)
	go oneConsumer(ctx)

	// waiting ctrl+c
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, os.Interrupt)
	<-sigterm
	fmt.Println("Shutting down, context will be cancelled...")
}

func producer(ctx context.Context) {
	config := sarama.NewConfig()
	// требуем подтверждения leader + всех живых followers
	config.Producer.RequiredAcks = sarama.WaitForAll
	// sarama producer не возвращает информацию об успешно отправленных сообщениях. Нам нужны offset и partition, требуем возвращать.
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer([]string{broker}, config)
	if err != nil {
		log.Fatal(err)
	}
	defer producer.Close()

	i := 0
	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg := &sarama.ProducerMessage{
				Topic: topic,
				Key:   sarama.StringEncoder(fmt.Sprintf("key-%d", i)),
				Value: sarama.StringEncoder(fmt.Sprintf("message-%d", i)),
			}

			partition, offset, err := producer.SendMessage(msg)
			if err != nil {
				log.Println("Failed to produce message:", err)
			} else {
				fmt.Printf("Produced message offset %d to partition %d\n", offset, partition)
			}

			i++
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func groupOfConsumers(ctx context.Context) {
	config := sarama.NewConfig()

	// Kafka будет использовать первую стратегию, которая подходит для всех участников consumer group.
	// (т.е. fallback вариант для например consumers со старыми стратегиями)
	config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{
		sarama.NewBalanceStrategyRange(),
	}

	consumerGroup, err := sarama.NewConsumerGroup([]string{broker}, "some_consumers_group", config)
	if err != nil {
		log.Fatalln("Failed to create consumer group:", err)
	}
	defer consumerGroup.Close()

	// конфигурируем handler
	handler := groupConsumersHandler{}

	// В отличии, от rabbit тут сами "тянем" данные
	for {
		select {
		case <-ctx.Done():
			return
		default:
			err := consumerGroup.Consume(ctx, []string{topic}, handler)
			if err != nil {
				log.Println("Error consuming:", err)
			}
		}
	}
}

// Реализуем ConsumerGroupHandler.
// Не забываем что один и тот же instance запускается из разных goroutines и заботимся о thread-safety.
type groupConsumersHandler struct {
}

// Setup is run at the beginning of a new session, before ConsumeClaim.
func (h groupConsumersHandler) Setup(s sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
// but before the offsets are committed for the very last time.
func (h groupConsumersHandler) Cleanup(s sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
// Once the Messages() channel is closed, the Handler must finish its processing
// loop and exit.
func (h groupConsumersHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		fmt.Printf("GROUP: Consumed message offset %d: %s = %s\n", msg.Offset, string(msg.Key), string(msg.Value))

		// так может делать только группа consumers
		sess.MarkMessage(msg, "") // коммит offset
	}
	return nil
}

func oneConsumer(ctx context.Context) {
	consumer, err := sarama.NewConsumer([]string{broker}, nil)
	if err != nil {
		log.Fatal("Failed to create consumer:", err)
	}
	defer consumer.Close()

	// Получаем все партиции топика
	partitions, err := consumer.Partitions(topic)
	if err != nil {
		log.Fatal("Failed to get partitions:", err)
	}

	// Для каждой партиции запускаем отдельную goroutine
	for _, partition := range partitions {
		go func(p int32) {
			partitionConsumer, err := consumer.ConsumePartition(topic, p, sarama.OffsetOldest)
			if err != nil {
				log.Printf("Failed to start p consumer %d: %v\n", p, err)
				return
			}
			defer partitionConsumer.Close()

			for {
				select {
				case <-ctx.Done():
					return
				case msg := <-partitionConsumer.Messages():
					fmt.Printf("ONE: Partition %d: offset %d: %s = %s\n", partition, msg.Offset, string(msg.Key), string(msg.Value))
				case err := <-partitionConsumer.Errors():
					fmt.Printf("ONE: Partition %d: consumer error: %v\n", partition, err)
				}
			}

		}(partition)
	}

	// @idiomatic: common pattern for long-running goroutines
	<-ctx.Done()
}
