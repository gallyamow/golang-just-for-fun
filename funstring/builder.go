package funstring

import "unicode/utf8"

type Builder struct {
	buf []byte
}

func (b *Builder) String() string {
	return string(b.buf)
}

func (b *Builder) Len() int {
	return len(b.buf)
}

func (b *Builder) Cap() int {
	return cap(b.buf)
}

func (b *Builder) Reset() {
	// при append будет аллокация
	b.buf = nil
}

func (b *Builder) Grow(n int) {
	space := b.Cap() - b.Len()
	if space > n {
		return
	}

	cp := make([]byte, b.Len(), 2*b.Cap()+n)
	copy(cp, b.buf)
	b.buf = cp
}

func (b *Builder) Write(p []byte) (int, error) {
	b.buf = append(b.buf, p...)
	return len(p), nil
}

func (b *Builder) WriteByte(c byte) error {
	b.buf = append(b.buf, c)
	return nil
}

func (b *Builder) WriteRune(r rune) (int, error) {
	// b.buf = append(b.buf, byte(r)) - нельзя приведение byte(r) обрезает руну до младшего байта, ломая Unicode
	b.buf = utf8.AppendRune(b.buf, r)
	return utf8.RuneLen(r), nil
}

func (b *Builder) WriteString(s string) (int, error) {
	b.buf = append(b.buf, s...)
	return len(s), nil
}
