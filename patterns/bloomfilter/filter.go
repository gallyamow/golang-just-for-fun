package cache

import (
	"github.com/spaolacci/murmur3"
	"math"
)

type BloomFilter struct {
	m    uint     // размер битового массива
	k    uint     // кол-во хэш функций
	bits []uint64 // вместо []bool будем для экономии использовать набор из uint64 и сдвиги
}

func NewBloomFilter(n uint, fpr float64) *BloomFilter {
	// Вычисляем нужные количества параметры согласно False positive rate (допустимая вероятность ложного срабатывания)
	// m = − n*ln(fpr)/ln(2)^2 - тут есть минус, потому что ln(fpr) = от маленького (<e) числа отрицателен.
	// k = (m/n)*ln(2)
	m := uint(math.Ceil(-(float64(n) * math.Log(fpr)) / (math.Ln2 * math.Ln2)))
	k := uint(math.Ceil(float64(m) / float64(n) * math.Ln2))

	words := int(math.Ceil(float64(m) / 64)) // или лучше так (m + 63) / 64 (тут при делении int/int будет усечение в сторону нуля)

	return &BloomFilter{
		m:    m,
		k:    k,
		bits: make([]uint64, words),
	}
}

func (bf *BloomFilter) Add(key string) error {
	for _, hash := range bf.murmurHashes([]byte(key)) {
		idx := hash % bf.m
		bf.setBit(idx)
	}

	return nil
}
func (bf *BloomFilter) MightContain(key string) bool {
	for _, hash := range bf.murmurHashes([]byte(key)) {
		idx := hash % bf.m
		if !bf.getBit(idx) {
			return false
		}
	}
	return true
}

func (bf *BloomFilter) setBit(idx uint) {
	word := idx / 64
	bit := idx - word*64 // или лучше idx % 64
	bf.bits[word] |= 1 << bit
}

func (bf *BloomFilter) getBit(idx uint) bool {
	word := idx / 64
	bit := idx - word*64 // или лучше idx % 64
	return bf.bits[word]&(1<<bit) != 0
}

// murmurHashes - генерация k-хешей через MurmurHash3
func (bf *BloomFilter) murmurHashes(val []byte) []uint {
	res := make([]uint, bf.k)

	// Используем с seed для того чтобы не хранить хеш-функции и иметь возможность применять одни и те же хеши несколько раз.
	// Sum32WithSeed тут заменяем New32WithSeed(i).Sum(keyBytes)
	for i := range bf.k {
		res[i] = uint(murmur3.Sum32WithSeed(val, uint32(i)))
	}

	return res
}
