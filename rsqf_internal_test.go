package rsqf

import (
	"testing"
	"unsafe"
)

func Test_Rank(t *testing.T) {
	td := [][]uint64{
		// input, expected
		{0x0, 0},
		{0x1, 1},
		{0x2, 1},
		{0x3, 2},
		{0x4, 1},
		{0x5, 2},
		{0x6, 2},
		{0x7, 3},
		{0x8, 1},
		{0x9, 2},
		{0xA, 2},
		{0xB, 3},
		{0xC, 2},
		{0xD, 3},
		{0xE, 3},
		{0xF, 4},
		{0xFFFFFFFFFFFFFFFF, 64},
	}

	for i, v := range td {
		a := Rank(v[0], 64)
		if v[1] != a {
			t.Errorf("[%v] want rank(B, 63) = %v, got %v", i, v[1], a)
		}
	}
}

func Test_New_filter_should_be_initialised_correctly(t *testing.T) {
	t.Parallel()
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

func Test_Hash_should_provide_expected_sum(t *testing.T) {
	t.Parallel()

	f := New(100000)
	sum := f.Hash([]byte("Hello world"))

	if 0xCBF29CE484222325 != sum {
		t.Errorf("want sum = 0xCBF29CE484222325, got 0x%X", sum)
	}
}

func Test_struct_is_contiguous(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
	p := calcP(100000, 0.05)
	if 0.1 != p {
		t.Errorf("%v", p)
	}
}
