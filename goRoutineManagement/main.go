package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

type WorkResponse struct {
	Result     int    `json:"result"`
	Status     string `json:"status"`
	StatusCode int    `json:"statusCode"`
}

var dummyFailureRecord = map[int]int{}

func worker(wg *sync.WaitGroup, mux *sync.Mutex, workerId int, tasksChnl chan int, tasksRecordChnl chan string, span *time.Duration) {
	log.Printf("worker %v started working..", workerId)
	for {
		select {
		case taskId := <-tasksChnl:
			//t := make(chan int)
			doTask(mux, taskId, workerId, tasksChnl, tasksRecordChnl, span)
		default:
			log.Printf("worker %v stopped due to no work..", workerId)
			wg.Done()
			return
		}
	}
	//for {
	//for taskId := range tasksChnl {
	//	log.Println(tasksChnl)
	//	doTask(mux, taskId, workerId, tasksChnl, tasksRecordChnl, span)
	//}
	//}

	//if len(tasksChnl) == 0 {
	//	switch "stop" {
	//	case "stop":
	//		log.Printf("worker %v stopped working due to no work..", workerId)
	//		break
	//	case "rest":
	//		log.Printf("worker %v stopped working due to no work..", workerId)
	//		break
	//	}
	//	wg.Done()
	//}
	//else {
	//	for taskId := range tasksChnl {
	//		log.Println(tasksChnl)
	//		doTask(mux, taskId, workerId, tasksChnl, tasksRecordChnl, span)
	//	}
	//}

	//log.Printf("worker %v stopped working due to no work..", workerId)
	//log.Printf("worker %v left..", workerId)
	//wg.Done()
}

type taskRecords struct {
	//gorm.Model
	WorkerId int           `json:"worker_id" gorm:"int(8)"`
	TaskId   int           `json:"task_id" gorm:"int(8)"`
	Score    int           `json:"score" gorm:"int(16)"`
	Duration time.Duration `json:"duration" gorm:"timestamp"`
	Status   string        `json:"status" gorm:"varchar(16)"`
	Message  string        `json:"message" gorm:"varchar(256)"`
}

func doTask(mux *sync.Mutex, taskId int, workerId int, tasksChnl chan int, tasksRecordChnl chan string, span *time.Duration) {
	log.Printf("worker %v picked task %v", workerId, taskId)
	startTime := time.Now()
	score, err := processTask(taskId, workerId)
	duration := time.Now().Sub(startTime)
	recordObj := taskRecords{WorkerId: workerId, TaskId: taskId, Score: score, Duration: duration, Status: "success", Message: "task completed"}
	record, _ := json.Marshal(recordObj)
	//record := fmt.Sprintf("task %v finished by worker %v in %v with a score of %v", taskId, workerId, duration, score)
	if err != nil {
		recordObj := taskRecords{WorkerId: workerId, TaskId: taskId, Score: score, Duration: duration, Status: "failure", Message: "connection failed"}
		record, _ = json.Marshal(recordObj)
		//record = fmt.Sprintf("task %v is not finished by worker %v due to connection issue", taskId, workerId)
		mux.Lock()
		if dummyFailureRecord[taskId] += 1; dummyFailureRecord[taskId] > 3 {
			recordObj := taskRecords{WorkerId: workerId, TaskId: taskId, Score: score, Duration: duration, Status: "success", Message: "task completed"}
			//recordObj := taskRecords{workerId, taskId, 0, duration, "failure", "connection failed, ending task"}
			record, _ = json.Marshal(recordObj)
			//record = fmt.Sprintf("task %v was not finished by worker %v due to connection issue : task %v withdrawn", taskId, workerId, taskId)
		} else {
			tasksChnl <- taskId
		}
		mux.Unlock()
	}
	tasksRecordChnl <- string(record)
	*span += duration
	log.Println(string(record))
	//time.Sleep(100 * time.Millisecond)
}

