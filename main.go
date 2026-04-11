package main

import "fmt"

func main() {
	var str string
	_, _ = fmt.Scanln(&str)
	fmt.Println("Hello, World!" + str)
}
