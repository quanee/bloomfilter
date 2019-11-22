package bloomfilter

import (
	"math"
	"unsafe"
)

const (
	// Constants for multiplication: four random odd 64-bit numbers.
	m1 = 16877499708836156737
	m2 = 2820277070424839065
	m3 = 9497967016996688599
	m4 = 15839092249703872147
)
const BigEndian = false

var hashkey [4]uintptr

type BloomFilter struct {
	Capacity          uint64  // n
	FalsePositiveRate float64 // p
	NumHashes         uint64  // k
	BitSize           uint64  // m
	numBuckets        uint64
	state             []uint64
}

func New(capacity uint64, probability float64) *BloomFilter {
	//bitSize := uint64(math.Abs(math.Ceil(float64(capacity) * math.Log2(math.E) * math.Log2(1/float64(probability)))))
	//numHashes := uint64(math.Floor(float64(bitSize/capacity) * math.Log(2)))
	bitSize := uint64(math.Ceil(-1 * float64(capacity) * math.Log(probability) / math.Pow(math.Log(2), 2)))
	numHashes := uint64(math.Ceil(math.Log(2) * float64(bitSize) / float64(capacity)))
	numBuckets := bitSize / 64

	return &BloomFilter{
		Capacity:          capacity,
		FalsePositiveRate: probability,
		NumHashes:         numHashes,
		BitSize:           bitSize,
		numBuckets:        numBuckets,
		state:             make([]uint64, uint(numBuckets)),
	}
}

func (bf *BloomFilter) setBit(index uint64) {
	bucket := (index / 64) % bf.numBuckets
	offset := index & (1<<6 - 1)
	bf.state[bucket] |= 1 << uint(offset)
}

func (bf *BloomFilter) testBit(index uint64) int {
	bucket := (index / 64) % bf.numBuckets
	offset := index & (1<<6 - 1)
	if bf.state[bucket]&(1<<uint(offset)) != 0 {
		return 1
	} else {
		return 0
	}
}

func (bf *BloomFilter) Add(input []byte) {
	for i := 0; i < int(bf.NumHashes); i++ {
		index := uint64(memhash32(unsafe.Pointer(&input), uintptr(i))) % bf.BitSize
		bf.setBit(index)
	}
}

func (bf *BloomFilter) Check(input []byte) bool {
	for i := 0; i < int(bf.NumHashes); i++ {
		index := uint64(memhash32(unsafe.Pointer(&input), uintptr(i))) % bf.BitSize
		if bf.testBit(index) != 1 {
			return false
		}
	}
	return true
}

func memhash(p unsafe.Pointer, seed, s uintptr) uintptr {
	h := uint64(seed + s*hashkey[0])
tail:
	switch {
	case s == 0:
	case s < 4:
		h ^= uint64(*(*byte)(p))
		h ^= uint64(*(*byte)(add(p, s>>1))) << 8
		h ^= uint64(*(*byte)(add(p, s-1))) << 16
		h = rotl_31(h*m1) * m2
	case s <= 8:
		h ^= uint64(readUnaligned32(p))
		h ^= uint64(readUnaligned32(add(p, s-4))) << 32
		h = rotl_31(h*m1) * m2
	case s <= 16:
		h ^= readUnaligned64(p)
		h = rotl_31(h*m1) * m2
		h ^= readUnaligned64(add(p, s-8))
		h = rotl_31(h*m1) * m2
	case s <= 32:
		h ^= readUnaligned64(p)
		h = rotl_31(h*m1) * m2
		h ^= readUnaligned64(add(p, 8))
		h = rotl_31(h*m1) * m2
		h ^= readUnaligned64(add(p, s-16))
		h = rotl_31(h*m1) * m2
		h ^= readUnaligned64(add(p, s-8))
		h = rotl_31(h*m1) * m2
	default:
		v1 := h
		v2 := uint64(seed * hashkey[1])
		v3 := uint64(seed * hashkey[2])
		v4 := uint64(seed * hashkey[3])
		for s >= 32 {
			v1 ^= readUnaligned64(p)
			v1 = rotl_31(v1*m1) * m2
			p = add(p, 8)
			v2 ^= readUnaligned64(p)
			v2 = rotl_31(v2*m2) * m3
			p = add(p, 8)
			v3 ^= readUnaligned64(p)
			v3 = rotl_31(v3*m3) * m4
			p = add(p, 8)
			v4 ^= readUnaligned64(p)
			v4 = rotl_31(v4*m4) * m1
			p = add(p, 8)
			s -= 32
		}
		h = v1 ^ v2 ^ v3 ^ v4
		goto tail
	}

	h ^= h >> 29
	h *= m3
	h ^= h >> 32
	return uintptr(h)
}

// Note: These routines perform the read with an native endianness.
func readUnaligned32(p unsafe.Pointer) uint32 {
	q := (*[4]byte)(p)
	if BigEndian {
		return uint32(q[3]) | uint32(q[2])<<8 | uint32(q[1])<<16 | uint32(q[0])<<24
	}
	return uint32(q[0]) | uint32(q[1])<<8 | uint32(q[2])<<16 | uint32(q[3])<<24
}

func readUnaligned64(p unsafe.Pointer) uint64 {
	q := (*[8]byte)(p)
	if BigEndian {
		return uint64(q[7]) | uint64(q[6])<<8 | uint64(q[5])<<16 | uint64(q[4])<<24 |
			uint64(q[3])<<32 | uint64(q[2])<<40 | uint64(q[1])<<48 | uint64(q[0])<<56
	}
	return uint64(q[0]) | uint64(q[1])<<8 | uint64(q[2])<<16 | uint64(q[3])<<24 | uint64(q[4])<<32 | uint64(q[5])<<40 | uint64(q[6])<<48 | uint64(q[7])<<56
}

func rotl_31(x uint64) uint64 {
	return (x << 31) | (x >> (64 - 31))
}

// Should be a built-in for unsafe.Pointer?
//go:nosplit
func add(p unsafe.Pointer, x uintptr) unsafe.Pointer {
	return unsafe.Pointer(uintptr(p) + x)
}

func memhash32(p unsafe.Pointer, seed uintptr) uintptr {
	h := uint64(seed + 4*hashkey[0])
	v := uint64(readUnaligned32(p))
	h ^= v
	h ^= v << 32
	h = rotl_31(h*m1) * m2
	h ^= h >> 29
	h *= m3
	h ^= h >> 32
	return uintptr(h)
}
