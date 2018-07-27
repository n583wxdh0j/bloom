package bloom

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
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
	String()
}

type bloomFilter struct {
	sync.RWMutex
	mask         []byte
	hashFunction func([]byte, []byte) uint64
	salts        [][]byte
}

func NewBloomFilter(size uint64, saltsAndHashFuncCount int) BloomFilter {
	salts := make([][]byte, 0, saltsAndHashFuncCount)
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < saltsAndHashFuncCount; i++ {
		n := rand.Int()
		salts = append(salts, []byte(strconv.Itoa(n)))
	}
	return &bloomFilter{
		mask:         make([]byte, size, size),
		hashFunction: getHasherUsesCRC64(size),
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
		hashNumber := bf.hashFunction(salt, item)
		res := bf.mask[hashNumber] & (1 << hashNumber)
		if res == 0 {
			return false
		}
	}
	return true
}

func (bf *bloomFilter) String() {
	var howManyTrue int
	b := strings.Builder{}
	for key, _ := range bf.mask {
		if key%200 == 0 {
			b.WriteString("\n")
		}
		bit := bf.mask[key]
		if bit > 0 {
			b.WriteString("▨")
			howManyTrue++
		} else {
			b.WriteString("□")
		}
	}
	log.Printf("Count of true positive: %v", howManyTrue)
	log.Println(b.String())
}

func getHasherUsesStdSHA256(size uint64) func(salt []byte, item []byte) uint64 {
	return func(salt []byte, item []byte) uint64 {
		hash := sha256.New()
		hash.Write(salt)
		hash.Write(item)
		data := binary.BigEndian.Uint64(hash.Sum(nil))
		return data % size
	}
}

func getHasherUsesCRC64(size uint64) func(salt []byte, item []byte) uint64 {
	return func(salt []byte, item []byte) uint64 {
		ECMATable := crc64.MakeTable(crc64.ISO)
		saltBinary := salt
		item = append(item, saltBinary...)
		data := crc64.Checksum(item, ECMATable)
		return data % size
	}
}

func CalcHshCountAndProbability(maskSize int, dataSize int) {
	hashCount := 0.6931 * (float64(maskSize) / float64(dataSize))
	probability := math.Pow(truePositiveProbability, float64(maskSize)/float64(dataSize)) * 100
	fmt.Printf("Hash count = %v, probability = %v", hashCount, probability)
}