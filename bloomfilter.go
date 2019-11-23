package bloomfilter

import (
	"log"
	"math"
	"unsafe"
)

type bloomfilter struct {
	cap     uint64  // n
	fprate  float64 // p
	nhash   uint64  // k
	bitsize uint64  // m
	nbucket uint64
	bmap    []uint64
}

func New() *bloomfilter {
	return &bloomfilter{
		cap:     10000000,
		nhash:   4,
		bitsize: 1 << 26,
		nbucket: 1 << 20,
		bmap:    make([]uint64, 1<<20),
	}
}

func NewWithConfig(cap uint64, fp float64) *bloomfilter {
	//bitsize := uint64(math.Abs(math.Ceil(float64(cap) * math.Log2(math.E) * math.Log2(1/float64(probability)))))
	//nhash := uint64(math.Floor(float64(bitsize/cap) * math.Log(2)))
	bitsize := uint64(math.Ceil(-1 * float64(cap) * math.Log(fp) / math.Pow(math.Log(2), 2)))
	nhash := uint64(math.Ceil(math.Log(2) * float64(bitsize) / float64(cap)))
	nbucket := bitsize / 64
	log.Printf("bitsize: %d hashnumber: %d", bitsize, nhash)
	return &bloomfilter{
		cap:     cap,
		fprate:  fp,
		nhash:   nhash,
		bitsize: bitsize,
		nbucket: nbucket,
		bmap:    make([]uint64, nbucket),
	}
}

func (bf *bloomfilter) setBit(index uint64) {
	bucket := (index / 64) % bf.nbucket
	offset := index & (1<<6 - 1)
	bf.bmap[bucket] |= 1 << uint(offset)
}

func (bf *bloomfilter) testBit(index uint64) int {
	bucket := (index / 64) % bf.nbucket
	offset := index & (1<<6 - 1)
	if bf.bmap[bucket]&(1<<uint(offset)) != 0 {
		return 1
	} else {
		return 0
	}
}

func (bf *bloomfilter) Add(input string) {
	fn := strhash
	for i := 0; i < int(bf.nhash); i++ {
		index := uint64(fn(noescape(unsafe.Pointer(&input)), uintptr(i))) % bf.bitsize
		bf.setBit(index)
	}
}

func (bf *bloomfilter) Check(input string) bool {
	for i := 0; i < int(bf.nhash); i++ {
		index := uint64(strhash(noescape(unsafe.Pointer(&input)), uintptr(i))) % bf.bitsize
		if bf.testBit(index) != 1 {
			return false
		}
	}
	return true
}
