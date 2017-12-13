package rsqf

import (
	"hash/fnv"
	"math"
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
	// round up to ensure size is allocated correctly for target error rate.
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
	q.Put(sum.h0, sum.h1)
}

// Put treats the Remainders block as a block of memory.
func (q *Rsqf) Put(h0, h1 uint64) {
	// ~10ns/op... le sigh complexity for now I suppose.
	bi := h0 / BLOCK_LEN
	bpos := h0 % BLOCK_LEN

	block := &q.Q[bi]

	var o uint64 = (0x01 << bpos)
	block.Occupieds |= o

	var re uint64 = (0x01 << bpos)
	block.Runends |= re

	rpos := bpos * q.remainder
	ri := rpos / BLOCK_LEN
	low := (h1 << (rpos % BLOCK_LEN))
	block.Remainders[ri] |= low

	// remainder spans multiple blocks
	if rpos+q.remainder > (ri+1)*BLOCK_LEN {
		ri2 := ri + 1
		high := h1 >> (BLOCK_LEN - (rpos % BLOCK_LEN))
		block.Remainders[ri2] |= high
	}
}

func oot(v uint64) uint64 {
	if 0 == v {
		return 0
	}
	return 1
}

// Put2 treats each row in the Remainders block as a bit field
// for the associated bit position in a given the remainder.
func (q *Rsqf) Put2(h0, h1 uint64) {
	// ~16ns/op sadly 6ns slower than Put()
	bi := h0 / BLOCK_LEN
	bpos := h0 % BLOCK_LEN

	block := &q.Q[bi]

	var o uint64 = (0x01 << bpos)
	block.Occupieds |= o

	var re uint64 = (0x01 << bpos)
	block.Runends |= re

	r := &block.Remainders

	r[0] |= (oot(h1&1) << bpos)
	r[1] |= (oot(h1&2) << bpos)
	r[2] |= (oot(h1&4) << bpos)
	r[3] |= (oot(h1&8) << bpos)
	r[4] |= (oot(h1&16) << bpos)
	r[5] |= (oot(h1&32) << bpos)
	r[6] |= (oot(h1&64) << bpos)
	r[7] |= (oot(h1&128) << bpos)
	r[8] |= (oot(h1&256) << bpos)
}
