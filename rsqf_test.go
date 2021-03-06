// rsqf provides a Rank-and-Select Quotient Filter which is an approximate
// membership datastructure that allows INSERT, QUERY, and DELETE.
//
// references
// ==========
//
// paper: https://www3.cs.stonybrook.edu/~ppandey/files/p775-pandey.pdf
// summary: https://blog.acolyer.org/2017/08/08/a-general-purpose-counting-filter-making-every-bit-count/
//
package rsqf_test

import (
	"testing"

	. "github.com/nfisher/rsqf"
)

func Test_Insert_simple_run(t *testing.T) {
	t.Parallel()
	f := New(100000)
	f.Insert(0x01F0)
	f.Insert(0x01FF)

	if 0x02 != f.Q[0].Runends {
		t.Errorf("want Q[0].Runends = 0x%X, got 0x%X", 0x02, f.Q[0].Runends)
	}

	if 0x01 != f.Q[0].Occupieds {
		t.Errorf("want Q[0].Occupieds = 0x%X, got 0x%X",
			0x01, f.Q[0].Occupieds)
	}

	if 0x3FFF0 != f.Q[0].Remainders[0] {
		t.Errorf("want Q[0].Remainders[0] = 0x3FFF0, got 0x%X", f.Q[0].Remainders[0])
	}
}

func Test_Insert_simple(t *testing.T) {
	t.Parallel()
	td := [][]uint64{
		// insert, runends, occupieds
		{0x01F0, 0x01, 0x01,
			0x01F0, 0x0, 0x0,
			0x0, 0x0, 0x0,
			0x0, 0x0, 0x0},
		{0x03F0, 0x02, 0x02,
			0x3E000, 0x0, 0x0,
			0x0, 0x0, 0x0,
			0x0, 0x0, 0x0},
	}

	for i, v := range td {
		h := v[0]
		re := v[1]
		o := v[2]

		f := New(100000)
		f.Insert(h)

		if re != f.Q[0].Runends {
			t.Errorf("[%v] want Q[0].Runends = 0x%X, got 0x%X",
				i, re, f.Q[0].Runends)
		}

		if o != f.Q[0].Occupieds {
			t.Errorf("[%v] want Q[0].Occupieds = 0x%X, got 0x%X",
				i, o, f.Q[0].Occupieds)
		}

		r0 := v[3]
		if r0 != f.Q[0].Remainders[0] {
			t.Errorf("[%v] want Q[0].Remainders = 0x%X, got 0x%X",
				i, r0, f.Q[0].Remainders[0])
		}
	}
}

func Test_Put2_run_element(t *testing.T) {
	t.Parallel()
	f := New(100000)
	f.Put2(0x00, 0x1F0)
	f.Put2(0x01, 0x10F)

	Q := f.Q[0]

	td := []uint64{
		0x0000000000000002, 0x0000000000000002, 0x0000000000000002, 0x0000000000000002,
		0x0000000000000001, 0x0000000000000001, 0x0000000000000001, 0x0000000000000001,
		0x0000000000000003,
	}

	for i, v := range td {
		remainders := v
		if remainders != Q.Remainders[i] {
			t.Errorf("want Q[0].Remainders[%v] = 0x%X, got 0x%X", i, remainders, Q.Remainders[i])
		}
	}
}

func Test_Put2_in_same_block_without_run(t *testing.T) {
	t.Parallel()
	td := [][]uint64{
		// h0,   h1,        Q.occupieds,
		// Q[0].Remainders[0], Q[0].Remainders[1], Q[0].Remainders[2],
		// Q[0].Remainders[3], Q[0].Remainders[4], Q[0].Remainders[5],
		// Q[0].Remainders[6], Q[0].Remainders[7], Q[0].Remainders[8]
		// b

		// 0 - 1st block, first rank, partial r bits on
		{0x00, 0x1F0, 0x0000000000000001,
			0x0000000000000000, 0x0000000000000000, 0x0000000000000000,
			0x0000000000000000, 0x0000000000000001, 0x0000000000000001,
			0x0000000000000001, 0x0000000000000001, 0x0000000000000001,
			0},
		// 1 - 1st block, first rank, all r bits on
		{0x00, 0x1FF, 0x0000000000000001,
			0x0000000000000001, 0x0000000000000001, 0x0000000000000001,
			0x0000000000000001, 0x0000000000000001, 0x0000000000000001,
			0x0000000000000001, 0x0000000000000001, 0x0000000000000001,
			0},
		// 2 - 1st block, last rank, all r bits on
		{0x3F, 0x1FF, 0x8000000000000000,
			0x8000000000000000, 0x8000000000000000, 0x8000000000000000,
			0x8000000000000000, 0x8000000000000000, 0x8000000000000000,
			0x8000000000000000, 0x8000000000000000, 0x8000000000000000,
			0},
		// 3 - 2nd block, first rank, all r bits on
		{0x40, 0x1FF, 0x0000000000000001,
			0x0000000000000001, 0x0000000000000001, 0x0000000000000001,
			0x0000000000000001, 0x0000000000000001, 0x0000000000000001,
			0x0000000000000001, 0x0000000000000001, 0x0000000000000001,
			1},
	}

	for i, v := range td {
		f := New(100000)
		f.Put2(v[0], v[1])
		b := v[12]
		Q := f.Q[b]

		for j := 0; j < 9; j++ {
			remainders := v[3+j]
			if remainders != Q.Remainders[j] {
				t.Errorf("[%v] want Q[%v].Remainders[%v] = 0x%X, got 0x%X",
					i, b, j, remainders, Q.Remainders[j])
			}
		}
	}
}

