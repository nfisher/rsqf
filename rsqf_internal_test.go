package rsqf

import (
	"hash/fnv"
	"testing"
	"unsafe"
)

func Test_New_filter_should_be_initialised_correctly(t *testing.T) {
	f := New(100000)
	if 26 != f.p {
		t.Errorf("want 26, got %v", f.p)
	}

	if 17 != f.quotient {
		t.Errorf("want 17, got %v", f.quotient)
	}

	if 2048 != len(f.Q) {
		t.Errorf("want len(Q) = 2048, got %v", len(f.Q))
	}

	var expected uint64 = 0x1FF
	if expected != f.rMask {
		t.Errorf("want rMask = 0x%X, got 0x%X", expected, f.rMask)
	}

	expected = 0x3FFFE00
	if expected != f.qMask {
		t.Errorf("want qMask = 0x%X, got 0x%X", expected, f.qMask)
	}
}

func Test_Hash_should_split_quotient_and_remainder_correctly(t *testing.T) {
	h := fnv.New64a()
	h.Sum([]byte("Hello world"))
	if 0xCBF29CE484222325 != h.Sum64() {
		t.Errorf("want Sum64 = 0xCBF29CE484222325, got 0x%X", h.Sum64())
	}

	f := New(100000)
	sum := f.Hash([]byte("Hello world"))

	if 0x1111 != sum.h0 {
		t.Errorf("want h0 = 0x1111, got 0x%X", sum.h0)
	}

	if 0x125 != sum.h1 {
		t.Errorf("want h1 = 0x125, got 0x%X", sum.h1)
	}
}

func Test_struct_is_contiguous(t *testing.T) {
	f := New(10000)
	p0 := unsafe.Pointer(&f.Q[3])
	p1 := unsafe.Pointer(&f.Q[4])
	sz := unsafe.Sizeof(block{})
	// 1(offset) + 11(occupieds, runends, remainders) * 8 + 1 (remainders len)
	if sz != 0x60 {
		t.Errorf("got sz = 0x%X, want 0x60\n0x%X\n0x%X", sz, p0, p1)
	}
}

func test_sample_p_values(t *testing.T) {
	p := calcP(100000, 0.05)
	if 0.1 != p {
		t.Errorf("%v", p)
	}
}
