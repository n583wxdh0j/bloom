package bloom

import (
	"crypto/sha256"
	"encoding/binary"
	"hash/crc64"
	"log"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"
)

const truePositiveProbability = 0.6185

type BloomFilter interface {
	Put([]byte)
	Check([]byte) bool
	Println(int)
}

type bloomFilter struct {
	sync.RWMutex
	mask         []byte
	hashFunction func([]byte, []byte) uint64
	salts        [][]byte
}

func NewBloomFilter(nBits uint64, nHashFunc int) BloomFilter {
	salts := make([][]byte, 0, nHashFunc)
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < nHashFunc; i++ {
		n := rand.Int()
		salts = append(salts, []byte(strconv.Itoa(n)))
	}
	nBytes := (nBits + 7) / 8
	return &bloomFilter{
		mask:         make([]byte, nBytes, nBytes),
		hashFunction: getHasherUsesCRC64(nBytes),
		salts:        salts,
	}
}

func (bf *bloomFilter) Put(item []byte) {
	bf.Lock()
	defer bf.Unlock()
	for _, salt := range bf.salts {
		hash := bf.hashFunction(salt, item)
		bf.mask[hash] |= 1 << (hash & 7)
	}
}

func (bf *bloomFilter) Check(item []byte) bool {
	bf.RLock()
	defer bf.RUnlock()
	for _, salt := range bf.salts {
		hash := bf.hashFunction(salt, item)
		res := bf.mask[hash] & (1 << (hash & 7))
		if res == 0 {
			return false
		}
	}
	return true
}

func (bf *bloomFilter) Println(lenOfLine int) {
	var truePositiveCounter int
	b := strings.Builder{}
	for key, _ := range bf.mask {
		if key%lenOfLine == 0 {
			b.WriteString("\n")
		}
		for i := uint8(0); i < 8; i++ {
			res := bf.mask[key] & (1 << i)
			if res == 0 {
				b.WriteString("□")
			} else {
				b.WriteString("▨")
				truePositiveCounter++
			}
		}
	}
	log.Printf("Count of true positive: %v", truePositiveCounter)
	log.Println(b.String())
}

func getHasherUsesStdSHA256(size uint64) func(salt []byte, item []byte) uint64 {
	return func(salt []byte, item []byte) uint64 {
		hasher := sha256.New()
		hasher.Write(salt)
		hasher.Write(item)
		hash := binary.BigEndian.Uint64(hasher.Sum(nil))
		return hash % size
	}
}

func getHasherUsesCRC64(size uint64) func(salt []byte, item []byte) uint64 {
	return func(salt []byte, item []byte) uint64 {
		ECMATable := crc64.MakeTable(crc64.ISO)
		item = append(item, salt...)
		hash := crc64.Checksum(item, ECMATable)
		return hash % size
	}
}

func CalcHashCountAndProbability(maskSize int, dataSize int) {
	hashCount := 0.6931 * (float64(maskSize) / float64(dataSize))
	probability := math.Pow(truePositiveProbability, float64(maskSize)/float64(dataSize)) * 100
	log.Printf("Hash count = %v, probability = %v", hashCount, probability)
}
