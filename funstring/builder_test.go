package funstring

import (
	"testing"
	"unicode/utf8"
)

func TestBuilder(t *testing.T) {
	t.Run("String", func(t *testing.T) {
		b := &Builder{}
		b.WriteString("hello")

		if got := b.String(); got != "hello" {
			t.Errorf("String() = %q, want %q", got, "hello")
		}
	})

	t.Run("Len", func(t *testing.T) {
		b := &Builder{}
		b.WriteString("123")

		if b.Len() != 3 {
			t.Errorf("Len() = %d, want 3", b.Len())
		}
	})

	t.Run("Cap", func(t *testing.T) {
		b := &Builder{}
		initialCap := b.Cap()
		n := 10
		b.Grow(n)

		if b.Cap() != (initialCap + n) {
			t.Errorf("Cap() = %d, want %d", b.Cap(), initialCap+n)
		}
	})

	t.Run("Reset", func(t *testing.T) {
		b := &Builder{}
		b.WriteString("123")
		b.Reset()

		if b.Len() != 0 {
			t.Errorf("Len() = %d, want 0", b.Len())
		}
		if b.String() != "" {
			t.Errorf("String() = %q, want %q", b.String(), "")
		}
	})

	t.Run("Grow", func(t *testing.T) {
		b := &Builder{}
		initialCap := b.Cap()
		n := 10
		grown := 2*initialCap + n
		b.Grow(n)

		if b.Len() != 0 {
			t.Errorf("Len() = %d, want 0", b.Len())
		}
		if b.Cap() != grown {
			t.Errorf("Len() = %d, want 0", grown)
		}
	})

	t.Run("Write", func(t *testing.T) {
		b := &Builder{}
		data := []byte{'a', 'b', 'c'}
		n, err := b.Write(data)

		if err != nil {
			t.Errorf("Write() error = %v", err)
		}
		if n != len(data) {
			t.Errorf("Write() = %d, want %d", n, len(data))
		}
		if b.String() != string(data) {
			t.Errorf("Write() content = %v, want %v", []byte(b.String()), data)
		}
	})

	t.Run("WriteByte", func(t *testing.T) {
		b := &Builder{}
		err := b.WriteByte('a')

		if err != nil {
			t.Errorf("Write() error = %v", err)
		}
		if b.String() != "a" {
			t.Errorf("WriteByte() content = %v, want %v", []byte(b.String()), "a")
		}
	})

	t.Run("WriteRune", func(t *testing.T) {
		b := &Builder{}
		n, err := b.WriteRune('я')

		if err != nil {
			t.Errorf("WriteRune() error = %v", err)
		}
		if n != utf8.RuneLen('я') {
			t.Errorf("WriteRune() = %d, want %d", n, utf8.RuneLen('я'))
		}
		if b.String() != "я" {
			t.Errorf("WriteRune() = %q, want 'я'", b.String())
		}
	})

	t.Run("WriteString", func(t *testing.T) {
		b := &Builder{}
		str := "test"
		n, err := b.WriteString(str)

		if err != nil {
			t.Errorf("WriteString() error = %v", err)
		}
		if n != len(str) {
			t.Errorf("WriteString() = %d, want %d", n, len(str))
		}
		if b.String() != str {
			t.Errorf("WriteString() = %q, want %q", b.String(), str)
		}
	})

	t.Run("Combined", func(t *testing.T) {
		b := &Builder{}

		b.WriteString("hello")
		b.WriteByte(' ')
		b.WriteRune('世')
		b.Write([]byte{' ', 'w', 'o', 'r', 'l', 'd'})

		want := "hello 世 world"
		if got := b.String(); got != want {
			t.Errorf("Combined operations = %q, want %q", got, want)
		}
	})
}
