package bloomfilter

import (
	"runtime"
	"unsafe"
)

const (
	// Constants for multiplication: four random odd 64-bit numbers.
	m1 = 16877499708836156737
	m2 = 2820277070424839065
	m3 = 9497967016996688599
	m4 = 15839092249703872147
)
const bigEndian = false

var hashkey [4]uintptr

type stringStruct struct {
	str unsafe.Pointer
	len int
}

func init() {
	/*if (runtime.GOARCH == "386" || runtime.GOARCH == "amd64") &&
		runtime.GOOS != "nacl" &&
		cpu.X86.HasAES && // AESENC
		cpu.X86.HasSSSE3 && // PSHUFB
		cpu.X86.HasSSE41 { // PINSR{D,Q}
		initAlgAES()
		return
	}
	if GOARCH == "arm64" && cpu.ARM64.HasAES {
		initAlgAES()
		return
	}
	getRandomData((*[len(hashkey) * sys.PtrSize]byte)(unsafe.Pointer(&hashkey))[:])*/
	hashkey[0] |= 1 // make sure these numbers are odd
	hashkey[1] |= 1
	hashkey[2] |= 1
	hashkey[3] |= 1
}

func strhash(a unsafe.Pointer, h uintptr) uintptr {
	x := (*stringStruct)(a)
	return memhash(x.str, h, uintptr(x.len))
}

func memhash(p unsafe.Pointer, seed, s uintptr) uintptr {
	if (runtime.GOARCH == "amd64" || runtime.GOARCH == "arm64") &&
		runtime.GOOS != "nacl" /*&& useAeshash*/ {
		//return aeshash(p, seed, s)
	}
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
	//log.Printf("memhash hash value \t%x %v", h, s)
	return uintptr(h)
}

// Note: These routines perform the read with an native endianness.
func readUnaligned32(p unsafe.Pointer) uint32 {
	q := (*[4]byte)(p)
	if bigEndian {
		return uint32(q[3]) | uint32(q[2])<<8 | uint32(q[1])<<16 | uint32(q[0])<<24
	}
	return uint32(q[0]) | uint32(q[1])<<8 | uint32(q[2])<<16 | uint32(q[3])<<24
}

func readUnaligned64(p unsafe.Pointer) uint64 {
	q := (*[8]byte)(p)
	if bigEndian {
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


//go:nosplit
func noescape(p unsafe.Pointer) unsafe.Pointer {
	x := uintptr(p)
	return unsafe.Pointer(x ^ 0)
}

