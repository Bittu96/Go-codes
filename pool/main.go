package main

import "fmt"

func loadWorks(tasksCount int) (workChan, resultChan chan int) {
	workChan = make(chan int, tasksCount)
	resultChan = make(chan int, tasksCount)

	for i := range tasksCount {
		workChan <- i
	}
	return
}

func worker(workChan, resultChan chan int) {
	for task := range workChan {
		fmt.Println("received task", task)
		result := task * task * task
		resultChan <- result
	}
}

func startPool(poolSize int, workChan, resultChan chan int) {
	for range poolSize {
		go worker(workChan, resultChan)
	}
}

func main() {
	var (
		tasksCount = 10
		poolSize   = 100
	)

	workChan, resultChan := loadWorks(tasksCount)
	defer func() {
		close(workChan)
		close(resultChan)
	}()

	startPool(poolSize, workChan, resultChan)

	for range tasksCount {
		result := <-resultChan
		fmt.Println(result)
	}
}
