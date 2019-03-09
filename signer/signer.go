package main

import (
        "sync"
        "fmt"
        "strings"
        "sort"
)

func ExecutePipeline(allJobFunctions ...job) {
        wg := &sync.WaitGroup{}
        in := make(chan interface{})
        for _, currentJob := range allJobFunctions {
                out := make(chan interface{})

                wg.Add(1)
                go func(in, out chan interface{}, currentJob job, wg *sync.WaitGroup) {
                        defer wg.Done()
                        defer close(out)
                        currentJob(in, out)
                }(in, out, currentJob, wg)

                in = out
        }
        wg.Wait()
}

func SingleHash(in, out chan interface{}) {
        wg := &sync.WaitGroup{}
        for i := range in {
                md5Data := DataSignerMd5(fmt.Sprintf("%v", i))

                wg.Add(1)
                go func(out chan interface{}, wg *sync.WaitGroup, md5StringData, data string) {
                        defer wg.Done()
                        ch1 := make(chan interface{})
                        ch2 := make(chan interface{})

                        go func(ch1 chan interface{}, data string) {
                                ch1 <- DataSignerCrc32(data)
                        }(ch1, data)

                        go func(ch2 chan interface{}, data string) {
                                ch2 <- DataSignerCrc32(data)
                        }(ch2, md5StringData)

                        var hashResultSlice []string
                        crcData := fmt.Sprintf("%v", <-ch1)
                        crcMD5Data := fmt.Sprintf("%v", <-ch2)
                        hashResultSlice = append(hashResultSlice, crcData, crcMD5Data)
                        out <- strings.Join(hashResultSlice, "~")
                }(out, wg, md5Data, fmt.Sprintf("%v", i))
        }
        wg.Wait()
}

func MultiHash(in, out chan interface{}) {
        wg := &sync.WaitGroup{}
        for i := range in {

                wg.Add(1)
                go func(out chan interface{}, wg *sync.WaitGroup, inItem string) {
                        defer wg.Done()
                        wgTmp := &sync.WaitGroup{}
                        multiHashSlice := make([]string, 6)
                        for th := 0; th < 6; th++ {
                                thData := fmt.Sprintf("%v", th) + inItem

                                wgTmp.Add(1)
                                go func(wg *sync.WaitGroup, resultData string, index int, slice []string ) {
                                        defer wg.Done()
                                        slice[index] = DataSignerCrc32(resultData)
                                }(wgTmp, thData, th, multiHashSlice)

                        }
                        wgTmp.Wait()
                        out <- strings.Join(multiHashSlice, "")
                }(out, wg, fmt.Sprintf("%v", i))

        }
        wg.Wait()
}

func CombineResults(in, out chan interface{}) {
        var multiHashResultSlice []string
        for i := range in {
                multiHashResultSlice = append(multiHashResultSlice, fmt.Sprintf("%v", i))
        }
        sort.Strings(multiHashResultSlice)
        out <- strings.Join(multiHashResultSlice, "_")
}
