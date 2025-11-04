package bytesbuffer

import (
	"bytes"
	"io"
	"os"
	"slices"
	"testing"
)

// bytes.Buffer предназначенная для эффективной работы с байтовыми данными (строками, файлами, сетевыми данными и т.п.).
//
// Для чего:
// 1) Не нужно вручную выделять или увеличивать массив — буфер сам растёт по мере необходимости.
// 2) Очень быстрый: минимизирует копирование памяти.
// 3) Удобен для сборки строк или двоичных данных. Часто используется при генерации данных в памяти (например,
// JSON, HTTP-запросы, логирование и т.п.).
//
// Как с ним работать:
// Ненулевой не стоит копировать. У него нет _noCopy - потому что нулевой все же можно копировать.
// Копировать его нельзя, потому что внутри срез байтов. Копии будут указывать на тот же underlying array.
// А вот смещение off у них разные.
// Копировать надо копируя срез байт в новый.
//
// Как реализован:
// Он обёртка вокруг среза байт, с отслеживанием того, что уже прочитано.
// 1) Внутри buf []byte, off - смещение чтения, lastRead - тип последней операции
// 2) При записи: смотрит есть ли место, если нет:
//   - 1) если размер буфера 0, а off не 0 - значит буфер можно переиспользовать, поэтому обнуляет off.
//   - 2) если буфер пустой и пишут немного, то делает аллокацию размером 64 элемента
//   - 3) если capacity среза позволяет увеличить len до n - делает
//   - 4) так как при чтении мы просто перемещаем off, то позади него скапливается уже прочитанное, значит мы можем просто
//     занулить off и скопировать данные в начало slice
//   - 5) крайний вариант делает 2x (исторически так сложилось, в будущем будут реагировать на темп роста), специально делает его так
//     append([]byte(nil), make([]byte, c)...) т.е. добавляем к новому slice, вместо append(b, make([]byte, n)...)
//     добавляем только нужное в текущий, чтобы не "убегала в кучу".
//     При этому получается что с начала массива будут нули, поэтому приходится копировать содержимое туда.
//
// Далее пишет данные, off - не трогает.
//
// 3) Копирует данные buf[off:] в data столько сколько влезает в data. Если данных недостаточно - то сколько есть,
// если данных нет - то EOF. Перемещает off на размер записанных данных.
func TestUsingBytesBuffer(t *testing.T) {
	// Запись данных в буфер
	t.Run("writing", func(t *testing.T) {
		var bb bytes.Buffer

		bb.WriteByte('A')
		bb.WriteString("hello")
		bb.WriteRune('Р')

		t.Logf("result %s", bb.String())
	})

	// Чтение данных из буфера в slice
	t.Run("reading", func(t *testing.T) {
		var bb bytes.Buffer
		bb.Write([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9})

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
		if err != io.EOF {
			t.Fatalf("got non-EOF error %v", err)
		}
	})

	t.Run("copy", func(t *testing.T) {
		var bb1 bytes.Buffer
		bb1.Write([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9})

		bb2 := bytes.NewBuffer(bb1.Bytes())
		if bb2.Len() != 9 {
			t.Errorf("got %d, want 9", bb2.Len())
		}
	})

	t.Run("io", func(t *testing.T) {
		var bb bytes.Buffer

		// io.Writer
		bb.Write([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9})

		// io.Reader
		_, err := io.Copy(os.Stdout, &bb)

		if err != nil {
			t.Fatalf("got error %v", err)
		}
	})

}
