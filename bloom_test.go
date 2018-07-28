package bloom

import (
	"testing"
)

func BenchmarkGetHasherUsesStdSHA256(b *testing.B) {
	h := getHasherUsesStdSHA256(100)
	bs := []byte(string("799942312321"))
	b.Run("SHA256", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			h(bs, bs)
		}
	})
}

func BenchmarkGetHasherUsesCRC64(b *testing.B) {
	h := getHasherUsesCRC64(100)
	bs := []byte(string("799942312321"))
	b.Run("CRC64", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			h(bs, bs)
		}
	})
}
