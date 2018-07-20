package main

import (
	"crypto/md5"
	"fmt"
	"hash/crc32"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

type job func(in, out chan interface{})

const (
	MaxInputDataLen = 100
)

var (
	dataSignerOverheat uint32 = 0
	DataSignerSalt            = ""
)

var OverheatLock = func() {
	for {
		if swapped := atomic.CompareAndSwapUint32(&dataSignerOverheat, 0, 1); !swapped {
			fmt.Println("OverheatLock happend")
			time.Sleep(time.Second)
		} else {
			break
		}
	}
}

var OverheatUnlock = func() {
	for {
		if swapped := atomic.CompareAndSwapUint32(&dataSignerOverheat, 1, 0); !swapped {
			fmt.Println("OverheatUnlock happend")
			time.Sleep(time.Second)
		} else {
			break
		}
	}
}

var DataSignerMd5 = func(data string) string {
	OverheatLock()
	defer OverheatUnlock()
	data += DataSignerSalt
	dataHash := fmt.Sprintf("%x", md5.Sum([]byte(data)))
	time.Sleep(10 * time.Millisecond)
	return dataHash
}

var DataSignerCrc32 = func(data string) string {
	data += DataSignerSalt
	crcH := crc32.ChecksumIEEE([]byte(data))
	dataHash := strconv.FormatUint(uint64(crcH), 10)
	time.Sleep(time.Second)
	return dataHash
}

func ExecutePipeline() {
	singleHash1 := SingleHash("0")
	multiHash1 := MultiHash(singleHash1)

	singleHash2 := SingleHash("1")
	multiHash2 := MultiHash(singleHash2)

	CombineResults(multiHash1, multiHash2)
}

func CombineResults(a string, b string) {
	var arr []string
	arr = append(arr, a, b)
	sort.Strings(arr)

	fmt.Println("CombineResults " + strings.Join(arr, "_"))
}

func SingleHash(data string) string {
	fmt.Println(data + " SingleHash data " + data)

	dataMd5 := DataSignerMd5(data)
	fmt.Println(data + " SingleHash md5(data) " + dataMd5)

	dataCrc32Md5 := DataSignerCrc32(dataMd5)
	fmt.Println(data + " SingleHash crc32(md5(data)) " + dataCrc32Md5)

	dataCrc32 := DataSignerCrc32(data)
	fmt.Println(data + " SingleHash crc32(data) " + dataCrc32)

	dataResult := dataCrc32 + "~" + dataCrc32Md5
	fmt.Println(data + " SingleHash result " + dataResult)

	return dataResult;
}

func MultiHash(data string) string {
	var arrayHash []string

	for i := 0; i < 6; i++ {
		newData := strconv.Itoa(i) + data
		crcParam := DataSignerCrc32(newData)

		arrayHash = append(arrayHash, crcParam)
		fmt.Println(data + " MultiHash: crc32(th+step1)) " + strconv.Itoa(i) + " " + crcParam)
	}
	result := strings.Join(arrayHash, "")
	fmt.Println(data + " MultiHash result " + result)

	return strings.Join(arrayHash, "")
}

func main() {
	ExecutePipeline()
}
