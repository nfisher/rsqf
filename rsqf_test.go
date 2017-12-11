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
	Remainder  [R_SIZE]uint64
	Remainders uint64
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

// New returns a new CountingQuotientFilter with a fixed 1% error rate.
func New(n float64) *CountingQuotientFilter {
	p := uint64(calcP(n, ERR_RATE))
	q := p - R_SIZE
	pmask := pow2(p) - 1
	rmask := pow2(R_SIZE) - 1
	qmask := pmask ^ rmask
	qlen := int(pow2(q) / 64)
	filter := &CountingQuotientFilter{
		p:         p,
		remainder: R_SIZE,
		rMask:     rmask,
		quotient:  q,
		qMask:     qmask,
		Q:         make([]block, qlen, qlen),
	}

	return filter
}

type CountingQuotientFilter struct {
	p         uint64 // number of bits required to achieve the target error rate.
	quotient  uint64 // number of bits that belong to the quotient.
	qMask     uint64 // used to mask h0 bits of the hash.
	remainder uint64 // number of bits that belong to the remainder.
	rMask     uint64 // used to mask h1 bits of the hash.
	Q         []block
}

// Hash applies a 64-bit hashing algorithm to b and then splits the
// result into h0 and h1. Shifting h0 to the right by the remainder size.
func (q *CountingQuotientFilter) Hash(b []byte) sum {
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
func (q *CountingQuotientFilter) MayContain(x []byte) {
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
func (q *CountingQuotientFilter) FirstAvailableSlot(x []byte) {
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
func (q *CountingQuotientFilter) Insert(x []byte) {
	sum := q.Hash(x)
	q.put(sum)
}

func (q *CountingQuotientFilter) put(s sum) {
	var h0 uint64 = s.h0
	var h1 uint64 = s.h1
	idx := h0 / BLOCK_LEN
	pos := h0 % BLOCK_LEN
	var o uint64 = 0x01
	var r1 uint64 = h1
	var r2 uint64 = 0

	block := &q.Q[idx]

	if pos < BLOCK_LEN-q.remainder {
		// remaindeer fits in same block
		sl := (BLOCK_LEN - pos - q.remainder)
		r1 = h1 << sl
		o = o << (sl + q.remainder - 1)

		block.Remainders = block.Remainders | r1
		block.Occupieds = block.Occupieds | o
	} else if pos > BLOCK_LEN-q.remainder {
		// overflows into next block
		sr := (q.remainder - (BLOCK_LEN - pos))
		r1 = h1 >> sr // bits drop bc uint doesn't wrap
		r2 = h1 << (pos + sr)
		o = o << (BLOCK_LEN - pos - 1)

		block.Remainders = block.Remainders | r1
		block.Occupieds = block.Occupieds | o

		block2 := &q.Q[idx+1]
		block2.Remainders = block2.Remainders | r2
	} else {
		// last position in block
		o = o << (q.remainder - 1)

		block.Remainders = block.Remainders | r1
		block.Occupieds = block.Occupieds | o
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

func Test_put_within_same_block(t *testing.T) {
	t.Skip()
	td := [][]uint64{
		// h0,   h1, Q[0].Remainder		 , Q[1].Remainders   , Q[1].Occupieds
		// first block, shift left
		{0x00, 0x07, 0xE000000000000000, 0x0000000000000000, 0x8000000000000000},
		// first block, do nothing
		{0x3D, 0x07, 0x0000000000000007, 0x0000000000000000, 0x0000000000000004},
		// first block, shift right to second cell
		{0x3E, 0x07, 0x0000000000000003, 0x8000000000000000, 0x0000000000000002},
	}

	for i, v := range td {
		f := New(100000)
		s := sum{v[0], v[1]}
		f.put(s)

		if v[2] != f.Q[0].Remainders {
			t.Errorf("[%v] want Q[0].Remainders = 0x%X, got 0x%X", i, v[2], f.Q[0].Remainders)
		}

		if v[3] != f.Q[1].Remainders {
			t.Errorf("[%v] want Q[1].Remainders = 0x%X, got 0x%X", i, v[3], f.Q[1].Remainders)
		}

		if v[4] != f.Q[0].Occupieds {
			t.Errorf("[%v] want Q[0].Occupieds = 0x%X, got 0x%X", i, v[4], f.Q[0].Occupieds)
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
	// 1 + 10 * 8
	if sz != 0x68 {
		t.Errorf("sz = 0x%X\n0x%X\n0x%X", sz, p0, p1)
	}
}

func Benchmark_hashing(b *testing.B) {
	f := New(10000000)
	str := []byte("executed by the go test command when its -bench flag is provided")
	for i := 0; i < b.N; i++ {
		f.Hash(str)
	}
}
