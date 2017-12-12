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
	"hash/fnv"
	"math"
	"testing"
	"unsafe"
)

const ERR_RATE float64 = 1.0 / 512.0
const R_SIZE = 9 // log2(1/ERR_RATE)
const BLOCK_LEN = 64

// calcP calculates the p exponent for the universe.
// n is the maximum number of insertions.
// errRate is the desired error rate.
//
// could error if exceeds 2^52 for float64 fractional length but yer fooked
// for in memory usage anyway unless ya got PB of memory lying around.
func calcP(n, errRate float64) float64 {
	// round up to ensure vector size is allocated correctly for target error rate.
	return math.Ceil(math.Log2(n / errRate))
}

// sum holds the split for the hash function into h0 and h1.
type sum struct {
	h0 uint64 // rotated right by r
	h1 uint64
}

// block is the backing bitmap store for the filter.
type block struct {
	Offset     uint8
	Occupieds  uint64
	Runends    uint64
	Remainders [R_SIZE]uint64
}

// rank returns the number of 1s in Q.occupieds up to position i.
func rank(Q []block, i int) {
}

// seleccionar returns the index of the ith 1 in Q.runends.
func seleccionar(Q []block, i int) {
}

// pow2 calculates 2^exp using the shift left operator.
func pow2(exp uint64) uint64 {
	var v uint64 = 1
	return v << exp
}

// New returns a new Rsqf with a fixed 1% error rate.
func New(n float64) *Rsqf {
	p := uint64(calcP(n, ERR_RATE))
	q := p - R_SIZE
	pmask := pow2(p) - 1
	rmask := pow2(R_SIZE) - 1
	qmask := pmask ^ rmask
	qlen := int(pow2(q) / 64)
	filter := &Rsqf{
		p:         p,
		remainder: R_SIZE,
		rMask:     rmask,
		quotient:  q,
		qMask:     qmask,
		Q:         make([]block, qlen, qlen),
	}

	return filter
}

type Rsqf struct {
	p         uint64 // number of bits required to achieve the target error rate.
	quotient  uint64 // number of bits that belong to the quotient.
	qMask     uint64 // used to mask h0 bits of the hash.
	remainder uint64 // number of bits that belong to the remainder.
	rMask     uint64 // used to mask h1 bits of the hash.
	Q         []block
}

// Hash applies a 64-bit hashing algorithm to b and then splits the
// result into h0 and h1. Shifting h0 to the right by the remainder size.
func (q *Rsqf) Hash(b []byte) sum {
	h := fnv.New64a()
	h.Sum(b)
	res := h.Sum64()
	return sum{
		h0: (res & q.qMask) >> q.remainder,
		h1: res & q.rMask,
	}
}

/*
func MayContain(Q, x)
	b <- h0(x)
	if Q.occupieds[b] = 0 then
		return 0
	t <- rank(Q.occupieds, b)
	l <- select(Q.runends, t)
	v <- h1(x)
	repeat
		if Q.remainders[l] == v then
			return 1
		l <- l - 1
	until l < b or Q.runends[l] = 1
  return false
*/
func (q *Rsqf) MayContain(x []byte) {
	//b := q.Hash(x)

}

/*
func FirstAvailableSlot(Q, x)
	r <- rank(Q.occupieds, x)
	s <- select(Q.runends, r)
	while x <= s do
		x <- s + 1
		r <- rank(Q.occupieds, b)
		s <- select(Q.runends, t)
	return x
*/
func (q *Rsqf) FirstAvailableSlot(x []byte) {
}

/*
func Insert(Q, x)
	r <- rank(Q.occupieds, b)
	s <- select(Q.runends, t)
	if h0(x) > s then
		Q.remainders[h0(x)] <- h1(x)
		Q.runends[h0(x)] <- 1
	else
		s <- s + 1
		n <- FirstAvailableSlot(Q, x)
		while n > s do
			Q.remainders[n] <- Q.remainders[n - 1]
			Q.runends[n] <- Q.runends[n - 1]
			n <- n - 1
		Q.remainders[s] <- h1(x)
		if Q.occupieds[h0(x)] == 1 then
			Q.runends[s - 1] <- 0
		Q.runends[s] <- 1
	Q.occupieds[h0(x)] <- 1
	return
*/
func (q *Rsqf) Insert(x []byte) {
	sum := q.Hash(x)
	q.put(sum)
}

func (q *Rsqf) put(s sum) {
	h0 := s.h0
	bi := h0 / BLOCK_LEN

	block := &q.Q[bi]

	bpos := h0 % BLOCK_LEN

	var o uint64 = (0x01 << bpos)
	block.Occupieds |= o

	var re uint64 = (0x01 << bpos)
	block.Runends |= re

	rpos := bpos * q.remainder
	ri := rpos / BLOCK_LEN
	low := (s.h1 << (rpos % BLOCK_LEN))
	block.Remainders[ri] |= low

	// remainder spans multiple blocks
	if rpos+q.remainder > (ri+1)*BLOCK_LEN {
		ri2 := ri + 1
		high := s.h1 >> (BLOCK_LEN - (rpos % BLOCK_LEN))
		block.Remainders[ri2] |= high
	}
}

// =============== tests

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

func Test_put_within_same_block_without_run(t *testing.T) {
	td := [][]uint64{
		// h0,   h1,        Q.occupieds,
		//[0],	[1],				        [2],
		// Q[0].Remainders[0], Q[0].Remainders[1], Q[0].Remainders[2],
		// Q[0].Remainders[3], Q[0].Remainders[4], Q[0].Remainders[5],
		// Q[0].Remainders[6], Q[0].Remainders[7], Q[0].Remainders[8]
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
		s := sum{v[0], v[1]}
		f.put(s)
		b := v[12]
		Q := f.Q[b]

		occupieds := v[2]
		runends := occupieds // no runs so should be equal

		if occupieds != Q.Occupieds {
			t.Errorf("[%v] want Q[%v].Occupieds = 0x%X, got 0x%X", i, b, occupieds, Q.Occupieds)
		}

		if runends != Q.Runends {
			t.Errorf("[%v] want Q[%v].Runends = 0x%X, got 0x%X", i, b, runends, Q.Runends)
		}

		for j := 0; j < 9; j++ {
			remainders := v[3+j]
			if remainders != Q.Remainders[j] {
				t.Errorf("[%v] want Q[%v].Remainders[%v] = 0x%X, got 0x%X",
					i, b, j, remainders, Q.Remainders[j])
			}
		}
	}
}

func test_sample_p_values(t *testing.T) {
	p := calcP(100000, 0.05)
	if 0.1 != p {
		t.Errorf("%v", p)
	}
}

func Benchmark_init(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = New(1000000)
	}
}

func Test_packing_distance(t *testing.T) {
	f := New(10000)
	p0 := unsafe.Pointer(&f.Q[3])
	p1 := unsafe.Pointer(&f.Q[4])
	sz := unsafe.Sizeof(block{})
	// 1 + 11 * 8
	if sz != 0x60 {
		t.Errorf("got sz = 0x%X, want 0x60\n0x%X\n0x%X", sz, p0, p1)
	}
}

func Benchmark_hashing(b *testing.B) {
	f := New(10000000)
	str := []byte("executed by the go test command when its -bench flag is provided")
	for i := 0; i < b.N; i++ {
		f.Hash(str)
	}
}
