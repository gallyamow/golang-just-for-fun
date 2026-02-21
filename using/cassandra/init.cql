-- Keyspace логическое пространство данных, примерно как БД или схема в реляционных.
-- Keyspace НЕ хранит данные напрямую, он хранит правила, по которым данные живут:
--  - Replication Strategy
--  - Replication Factor (RF) - ОН задается только в рамках keyspace. Делают разные keyspace с разным RF.
--  - Durable writes
--  - Набор таблиц
--
-- Есть параметр durable_writes = true, означает сначала в commit log. Тогда данные не теряются при падении.
CREATE KEYSPACE IF NOT EXISTS demo
    -- На prod почти всегда: NetworkTopologyStrategy + RF ≥ 3
    WITH replication =
        {
            'class': 'SimpleStrategy', -- для тестовых целей
            'replication_factor': 1 -- в тестовых целях rf=1
            };

-- Благодаря настройкам primary key и clustering order:
-- Такая таблица позволяет быстро взять последние события пользователя без сортировки и без скана.
--
-- Вот такие запросы будут быстрымиЖ
-- 1) SELECT * -- последние события:
-- FROM demo.user_events
-- WHERE user_id = ?;
-- 2) SELECT * -- диапазон по времени
-- FROM demo.user_events
-- WHERE user_id = ?
--   AND event_time >= ?
--   AND event_time < ?;
-- Такие будут медленными:
-- 1) все события всех пользователей
-- SELECT * FROM demo.user_events;
-- 2) поиск по event_type
-- SELECT * FROM demo.user_events
-- WHERE event_type = 'login';
CREATE TABLE IF NOT EXISTS demo.last_user_events
(
    user_id    UUID,
    event_time TIMESTAMP,
    event_type TEXT,
    payload    TEXT,
    -- Ключи user_id → partition key, event_time → clustering key
    -- Значит:
    -- все события одного пользователя лежат в одной партиции
    -- внутри партиции они отсортированы по времени
    PRIMARY KEY (user_id, event_time)
-- события внутри партиции хранятся: от новых к старым, физический порядок на диске и в MemTable
) WITH CLUSTERING ORDER BY (event_time DESC);