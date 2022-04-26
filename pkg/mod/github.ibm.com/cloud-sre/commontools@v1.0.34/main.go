package main

import (
	"fmt"
	"github.ibm.com/cloud-sre/commontools/queue"
)

// this main function is just for passing oss cicd pipeline
func main(){
	q := &queue.Queue{}
	q.Push(1)
	q.Push(2)
	q.Push(3)
	q.Push(4)
	q.Push(5)

	for !q.Empty() {
		fmt.Println(q.Pop())
	}
	fmt.Println(q.Front())
	fmt.Println(q.Rear())
	fmt.Println(q.Size())
	fmt.Println()
}

