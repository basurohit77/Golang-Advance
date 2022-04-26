package main

import (
	"fmt"
	"strings"
)

func main() {
	for {
		var input string
		fmt.Print(`Type "continue" to proceed to writing the output, or "stop" to abort: `)
		fmt.Scanln(&input)
		fmt.Printf("Got input \"%s\"\n", input)
		input = strings.TrimSpace(input)
		switch input {
		case "continue":
			fmt.Println("Continuing")
			return
		case "stop":
			fmt.Println("Stopping")
			return
		}
	}
}
