package stackgrowth

import (
	"math/rand"
	"unsafe"
)

// go:noinline
// @idiomatic: blank identifier assignment - способ обойти запрет на объявление переменной без ее использования
func allocateArray(depth int) {
	if depth == 0 {
		return
	}

	arr := [1_000_00]int{}
	_ = arr

	allocateArray(depth - 1)
}

// go:noinline
// @idiomatic: blank identifier assignment - способ обойти запрет на объявление переменной без ее использования
func allocateSlice(depth int) {
	if depth == 0 {
		return
	}

	sl := make([]int, rand.Intn(1_000_00)+1000)
	_ = sl

	allocateSlice(depth - 1)
}

// ConfirmForArray - подтверждает в случае использования массива фиксированного размера.
// Используем uintptr потому что иначе компилятор видит что переменная "утекает" наружу и сразу аллоцирует ее в куче
func ConfirmForArray() (uintptr, uintptr) {
	var a = 10

	// используем uintptr потому что иначе компилятор видит что переменная "утекает" наружу и сразу аллоцирует ее в куче
	p1 := uintptr(unsafe.Pointer(&a))

	allocateArray(2000)

	p2 := uintptr(unsafe.Pointer(&a))

	return p1, p2
}

// ConfirmRandSizedSlice - подтверждает в случае использования среза случайной длины.
func ConfirmRandSizedSlice() (uintptr, uintptr) {
	var a = 10

	// используем uintptr потому что иначе компилятор видит что переменная "утекает" наружу и сразу аллоцирует ее в куче
	p1 := uintptr(unsafe.Pointer(&a))

	allocateSlice(2000)

	p2 := uintptr(unsafe.Pointer(&a))

	return p1, p2
}

// ConfirmNoChangesForEscapedValue - подтверждает что компилятор сразу помещает значение в кучу в случае использования указателя.
// Странно что он не видит, что указатель не уходит за пределы функции.
func ConfirmNoChangesForEscapedValue() (uintptr, uintptr) {
	var a = 10

	p1 := &a

	allocateArray(2000)

	p2 := &a

	return uintptr(unsafe.Pointer(p1)), uintptr(unsafe.Pointer(p2))
}
