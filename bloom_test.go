package bloom

import (
	"testing"
)

// 2000000	       732 ns/op
func BenchmarkGetHasherUsesStdSHA256(b *testing.B) {
	h := getHasherUsesStdSHA256(100)
	bs := []byte(string("799942312321"))
	b.Run("SHA256", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			h(bs, bs)
		}
	})
}

// 10000000	       152 ns/op
func BenchmarkGetHasherUsesCRC64(b *testing.B) {
	h := getHasherUsesCRC64(100)
	bs := []byte(string("799942312321"))
	b.Run("CRC64", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			h(bs, bs)
		}
	})
}

func BenchmarkBloomFilter_Put(b *testing.B) {
	item := []byte(string("799942312321"))
	var mask [100]byte
	h := getHasherUsesCRC64(100)
	b.Run("Put test", func(b *testing.B) {
		hash := h(item, item)
		mask[hash] |= 1 << (hash & 7)
	})
}