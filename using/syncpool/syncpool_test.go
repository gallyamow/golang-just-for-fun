package syncpool

import (
	"math/rand"
	"sync"
	"testing"
)

// TestUsingSyncPool показывает пример как используется sync.Pool для переиспользования объектов.
// sync.Pool — это структура из стандартной библиотеки sync, которая предназначена для эффективного повторного
// использования временных объектов.
//
// Как реализован:
// sync.Pool — это не просто список объектов. Каждый пул имеет локальные хранилища для каждого потока (P), чтобы минимизировать блокировки.
// Есть глобальный пул, который используется, если локальный пуст.
// Главная фишка - минимум блокировок, поддержка GC, повторное использование объектов - что дает экономию памяти и CPU для
// тяжело инициируемых.
//
// Get:
// 1) Пытается взять объект из локального пула текущего потока.
// 2) Если локальный пул пуст → берёт из глобального пула (victim). Если он не пуст, то берётся объект и помещается
// в локальный пул текущего потока.
// 3) Если глобальный пуст → вызывает New, если оно задано.
// 4) В victim хранятся значения для очистки через GC.
// Put:
// 1) Кладёт объект в локальный пул текущего потока.
// 2) Не блокирует другие потоки.
func TestUsingSyncPool(t *testing.T) {
	var decoderPool = sync.Pool{
		// переопределяем именно
		New: func() interface{} {
			return &SomeDecoder{
				instance: rand.Intn(1000),
			} // возвращаем указатель
		},
	}

	var instances []int
	for range 100 {
		decoder := decoderPool.Get().(*SomeDecoder)
		decoder.Use()
		instances = append(instances, decoder.instance)
		decoderPool.Put(decoder)
	}

	for i := 0; i < len(instances)-1; i++ {
		if instances[i] != instances[i+1] {
			t.Fatalf("used several instances")
		}
	}
}

type SomeDecoder struct {
	instance int
	ready    bool
}

func (d SomeDecoder) Use() {
	d.ready = false
}

func (d SomeDecoder) Reset() int {
	return d.instance
}
