# commontools

a common tools collection for reuse


sample：

```
  import (
    fmt
    github.ibm.com/cloud-sre/commontools/queue
  )
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
  
 
  ```
