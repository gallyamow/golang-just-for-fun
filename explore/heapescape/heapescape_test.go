package zerovalues

import (
	"fmt"
	"testing"
)

// Escape to heap - происходит, когда компилятор понимает, что переменная должна жить дольше,
// чем функция, в которой она была объявлена. Тогда её нельзя хранить на стеке, и она «убегает» в кучу.
//
// Go компилятор умеет показывать, какие переменные убегают, с флагом -gcflags="-m":
// go build -gcflags="-m" main.go
//
// В каких то кейсах касается только указателей, в каких-то может касаться и значений.
// Убегание в кучу — не всегда плохо. Оно необходимо, когда:
//   - Переменная должна пережить функцию
//   - Используются goroutines
//   - Работа с большими структурами данных
//
// Куда размешать переменную определяется на этапе компиляции.
// Компилятор Go статически анализирует код и решает, где размещать переменные.
//
// Когда это случается:
// 1) Возврат указателя на локальную переменную
// 2) Передача указателя в другую функцию
// 3) Использование значения/указателя в замыканиях (closures)
// 4) Передача значения/указателя в каналы или goroutines
// 5) Присваивание указателя глобальной переменной
// 6) Структуры с указателями
// 7) Передача переменной как интерфейса
//
// Как реализованы:
// Компилятор заранее анализирует и размещает переменную в куче вместо stack.
func TestZeroValues(t *testing.T) {
	// 1) Возврат указателя на локальную переменную
	t.Run("return_pointer_to_local_variable", func(t *testing.T) {
		returnInnerPointer()
	})

	// 2) Передача указателя в другую функцию
	t.Run("pass_pointer_to_another_function", func(t *testing.T) {
		passPointerToAnotherFunction()
	})

	// 3) Использование в замыканиях (closures)
	t.Run("closure", func(t *testing.T) {
		closure()
	})

	// 4. Передача в каналы или goroutines
	t.Run("return_pointer_to_local_variable", func(t *testing.T) {
		passToGoroutine()
		passToChannel()
	})

	// 5) Структуры с указателями
	t.Run("return_pointer_to_local_variable", func(t *testing.T) {
		structWithPointer()
	})

	// 6) Присваивание указателя глобальной переменной
	t.Run("pass_as_interface", func(t *testing.T) {
		usingGlobal()
	})

	// 7) Передача переменной как интерфейса
	t.Run("pass_as_interface", func(t *testing.T) {
		passAsInterface()
	})
}

func returnInnerPointer() *int {
	x := 42
	return &x
}

func passPointerToAnotherFunction() {
	x := 42
	anotherFunc(&x)
}

func anotherFunc(x *int) {
	_ = x
}

func closure() {
	x := 10
	f := func() int { return x } // Замыкание захватывает x по ссылке. Компилятор не может гарантировать,
	// что замыкание не будет использовано после завершения функции
	_ = f // Убегает в кучу даже так без return
}

func passToGoroutine() {
	s := "hello"

	// Так как она работает отдельно и может работать после выхода из внешней функции
	go func() {
		fmt.Println(s)
	}()
}

func passToChannel() chan string {
	ch := make(chan string)

	s := "hello"
	go func() {
		ch <- s
	}()

	return ch // попадает наружу
}

func structWithPointer() Data {
	x := 10
	return Data{Value: &x} // попадает наружу
}

type Data struct {
	Value *int
}

func passAsInterface() {
	x := 3
	usingInterface(x) // потому что как ссылка передается в другую функцию и не понятно как там будет использоваться
}

func usingInterface(i interface{}) {
	_ = i
}

var globalX *int

func usingGlobal() {
	x := 42
	globalX = &x
}