func processTask(taskId, workerId int) (int, error) {
	postBody, _ := json.Marshal(map[string]int{
		"workerId": workerId,
		"taskId":   taskId,
	})
	requestBody := bytes.NewBuffer(postBody)
	resp, err := http.Post("http://localhost:8080/work", "application/json", requestBody)
	if err != nil {
		//log.Fatalf("An Error Occured %v", err)
		//log.Println(err)
		//var workBody = WorkResponse{0, "failure", 422}
		return 0, err
	}
	defer resp.Body.Close()

	var workBody WorkResponse
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	if uErr := json.Unmarshal(bodyBytes, &workBody); uErr != nil {
		log.Println(uErr)
	}
	return workBody.Result, nil
}

func auditor(wg *sync.WaitGroup, mux *sync.Mutex, id int, tasksChnl chan int, tasksRecordChnl chan string) {
	log.Printf("auditor %v work start", id)
	for !(len(tasksChnl) == 0 && len(tasksRecordChnl) == 0) {
		record, ok := <-tasksRecordChnl
		if !ok {
			log.Println(" not ok -_- ")
		}
		//log.Println(record)
		signedRecord := func(id int) string {
			return fmt.Sprintf("%v --- signed by auditor %v", record, id)
		}(id)
		fmt.Println(signedRecord)
		sendRecordToQueue(signedRecord)
	}
	log.Printf("auditor %v work done", id)
	wg.Done()
}

func manager(wg *sync.WaitGroup, mux *sync.Mutex, id int, tasksChnl chan int, numberOfTasks *int) {
	log.Printf("manager %v work start", id)
	for {
		taskId := *numberOfTasks
		if taskId > 0 {
			//time.Sleep(10 * time.Millisecond)
			tasksChnl <- taskId
			*numberOfTasks--
		} else {
			break
		}
	}
	log.Printf("manager %v work done", id)
	wg.Done()
}

func sendRecordToQueue(message string) {
	postBody, _ := json.Marshal(map[string]string{
		"message": message,
	})
	requestBody := bytes.NewBuffer(postBody)
	resp, err := http.Post("http://localhost:8080/sendToQueue", "application/json", requestBody)
	if err != nil {
		//log.Fatalf("An Error Occured %v", err)
		//log.Println(message)
		return
	}
	defer resp.Body.Close()

	//body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	//sb := string(body)
	//log.Printf(sb)
}

type Task struct {
}

func main() {
	var (
		wg               = &sync.WaitGroup{}
		mux              = &sync.Mutex{}
		numberOfTasks    = 1000
		numberOfWorkers  = 50
		numberOfManagers = 10
		numberOfAuditors = 3
		tasksChnl        = make(chan int, numberOfTasks)
		tasksRecordChnl  = make(chan string, numberOfTasks)
		wSpan            time.Duration
	)

	wg.Add(numberOfWorkers + numberOfManagers + numberOfAuditors)

	for id := 1; id <= numberOfManagers; id++ {
		go manager(wg, mux, id, tasksChnl, &numberOfTasks)
	}
	for id := 1; id <= numberOfWorkers; id++ {
		go worker(wg, mux, id, tasksChnl, tasksRecordChnl, &wSpan)
	}
	for id := 1; id <= numberOfAuditors; id++ {
		go auditor(wg, mux, id, tasksChnl, tasksRecordChnl)
	}
	//for i := 1; i <= numberOfTasks; i++ {
	//	//time.Sleep(10 * time.Millisecond)
	//	tasksChnl <- i
	//}
	//for !(len(tasksChnl) == 0 && len(tasksRecordChnl) == 0) {
	//	record, ok := <-tasksRecordChnl
	//	if !ok {
	//		log.Println(" not ok -_- ")
	//	}
	//	//log.Println(record)
	//	sendRecordToQueue(record)
	//}
	wg.Wait()
	log.Printf("%v tasks left.. work span of %v", numberOfTasks, wSpan)
}
