package main

import "fmt"

func main() {
	n := 9
	s := make(chan bool)
	e := make(chan bool)

	defer close(s)
	defer close(e)

	go printEven(n, s, e)
	s <- true
	go printOdd(n, s, e)

	<-e
}

func printEven(n int, s, e chan bool) {
	for i := range n {
		if i%2 == 0 {
			<-s
			fmt.Println(i)
			s <- true
		}
	}
	e <- true
}

func printOdd(n int, s, e chan bool) {
	for i := range n {
		if i%2 != 0 {
			<-s
			fmt.Println(i)
			s <- true
		}
	}
	e <- true
}
