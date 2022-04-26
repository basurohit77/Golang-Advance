package main

import (
	"fmt"
	"time"
)

//var wg sync.WaitGroup

func main() {
	resultCh := make(chan int)
	fmt.Println("main started")
	//wg.Add(1)
	go add(100, 200, resultCh)
	fmt.Println("Initiated the add operation")
	result := 0
	fmt.Println("result = ", result)
	//wg.Wait()
	fmt.Println("main completed")
}

func add(x, y int, resultCh chan int) {
	time.Sleep(4 * time.Second)
	result := x + y
	resultCh <- result
	//wg.Done()
}
