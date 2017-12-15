package rsqf_test

import (
	"testing"

	. "github.com/nfisher/rsqf"
)

func Benchmark_init(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = New(1000000)
	}
}

func Benchmark_put_on_high_boundary(b *testing.B) {
	f := New(100000)
	r := &f.Q[0].Remainders
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r[0] = 0
		r[1] = 0
		r[2] = 0
		r[3] = 0
		r[4] = 0
		r[5] = 0
		r[6] = 0
		r[7] = 0
		r[8] = 0
		f.Put(0x38, 0x1FF)
	}
}

func Benchmark_put2_on_high_cell(b *testing.B) {
	f := New(100000)
	r := &f.Q[0].Remainders
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r[0] = 0
		r[1] = 0
		r[2] = 0
		r[3] = 0
		r[4] = 0
		r[5] = 0
		r[6] = 0
		r[7] = 0
		r[8] = 0
		f.Put2(0x3F, 0x1FF)
	}
}

func Benchmark_Hash(b *testing.B) {
	f := New(10000000)
	str := []byte("executed by the go test command when its -bench flag is provided")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.Hash(str)
	}
}

func Benchmark_Rank(b *testing.B) {
	var v uint64 = 0xFFFFFFFFFFFFFFFF
	for i := 0; i < b.N; i++ {
		Rank(v, 64)
	}
}
