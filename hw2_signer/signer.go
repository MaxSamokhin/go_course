package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
)

func ExecutePipeline(jobs ...job) {

	wg := &sync.WaitGroup{}
	defer wg.Wait()

	in := make(chan interface{}, 1)

	for _, currentJob := range jobs {
		wg.Add(1)
		out := make(chan interface{}, 1)

		go func(currentJob job, in, out chan interface{}) {
			defer wg.Done()
			defer close(out)

			currentJob(in, out)
		}(currentJob, in, out)

		in = out
	}
}

func CombineResults(in, out chan interface{}) {
	var arr []string

	for elem := range in {
		stringElem, ok := elem.(string)
		if !ok {
			fmt.Println("[ERROR]: failed convert data to int")
			return
		}

		arr = append(arr, stringElem)
	}

	sort.Strings(arr)
	out <- strings.Join(arr, "_")
}

func getSingleHash(data string, wg *sync.WaitGroup, wgInside *sync.WaitGroup, mu *sync.Mutex, out chan interface{}) {
	defer wg.Done()
	fmt.Println(data + " SingleHash data " + data)
	var dataCrc32Md5 string

	wgInside.Add(1)

	go func(data string) {
		defer wgInside.Done()

		mu.Lock()
		dataMd5 := DataSignerMd5(data)
		mu.Unlock()
		fmt.Println(data + " SingleHash md5(data) " + dataMd5)
		dataCrc32Md5 = DataSignerCrc32(dataMd5)
		fmt.Println(data + " SingleHash crc32(md5(data)) " + dataCrc32Md5)

	}(data)

	dataCrc32 := DataSignerCrc32(data)
	fmt.Println(data + " SingleHash crc32(data) " + dataCrc32)

	wgInside.Wait()

	dataResult := dataCrc32 + "~" + dataCrc32Md5
	fmt.Println(data + " SingleHash result " + dataResult)

	out <- dataResult
}

func SingleHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	mu := &sync.Mutex{}

	for input := range in {

		wg.Add(1)
		numberInput, ok := input.(int)
		if !ok {
			fmt.Println("[ERROR]: failed convert data to int")
			return
		}

		wgInside := &sync.WaitGroup{}
		go getSingleHash(strconv.Itoa(numberInput), wg, wgInside, mu, out)
	}

	wg.Wait()
}

func getMultiHash(data string, wg *sync.WaitGroup, out chan interface{}) {
	defer wg.Done()

	wgInside := &sync.WaitGroup{}
	var arrayHash = make([]string, 6, 6)

	for i := 0; i < 6; i++ {
		wgInside.Add(1)

		go func(i int) {
			defer wgInside.Done()
			crcParam := DataSignerCrc32(strconv.Itoa(i) + data)
			arrayHash[i] = crcParam
		}(i)
	}

	wgInside.Wait()

	result := strings.Join(arrayHash, "")
	fmt.Println(data + " MultiHash: crc32(th+step1)) " + result)

	out <- result
}

func MultiHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	for input := range in {

		inputString, ok := input.(string)
		if !ok {
			fmt.Println("[ERROR]: failed convert data to int")
			return
		}
		wg.Add(1)

		go getMultiHash(inputString, wg, out)
	}

	wg.Wait()
}
