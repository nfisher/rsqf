package rsqf

import (
	"testing"
	"unsafe"
)

func Test_firstAvailableSlot(t *testing.T) {
	t.Parallel()
	td := [][]uint64{
		// x, occupieds, runends, bi, expected
		{0x00, 0x00, 0x00, 0, 0x00},
		{0x00, 0x01, 0x01, 0, 0x01},
		{0x00, 0x01, 0x08, 0, 0x04},
		{0x01, 0x01, 0x02, 0, 0x02},
		{0x00, 0x01, 0x02, 0, 0x02},
		{0x00, 0x01, 0x04, 0, 0x03},
		{0x02, 0x02, 0x02, 0, 0x02},
		{0x02, 0x02, 0x04, 0, 0x03},
		{0x03, 0x0F, 0x0F, 0, 0x04},
		{0x80, 0x00, 0x00, 0, 0x80},
		{0x3FFFFFF, 0x00, 0x00, 0x7FF, 0x3FFFE},
	}

	for i, v := range td {
		f := New(100000)

		bi := v[3]
		f.Q[bi].Occupieds = v[1]
		f.Q[bi].Runends = v[2]

		x := v[0] & (f.qMask | f.rMask)
		actual, _ := f.firstAvailableSlot(x)
		expected := v[4]
		if expected != actual {
			t.Errorf("[%v] want FAS(0x%X) = %v, got %v",
				i, x, expected, actual)
		}
	}
}

func Test_Select(t *testing.T) {
	t.Parallel()
	td := [][]uint64{
		// B, count, expected
		{0x0, 1, 64},
		{0x1, 1, 0},
		{0x3, 2, 1},
		//xFFFFFFFFFFFFFFFF
		{0x8800000000000000, 2, 63},
		{0x8000000000000000, 1, 63},
	}

	for i, v := range td {
		a := Select(v[0], v[1])
		if v[2] != a {
			t.Errorf("[%v] want select(B, %v) = %v, got %v", i, v[1], v[2], a)
		}
	}
}

func Test_Rank(t *testing.T) {
	td := [][]uint64{
		// input, pos, expected
		{0x0, 64, 0},
		{0x1, 0, 1},
		{0x2, 1, 1},
		{0x3, 1, 2},
		{0x4, 64, 1},
		{0x5, 64, 2},
		{0x6, 64, 2},
		{0x7, 64, 3},
		{0x8, 64, 1},
		{0x9, 64, 2},
		{0xA, 64, 2},
		{0xB, 64, 3},
		{0xC, 64, 2},
		{0xD, 64, 3},
		{0xE, 64, 3},
		{0xF, 64, 4},
		{0xFFFFFFFFFFFFFFFF, 64, 64},
	}

	for i, v := range td {
		a := Rank(v[0], 63)
		if v[2] != a {
			t.Errorf("[%v] want rank(B, 63) = %v, got %v", i, v[2], a)
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

func Test_sample_p_values(t *testing.T) {
	t.Parallel()
	td := [][]int64{
		{100000, 26},
		{1000000, 29},
		{10000000, 33},
		{100000000, 36},
		{1000000000, 39},
	}

	for i, v := range td {
		p := calcP(float64(v[0]), 1.0/512.0)

		if v[1] != int64(p) {
			t.Errorf("[%v] want calcP(%v, 1/512) = %v, got %v",
				i, v[0], v[1], p)
		}
	}
}
