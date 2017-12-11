# Rank and Seek Quotient Filter

The specification and analysis of RSQF can be found in the following paper:

 https://www3.cs.stonybrook.edu/~ppandey/files/p775-pandey.pdf

A Rank and Seek Quotient Filter (RSQF) is an Approximate Membership Query data structure. It is similar to the more popular Bloom Filter (BF) however
where a BF only provides insert, and lookup.

An RSQF provides:

 * insert
 * delete
 * lookup
 * resize
 * merge

## Glossary

This glossary summarises the variables specified in the [stonybrook paper](https://www3.cs.stonybrook.edu/~ppandey/files/p775-pandey.pdf).

**Note**: Where integers are specified, fractional values are rounded up to
the nearest integer when calculated from floating point values.

<dl>
<dt>n - (integer)</dt>
<dd>Maximum number of insertions (e.g. 1,000,000).</dd>

<dt>δ - (fraction)</dt>
<dd>Error rate or false-positive rate (e.g. 1/512 or 1/100).</dd>

<dt>p - (integer)</dt>
<dd>Number of bits required from the hashed input to achieve the target error
for the given number of insertions (n). The p-bit hash is split into high bits (quotient) and low bits (remainder).<br>
`p = log2(n/δ)`</dd>

<dt>r / remainder - (integer)</dt>
<dd>The number of remainder bits which are written to `Q.remainders`.<br>
`r = log2(1/δ)`</dd>

<dt>q / quotient - (integer)</dt>
<dd>The number of quotient bits used to indicate the expected `home slot` in
the filter.<br>
`q = p - r`</dd>

<dt>run</dt>
<dd>A run is a consecutive group of remainders where the quotient is
equal.<br>`h0(a) = h0(b) = h0(c)`</dd>

<dt>occupied - (bit)</dt>
<dd>A bit that indicates the position of a home slot for a given `run`.</dd>

<dt>runend - (bit)</dt>
<dd>A bit that indicates the end of a `run`.</dd>

<dt>Q - (struct)</dt>
<dd>The RSQF data structure which contains 2<sup>q</sup> r-bits of available
space allocated by a `block` array. The memory in bits required by the
struct can be calculated as follows:<br>
`2^q/64 * (8+64(r+2))`</dd>

<dt>Q.occupieds - (bit vector)</dt>
<dd></dd>

<dt>Q.runends - (bit vector)</dt>
<dd></dd>

<dt>Q.remainders - (bit vector)</dt>
<dd></dd>

<dt>block</dt>
<dd>A block is `64(r + 2) + 8` bit structure. It is composed of the
following fields:
<ul>
  <li>offset - 1 x 8-bit.
  <li>runends - 1 x 64-bit.
  <li>occupieds - 1 x 64-bit.
  <li>remainders - r x 64-bit.
</ul>
</dd>

<dt>home slot - (array index)</dt>
<dd>The home slot is the location where a remainder would be placed if h0(x)
is unoccupied by another value.<br>
`i = h0(x), Q[i].remainders == h1(x)`</dd>

<dt>slot i - (array index)</dt>

<dt>h(x) - (integer)</dt>
<dd>A universal hashing function. For this library FNV-1a (64-bit) was
employed as it is available in the standard library.</dd>

<dt>h0(x) / i - (integer)</dt>
<dd>The masked upper bits of the hash shifted right `r` times.<br>
`h0 = (h(x) >> r) & (2^q - 1)`</dd>

<dt>h1(x) - (integer)</dt>
<dd>The masked lower half of the hash.<br>
`h1 = h(x) & (2^r - 1)`</dd>

<dt>B - (bit vector)</dt>
<dd>Variable representing a bit-vector. Typically one of `Q.occupieds`,
`Q.runends`, or `Q.remainders`.</dd>

<dt>RANK(B, i) - (integer)</dt>
<dd>Rank returns the number of 1s in B up to position i.<br>
`RANK(B, i) = `</dd>

<dt>SELECT(B, i) - (integer)</dt>
<dd>Select returns the index of the i<sup>th</sup> 1 in B.<br>
`SELECT(B, i) = `</dd>

<dt>O<sub>i</sub> - (integer)</dt>
<dd>Is every 64<sup>th</sup> slot which is stored in `Q[i].offset` to save
space. The offset is calculated with the algorithm that follows.<br>
`Oi = SELECT(Q.runends, RANK(Q.occupieds, i)) - i`</dd>

<dt>O<sub>j</sub> - (integer)</dt>
<dd>O<sub>j</sub> is a derived intermediate slot value which is discovered
using the algorithm that follows.<br>
<pre>
```
i = h0(x)
Oi = SELECT(Q.runends, RANK(Q.occupieds, i)) - i
d = RANK(q.occupieds[i + 1, ..., j], j - i - 1)
t = SELECT(Q.runends[i + Oi + 1,...,2^q - 1], d)
Oj = i + Oi + t - j
```
</pre>
</dd>
</dl>
