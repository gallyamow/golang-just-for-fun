package syncpool

import (
	"sync"
	"testing"
)

// TestUsingSyncMap
//
// Для чего он:
// sync.Map — это потокобезопасная map.
// Его преимущества проявляются когда: много goroutines, много чтений, мало записей.
// Если записей много - то mutex + map - лучше.
//
// Типичный кейс использования:
// Кэш с конкурентным доступом
//
// Как с ним работать:
// 1) Store - для записи. И ключи и значения - any.
// 2) Load - для чтения. Так как any - нужно приведение типов.
// 3) Delete - для удаления
// 4) LoadOrStore - upsert
// 5) Range - итерирование (возврат false останавливает)
//
// Как реализован:
// Внутри 2 map:
// - read для чтения, lock free
// - dirty для записи под mutex
// Также внутри счетчик промахов: misses int
//
//	type Map struct {
//		mu Mutex
//		read atomic.Pointer[readOnly]
//		dirty map[any]*entry
//		misses int
//	}
//
//	type readOnly struct {
//		m       map[any]*entry
//		amended bool // true = "в dirty есть ключи, которых нет в read"
//	}
//
//	type entry struct {
//	   p atomic.Pointer[any]
//	}
//
// При load:
// 1) Смотрим в read map (БЕЗ LOCK)
// 2) Если нашли — возвращаем
// 3) Если НЕ нашли И amended = true -> идём в dirty map под mutex
// 4) Если нашли в dirty — увеличиваем misses
//
// 5) Если слишком много промахов → dirty копируется в read. Так: dirty → становится новым read, dirty → nil, misses → 0, amended → false
//
// При Store:
// Если ключ есть в read:
// 1) Берём entry
// 2) atomically меняем pointer
// 3) Все (без mutex)
// Если ключа нет в read:
// 1) lock(mu)
// 2) создаём dirty, если её нет
// 3) кладём ключ в dirty
// 4) read.amended = true
func TestUsingSyncMap(t *testing.T) {
	var sm sync.Map

	sm.Store("key1", 1)
	sm.Store(2, 3)
	sm.Store([3]int{1, 2, 3}, []string{"1", "2", "3"})

	v1, ok := sm.Load("key1")
	if !ok {
		t.Fatalf("unexpected fail")
	}
	if v1 != 1 {
		t.Fatalf("want 1, got %v", v1)
	}

	// для использования надо приводить типы
	x := v1.(int) + 2
	if x != 3 {
		t.Fatalf("want 3, got %v", x)
	}

	sm.Delete(2)
	v2, ok := sm.Load(2)
	if ok {
		t.Fatalf("unexpected val %v", v2)
	}

	v, ok := sm.LoadOrStore("key2", 13.0)
	if ok {
		t.Fatalf("unexpected val %v", v)
	}

	sm.Range(func(key, value any) bool {
		t.Logf("%s = %v", key, value)
		return true
	})
}
