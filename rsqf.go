package rsqf

import (
	"errors"
	"hash/fnv"
	"math"
)

const errRate float64 = 1.0 / 512.0
const rSize = 9 // log2(1/ERR_RATE)
const blockLen = 64

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
	Remainders [rSize]uint64
}

var rankNibbleTable = [16]uint64{
	//0, 1, 2, 3, 4, 5, 6, 7, 8, 9, A, B, C, D, E, F
	0, 1, 1, 2, 1, 2, 2, 3, 1, 2, 2, 3, 2, 3, 3, 4,
}

const one uint64 = 0x1
const rankMask uint64 = 0x0F

// Rank returns the number of 1s in B up to position i.
func Rank(B, i uint64) uint64 {
	masked := B & ((one << (i + 1)) - 1)

	// TODO: Look into using SIMD for this junk or create a 1-byte table
	c0 := rankNibbleTable[masked&rankMask]
	masked = masked >> 4
	c1 := rankNibbleTable[masked&rankMask]
	masked = masked >> 4
	c2 := rankNibbleTable[masked&rankMask]
	masked = masked >> 4
	c3 := rankNibbleTable[masked&rankMask]

	masked = masked >> 4
	c4 := rankNibbleTable[masked&rankMask]
	masked = masked >> 4
	c5 := rankNibbleTable[masked&rankMask]
	masked = masked >> 4
	c6 := rankNibbleTable[masked&rankMask]
	masked = masked >> 4
	c7 := rankNibbleTable[masked&rankMask]

	masked = masked >> 4
	c8 := rankNibbleTable[masked&rankMask]
	masked = masked >> 4
	c9 := rankNibbleTable[masked&rankMask]
	masked = masked >> 4
	c10 := rankNibbleTable[masked&rankMask]
	masked = masked >> 4
	c11 := rankNibbleTable[masked&rankMask]

	masked = masked >> 4
	c12 := rankNibbleTable[masked&rankMask]
	masked = masked >> 4
	c13 := rankNibbleTable[masked&rankMask]
	masked = masked >> 4
	c14 := rankNibbleTable[masked&rankMask]
	masked = masked >> 4
	c15 := rankNibbleTable[masked&rankMask]

	return c0 + c1 + c2 + c3 + c4 + c5 + c6 + c7 +
		c8 + c9 + c10 + c11 + c12 + c13 + c14 + c15
}

// Select returns the index of the ith 1 in B.
// References:
//  https://graphics.stanford.edu/~seander/bithacks.html#CountBitsSetParallel
func Select(B, i uint64) uint64 {
	/*
	   uint64_t v;          // Input value to find position with rank r.
	   unsigned int r;      // Input: bit's desired rank [1-64].
	   unsigned int s;      // Output: Resulting position of bit with rank r [1-64]
	   uint64_t a, b, c, d; // Intermediate temporaries for bit count.
	   unsigned int t;      // Bit count temporary.

	   // Do a normal parallel bit count for a 64-bit integer,
	   // but store all intermediate steps.
	   a =  v - ((v >> 1) & ~0UL/3);
	   b = (a & ~0UL/5) + ((a >> 2) & ~0UL/5);
	   c = (b + (b >> 4)) & ~0UL/0x11;
	   d = (c + (c >> 8)) & ~0UL/0x101;
	   t = (d >> 32) + (d >> 48);

	   // Now do branchless select!
	   s  = 64;
	   s -= ((t - r) & 256) >> 3; r -= (t & ((t - r) >> 8));
	   t  = (d >> (s - 16)) & 0xff;
	   s -= ((t - r) & 256) >> 4; r -= (t & ((t - r) >> 8));
	   t  = (c >> (s - 8)) & 0xf;
	   s -= ((t - r) & 256) >> 5; r -= (t & ((t - r) >> 8));
	   t  = (b >> (s - 4)) & 0x7;
	   s -= ((t - r) & 256) >> 6; r -= (t & ((t - r) >> 8));
	   t  = (a >> (s - 2)) & 0x3;
	   s -= ((t - r) & 256) >> 7; r -= (t & ((t - r) >> 8));
	   t  = (v >> (s - 1)) & 0x1;
	   s -= ((t - r) & 256) >> 8;
	   s = 65 - s;
	*/
	var c uint64
	var j uint64
	for ; j < 64; j++ {
		c += (B >> j) & 0x1
		if i == c {
			return j
		}
	}
	return 64
}