func Test_Put_in_same_block_without_run(t *testing.T) {
	t.Parallel()
	td := [][]uint64{
		// h0,   h1,        Q.occupieds,
		// Q[0].Remainders[0], Q[0].Remainders[1], Q[0].Remainders[2],
		// Q[0].Remainders[3], Q[0].Remainders[4], Q[0].Remainders[5],
		// Q[0].Remainders[6], Q[0].Remainders[7], Q[0].Remainders[8]
		// b

		// 0 - span 1st and 2nd r-bit cell
		{0x07, 0x1FF, 0x0000000000000080,
			0x8000000000000000, 0x00000000000000FF, 0x0000000000000000,
			0x0000000000000000, 0x0000000000000000, 0x0000000000000000,
			0x0000000000000000, 0x0000000000000000, 0x0000000000000000,
			0},
		// 1 - span 2nd and 3rd r-bit cell
		{0x0E, 0x1FF, 0x0000000000004000,
			0x0000000000000000, 0xC000000000000000, 0x000000000000007F,
			0x0000000000000000, 0x0000000000000000, 0x0000000000000000,
			0x0000000000000000, 0x0000000000000000, 0x0000000000000000,
			0},
		// 2 - span 3rd and 4th r-bit cell
		{0x15, 0x1FF, 0x0000000000200000,
			0x0000000000000000, 0x0000000000000000, 0xE000000000000000,
			0x000000000000003F, 0x0000000000000000, 0x0000000000000000,
			0x0000000000000000, 0x0000000000000000, 0x0000000000000000,
			0},
		// 3 - span 4th and 5th r-bit cell
		{0x1C, 0x1FF, 0x0000000010000000,
			0x0000000000000000, 0x0000000000000000, 0x0000000000000000,
			0xF000000000000000, 0x000000000000001F, 0x0000000000000000,
			0x0000000000000000, 0x0000000000000000, 0x0000000000000000,
			0},
		// 4 - span 5th and 6th r-bit cell
		{0x23, 0x1FF, 0x0000000800000000,
			0x0000000000000000, 0x0000000000000000, 0x0000000000000000,
			0x0000000000000000, 0xF800000000000000, 0x000000000000000F,
			0x0000000000000000, 0x0000000000000000, 0x0000000000000000,
			0},
		// 5 - span 6th and 7th r-bit cell
		{0x2A, 0x1FF, 0x0000040000000000,
			0x0000000000000000, 0x0000000000000000, 0x0000000000000000,
			0x0000000000000000, 0x0000000000000000, 0xFC00000000000000,
			0x0000000000000007, 0x0000000000000000, 0x0000000000000000,
			0},
		// 6 - span 7th and 8th r-bit cell
		{0x31, 0x1FF, 0x0002000000000000,
			0x0000000000000000, 0x0000000000000000, 0x0000000000000000,
			0x0000000000000000, 0x0000000000000000, 0x0000000000000000,
			0xFE00000000000000, 0x0000000000000003, 0x0000000000000000,
			0},
		// 7 - span 8th and 9th r-bit cell
		{0x38, 0x1FF, 0x0100000000000000,
			0x0000000000000000, 0x0000000000000000, 0x0000000000000000,
			0x0000000000000000, 0x0000000000000000, 0x0000000000000000,
			0x0000000000000000, 0xFF00000000000000, 0x0000000000000001,
			0},
		// 8 - last entry, last r-bit cell
		{0x3F, 0x1FF, 0x8000000000000000,
			0x0000000000000000, 0x0000000000000000, 0x0000000000000000,
			0x0000000000000000, 0x0000000000000000, 0x0000000000000000,
			0x0000000000000000, 0x0000000000000000, 0xFF80000000000000,
			0},
		// 9 - first entry, first r-bit cell
		{0x00, 0x1FF, 0x0000000000000001,
			0x00000000000001FF, 0x0000000000000000, 0x0000000000000000,
			0x0000000000000000, 0x0000000000000000, 0x0000000000000000,
			0x0000000000000000, 0x0000000000000000, 0x0000000000000000,
			0},
		// 10 - second entry, first r-bit cell
		{0x01, 0x1FF, 0x0000000000000002,
			0x000000000003FE00, 0x0000000000000000, 0x0000000000000000,
			0x0000000000000000, 0x0000000000000000, 0x0000000000000000,
			0x0000000000000000, 0x0000000000000000, 0x0000000000000000,
			0},
		// 11 - last block, last r-bit cell, last entry
		{0x1FFFF, 0x1FF, 0x8000000000000000,
			0x0000000000000000, 0x0000000000000000, 0x0000000000000000,
			0x0000000000000000, 0x0000000000000000, 0x0000000000000000,
			0x0000000000000000, 0x0000000000000000, 0xFF80000000000000,
			2047},
	}

	for i, v := range td {
		f := New(100000)
		f.Put(v[0], v[1])
		b := v[12]
		Q := f.Q[b]

		for j := 0; j < 9; j++ {
			remainders := v[3+j]
			if remainders != Q.Remainders[j] {
				t.Errorf("[%v] want Q[%v].Remainders[%v] = 0x%X, got 0x%X",
					i, b, j, remainders, Q.Remainders[j])
			}
		}
	}
}
