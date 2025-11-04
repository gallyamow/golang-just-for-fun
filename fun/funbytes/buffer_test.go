package funbytes

import (
	"bytes"
	"io"
	"os"
	"slices"
	"testing"
)

func TestUsingFunBytesBuffer(t *testing.T) {
	// Запись данных в буфер
	t.Run("writer", func(t *testing.T) {
		var bb Buffer
		writer(t, &bb)
	})

	// Чтение данных из буфера в slice
	t.Run("reader", func(t *testing.T) {
		var bb Buffer
		_, err := bb.Write([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9})
		if err != nil {
			t.Fatalf("got error %v", err)
		}
		reader(t, &bb)
	})

	t.Run("readWriter", func(t *testing.T) {
		var bb Buffer
		readWriter(t, &bb)
	})

	t.Run("proper_copy", func(t *testing.T) {
		var bb1 Buffer
		_, err := bb1.Write([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9})
		if err != nil {
			t.Fatalf("got error %v", err)
		}

		bb2 := bytes.NewBuffer(bb1.Bytes())
		if bb2.Len() != 9 {
			t.Errorf("got %d, want 9", bb2.Len())
		}
	})
}

func TestUsingBytesBuffer(t *testing.T) {
	// Запись данных в буфер
	t.Run("writer", func(t *testing.T) {
		var bb bytes.Buffer
		writer(t, &bb)
	})

	// Чтение данных из буфера в slice
	t.Run("reader", func(t *testing.T) {
		var bb bytes.Buffer
		_, err := bb.Write([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9})
		if err != nil {
			t.Fatalf("got error %v", err)
		}

		reader(t, &bb)
	})

	t.Run("readWriter", func(t *testing.T) {
		var bb bytes.Buffer
		readWriter(t, &bb)
	})

	t.Run("proper_copy", func(t *testing.T) {
		var bb1 bytes.Buffer
		_, err := bb1.Write([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9})
		if err != nil {
			t.Fatalf("got error %v", err)
		}

		bb2 := bytes.NewBuffer(bb1.Bytes())
		if bb2.Len() != 9 {
			t.Errorf("got %d, want 9", bb2.Len())
		}
	})
}

func writer(t *testing.T, bb io.Writer) {
	_, err := bb.Write([]byte("hello"))
	if err != nil {
		t.Fatalf("got error %v", err)
	}

	_, err = bb.Write([]byte{1, 2, 3})
	if err != nil {
		t.Fatalf("got error %v", err)
	}
}

func reader(t *testing.T, bb io.Reader) {
	data5 := make([]byte, 5)

	// Читаем данные из буфера в data5
	n, err := bb.Read(data5)

	if err != nil {
		t.Fatalf("got error %v", err)
	}

	if !slices.Equal([]byte{1, 2, 3, 4, 5}, data5) {
		t.Errorf("got %v, want %v", data5, []byte{1, 2, 3, 4, 5})
	}

	if n != 5 {
		t.Errorf("got %d bytes, expected 5", n)
	}

	data4 := make([]byte, 4)
	n, err = bb.Read(data4)
	if err != nil {
		t.Fatalf("got error %v", err)
	}

	if !slices.Equal([]byte{6, 7, 8, 9}, data4) {
		t.Errorf("got %v, want %v", data5, []byte{6, 7, 8, 9})
	}

	data1 := make([]byte, 2)
	n, err = bb.Read(data1)
	if err == nil || err != io.EOF {
		t.Fatalf("got non-EOF or nil error %v", err)
	}
}

func readWriter(t *testing.T, bb io.ReadWriter) {
	// io.Writer
	_, err := bb.Write([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9})
	if err != nil {
		t.Fatalf("got error %v", err)
	}

	// io.Reader
	_, err = io.Copy(os.Stdout, bb)
	if err != nil {
		t.Fatalf("got error %v", err)
	}
}

var testData = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func BenchmarkBytesBufferWrite(b *testing.B) {
	var bb bytes.Buffer
	b.ReportAllocs() // принудительный -benchmem

	for range b.N {
		bb.Reset()

		for j := 0; j < 1000; j++ {
			bb.Write(testData)
		}
	}
}

func BenchmarkFunBytesBufferWrite(b *testing.B) {
	var bb Buffer
	b.ReportAllocs()

	for range b.N {
		bb.Reset()

		for j := 0; j < 1000; j++ {
			bb.Write(testData)
		}

	}
}

func BenchmarkBuffer_Read(b *testing.B) {
	var bb bytes.Buffer
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		bb.Reset()
		for j := 0; j < 1000; j++ {
			bb.Write(testData)
		}

		readBuf := make([]byte, 64)
		for {
			_, err := bb.Read(readBuf)
			if err == io.EOF {
				break
			}
		}
	}
}

func BenchmarkFunBuffer_Read(b *testing.B) {
	var bb Buffer
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		bb.Reset()
		for j := 0; j < 1000; j++ {
			bb.Write(testData)
		}

		readBuf := make([]byte, 64)
		for {
			_, err := bb.Read(readBuf)
			if err == io.EOF {
				break
			}
		}
	}
}