// pow2 calculates 2^exp using the shift left operator.
func pow2(exp uint64) uint64 {
	var v uint64 = 1
	return v << exp
}

// New returns a new Rsqf with a fixed 1% error rate.
func New(n float64) *Rsqf {
	p := uint64(calcP(n, errRate))
	q := p - rSize
	pmask := pow2(p) - 1
	rmask := pow2(rSize) - 1
	qmask := pmask ^ rmask
	qlen := int(pow2(q) / 64)
	filter := &Rsqf{
		p:         p,
		remainder: rSize,
		rMask:     rmask,
		quotient:  q,
		qMask:     qmask,
		Q:         make([]block, qlen, qlen),
	}

	return filter
}

// Rsqf is the core datastructure for this filter. Might evolve to using
// a 64-bit array which expands the filters size to 3-bits + r per slot from
// 2.125 + r.
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
func (q *Rsqf) Hash(b []byte) uint64 {
	h := fnv.New64a()
	h.Sum(b)
	return h.Sum64()
}

/*
MayContain tests if the hash exists in this filter. False positives are possible
however false negatives cannot occur.

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

// ErrFilterOverflow is returned if an insert would result in an overflow within
// the filter.
var ErrFilterOverflow = errors.New("RSQF overflow")

/*
firstAvailableSlot finds the first available slot for the hash x in this filter.

func firstAvailableSlot(Q, x)
	r <- rank(Q.occupieds, x)
	s <- select(Q.runends, r)
	while x <= s do
		x <- s + 1
		r <- rank(Q.occupieds, x)
		s <- select(Q.runends, r)
	return x
*/
func (q *Rsqf) firstAvailableSlot(x uint64) (uint64, error) {
	bi := x / blockLen
	bpos := x % blockLen

	if bi >= uint64(len(q.Q)) {
		return 0, ErrFilterOverflow
	}

	r := Rank(q.Q[bi].Occupieds, bpos)
	s := Select(q.Q[bi].Runends, r)

	if r == 0 && s == 0 {
		return x, nil
	}

	for x <= s {
		x = s + 1
		bi = x / blockLen
		bpos = x % blockLen
		if bi >= uint64(len(q.Q)) {
			return 0, ErrFilterOverflow
		}
		r = Rank(q.Q[bi].Occupieds, bpos)
		s = Select(q.Q[bi].Runends, r)
	}
	return x, nil
}

/*
Insert places the hash x into the filter where space is available.

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
func (q *Rsqf) Insert(x uint64) {
	h0 := (x & q.qMask) >> q.remainder
	h1 := x & q.rMask

	bi := h0 / blockLen
	bpos := h0 % blockLen

	r := Rank(q.Q[bi].Occupieds, bpos)
	s := Select(q.Q[bi].Runends, r)

	if h0 > s || (r == 0 && s == 0) {
		var re uint64 = (0x01 << bpos)

		q.Put(h0, h1)
		q.Q[bi].Runends |= re
	} else {
		/*
			s += 1
			n := q.FirstAvailableSlot(x)
			for n > s {
				n := -1
			}
		*/
	}

	var o uint64 = (0x01 << bpos)
	q.Q[bi].Occupieds |= o
}

// Put treats the Remainders block as a block of memory.
func (q *Rsqf) Put(h0, h1 uint64) {
	// ~10ns/op... le sigh complexity for now I suppose.
	bi := h0 / blockLen
	bpos := h0 % blockLen

	block := &q.Q[bi]

	rpos := bpos * q.remainder
	ri := rpos / blockLen
	low := (h1 << (rpos % blockLen))
	block.Remainders[ri] |= low

	// remainder spans multiple blocks
	if rpos+q.remainder > (ri+1)*blockLen {
		ri2 := ri + 1
		high := h1 >> (blockLen - (rpos % blockLen))
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
	bi := h0 / blockLen
	bpos := h0 % blockLen

	block := &q.Q[bi]

	block.Remainders[0] |= (oot(h1&1) << bpos)
	block.Remainders[1] |= (oot(h1&2) << bpos)
	block.Remainders[2] |= (oot(h1&4) << bpos)
	block.Remainders[3] |= (oot(h1&8) << bpos)
	block.Remainders[4] |= (oot(h1&16) << bpos)
	block.Remainders[5] |= (oot(h1&32) << bpos)
	block.Remainders[6] |= (oot(h1&64) << bpos)
	block.Remainders[7] |= (oot(h1&128) << bpos)
	block.Remainders[8] |= (oot(h1&256) << bpos)
}
