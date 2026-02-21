package main

import (
	"fmt"
	"github.com/gocql/gocql"
	"log"
	"time"
)

// Для чего:
// Apache Cassandra — распределённая NoSQL БД, оптимизированная под большие объёмы данных, высокую запись и горизонтальное масштабирование.
// Главная фишка: это не база данных, а распределённый лог с быстрым доступом по ключу и mutli-master.
// Написана на java.
// Консоль для выполнения запросов - cqlsh, запросы делают на языке - Cassandra Query Language (CQL).
//
// Особенности:
// - нет мастер ноды
// - из-за этого нет single point of failure
// - при добавлении новых нод, больше места, больше RPS, больше отказоустойчивости.
// - быстрая запись
// - нет транзакций, сложных where
// - нет автоинкремента, глобальных sequences
// - нет join, все что нужно для запроса должно лежать в отдельной таблице и партицированнно по какому-то ключу (partition key)
// - не умеет сортировать данные на лету, ORDER BY работает только по clustering key
// - нельзя фильтровать по clustering key без указания partition key
// - порядок clustering columns в PRIMARY KEY критичен
// - нужно ограничивать размер партиции. Например, если храним лог действий пользователя, то PRIMARY KEY (user_id, event_time) -
// партиция будет огромная, потому что все действия пользователя в одной партиции. Поэтому здесь лучше PRIMARY KEY ((user_id, event_day), event_time) -
// т.е. еще и по дням разделим.
//
// Хорошо подходит для:
// - метрик, события
// - time series
// - chats, iot, billing events
//
// Как работает:
// 1) запись => commit log (диск, append only file) - это default включено, пока недоступны для чтения => memtable (в ram), уже доступны для чтения
// и отсортированы по ключу. При превышении размера они вытeсняются (flush) => flush в SSTable (диск). Это Sorted String Table он immutable никогда не меняется, тысячи файлов.
// insert добавляют новые строки, update - тоже, delete - ставят метку (tombstone).
// Эти тысячи файлов потом объединяются (compaction), с удалением tombstoned, c обработкой updated.
// 2) чтение => bloom filter (проверяет есть ли ключ в SSTable, если нет, то значит в memtable - вернем их памяти).
// Если да, читает SSTable, делается merge (применение выбор последней версии, применение tombstones)
// Tombstones - живут до compaction, если их много, снижают скорость read.
// Bloom filter - вероятностная структура, позволяет быстро понять "этого ключа здесь точно НЕТ". Внутри: битовый массив и
// несколько hash функций. Каждый ключ прогоняется через N-хешей. Каждый хеш указывает на бит в массиве, этому биту ставят 1.
// При проверке - считаются все хеши для ключа, проверяют все биты, если хоть один не установлен, то точно нет записи. Если все 1 = скорее всего есть,
//
// Как с ней работать:
// 1) Основной концепт - partition key
// PRIMARY KEY (user_id, event_time) - user_id → partition key, event_time → clustering key
// Partition key → определяет, в какую партицию попадут данные
// Clustering key → определяет, в каком порядке строки хранятся внутри партиции. Можно задать ASC/DESC. Помогает делать range-запросы по диапазону.
// Можно задать несколько полей, сортироваться будет сначала по одному, потому по другому.
// 2) Под конкретный запрос делают таблицу (так как нет join), дублирование данных - норма.
// В Cassandra модель проектируют под запросы, а clustering key — основной инструмент для этого.
// 3) Consistency выбираешь сам, можно задать:
// ONE — быстрее, но риск - может потеряться, подтверждает только нода куда писалось.
// QUORUM — баланс, подтверждает кворум, то есть большинство. Есть понятие replication factor (rf), например, если rf=5, то надо 3 подтверждения.
// ALL — медленно, но строго - подтверждает все. Если хотя бы одна реплика упала - запрос падает.
//
// Для ключей можно использовать
// UUID - идеален для partition key: равномерное распределение (не будет hot partitions), максимум масштабируемости
// TIMEUUID - содержит timestampt, естественная сортировка, можно делать range запросы range-запросы по времени event_time > maxTimeuuid('2025-01-01')
//
// Что такое SSTable: Sorted String Table — это файл на диске, с:
// immutable - один раз записывается (поэтому append-only), нет races, нет random writes, нет update in-place
// sorted - позволяет делать range scan, эффективно использовать index, быстро искать partition key
// lock-free
//
// Со временем таких файлов становится много, поэтому они сливаются, делается compaction.
// Итого: Фишка SSTable — в immutable + sorted структуре, которая превращает хаотичные записи в быстрый append-only поток и перекладывает сложность на compaction.
func main() {
	cluster := gocql.NewCluster("localhost")
	cluster.Keyspace = "demo"
	cluster.Consistency = gocql.Quorum
	cluster.Timeout = 10 * time.Second

	session, err := cluster.CreateSession()
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()

	fmt.Println("Connected to Cassandra")

	userID := gocql.TimeUUID()

	// write some events
	for range 10 {
		err = insertUserEvent(session, userID, "login", `{"ip":"127.0.0.1"}`)
		if err != nil {
			log.Fatal(err)
		}
	}

	// print all events
	printUserEvents(session, userID)
}

func insertUserEvent(session *gocql.Session, userID gocql.UUID, eventType, payload string) error {
	return session.Query(`
		INSERT INTO last_user_events (user_id, event_time, event_type, payload)
		VALUES (?, ?, ?, ?)`,
		userID,
		time.Now(),
		eventType,
		payload,
	).Exec()
}

func printUserEvents(session *gocql.Session, userID gocql.UUID) error {
	// итератор
	iter := session.Query(`
		SELECT event_time, event_type, payload
		FROM last_user_events
		WHERE user_id = ?`,
		userID,
	).Iter()

	var (
		eventTime time.Time
		eventType string
		payload   string
	)

	for iter.Scan(&eventTime, &eventType, &payload) {
		fmt.Println(eventTime, eventType, payload)
	}

	return iter.Close()
}
