package bloomfilter

import (
	"log"
	"math"
	"unsafe"
)

type bloomfilter struct {
	Capacity          uint64  // n
	FalsePositiveRate float64 // p
	NumHashes         uint64  // k
	BitSize           uint64  // m
	numBuckets        uint64
	state             []uint64
}



func New() *bloomfilter {
	//bitSize := uint64(math.Abs(math.Ceil(float64(capacity) * math.Log2(math.E) * math.Log2(1/float64(probability)))))
	//numHashes := uint64(math.Floor(float64(bitSize/capacity) * math.Log(2)))
	//bitSize := uint64(math.Ceil(-1 * float64(capacity) * math.Log(probability) / math.Pow(math.Log(2), 2)))
	//numHashes := uint64(math.Ceil(math.Log(2) * float64(bitSize) / float64(capacity)))
	//numBuckets := 1 << 26 / 64

	return &bloomfilter{
		Capacity:   10000000,
		NumHashes:  4,
		BitSize:    1 << 26,
		numBuckets: 1 << 20,
		state:      make([]uint64, 1<<20),
	}
}

func NewWithConfig(capacity uint64, probability float64) *bloomfilter {
	//bitSize := uint64(math.Abs(math.Ceil(float64(capacity) * math.Log2(math.E) * math.Log2(1/float64(probability)))))
	//numHashes := uint64(math.Floor(float64(bitSize/capacity) * math.Log(2)))
	bitSize := uint64(math.Ceil(-1 * float64(capacity) * math.Log(probability) / math.Pow(math.Log(2), 2)))
	numHashes := uint64(math.Ceil(math.Log(2) * float64(bitSize) / float64(capacity)))
	numBuckets := bitSize / 64
	log.Printf("bitsize: %d hashnumber: %d", bitSize, numHashes)
	return &bloomfilter{
		Capacity:          capacity,
		FalsePositiveRate: probability,
		NumHashes:         numHashes,
		BitSize:           bitSize,
		numBuckets:        numBuckets,
		state:             make([]uint64, numBuckets),
	}
}

func (bf *bloomfilter) setBit(index uint64) {
	bucket := (index / 64) % bf.numBuckets
	offset := index & (1<<6 - 1)
	bf.state[bucket] |= 1 << uint(offset)
}

func (bf *bloomfilter) testBit(index uint64) int {
	bucket := (index / 64) % bf.numBuckets
	offset := index & (1<<6 - 1)
	if bf.state[bucket]&(1<<uint(offset)) != 0 {
		return 1
	} else {
		return 0
	}
}

func stringStructOf(sp *string) *stringStruct {
	return (*stringStruct)(unsafe.Pointer(sp))
}



func (bf *bloomfilter) Add(input string) {
	for i := 0; i < int(bf.NumHashes); i++ {
		//index := uint64(memhash(noescape(unsafe.Pointer(&p)), uintptr(i), uintptr(p.len))) % bf.BitSize
		index := uint64(strhash(noescape(unsafe.Pointer(&input)), uintptr(i))) % bf.BitSize
		bf.setBit(index)
	}
}


func (bf *bloomfilter) Check(input string) bool {
	//p := stringStructOf(&input)
	for i := 0; i < int(bf.NumHashes); i++ {
		//index := uint64(memhash(noescape(unsafe.Pointer(&p)), uintptr(i), uintptr(p.len))) % bf.BitSize
		index := uint64(strhash(noescape(unsafe.Pointer(&input)), uintptr(i))) % bf.BitSize
		if bf.testBit(index) != 1 {
			return false
		}
	}
	return true
}


