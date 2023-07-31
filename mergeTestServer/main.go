package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/streadway/amqp"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type opts struct {
	EncryptionKey string
	Port          string
}

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.Writer.Header()
		header.Set("Access-Control-Allow-Origin", "*")
		header.Set("Access-Control-Allow-Methods", "*")
		header.Set("Access-Control-Allow-Headers", "Content-Type,Authorization")
		header.Set("Access-Control-Max-Age", "86400")

		if c.Request.Method == http.MethodOptions {
			c.Writer.WriteHeader(http.StatusOK)
			return
		}

		c.Next()
		return
	}
}

type InMessage struct {
	Message string `json:"message"`
}

type WorkMessage struct {
	WorkerId int `json:"workerId"`
	TaskId   int `json:"taskId"`
}

func ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func sendToQueue(c *gin.Context) {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"hello", // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	failOnError(err, "Failed to declare a queue")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// body := c.Request.Body
	var body InMessage
	bodyBytes, err := ioutil.ReadAll(c.Request.Body)
	if uErr := json.Unmarshal(bodyBytes, &body); uErr != nil {
		failOnError(uErr, "Failed to unmarshall")
		log.Println(uErr)
		statusCode := http.StatusBadRequest
		c.JSON(statusCode, map[string]interface{}{
			"error":      err,
			"status":     "failure",
			"statusCode": statusCode,
		})
		return
	}

	fmt.Println(ctx, q)
	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        bodyBytes,
		})
	if err != nil {
		statusCode := http.StatusBadRequest
		c.JSON(statusCode, map[string]interface{}{
			"error":      err,
			"status":     "failure",
			"statusCode": statusCode,
		})
		return
	}
	failOnError(err, "Failed to publish a message")
	log.Printf(" [x] Sent %v\n", string(body.Message))

	statusCode := http.StatusOK
	c.JSON(statusCode, map[string]interface{}{
		"message":    body.Message,
		"status":     "success",
		"statusCode": statusCode,
	})
	return
}

func work(c *gin.Context) {
	var body WorkMessage
	bodyBytes, err := ioutil.ReadAll(c.Request.Body)
	if uErr := json.Unmarshal(bodyBytes, &body); uErr != nil {
		failOnError(uErr, "Failed to unmarshall")
		log.Println(uErr)
		statusCode := http.StatusBadRequest
		c.JSON(statusCode, map[string]interface{}{
			"error":      err,
			"status":     "failure",
			"statusCode": statusCode,
		})
		return
	}

	fCh := make(chan int)
	go func(fn int) {
		a, b := 0, 1
		if fn == 0 || fn == 1 {
			fCh <- b
			return
		}
		for i := 0; i < fn; i++ {
			a, b = b, a+b
		}
		fCh <- b
		return
	}(body.WorkerId * body.TaskId)
	fib := <-fCh

	statusCode := http.StatusOK
	c.JSON(statusCode, map[string]interface{}{
		"result":     fib,
		"status":     "success",
		"statusCode": statusCode,
	})
	return
}

//func sendToQueueInterval(c *gin.Context) {
//	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
//	failOnError(err, "Failed to connect to RabbitMQ")
//	defer conn.Close()
//
//	ch, err := conn.Channel()
//	failOnError(err, "Failed to open a channel")
//	defer ch.Close()
//
//	q, err := ch.QueueDeclare(
//		"hello", // name
//		false,   // durable
//		false,   // delete when unused
//		false,   // exclusive
//		false,   // no-wait
//		nil,     // arguments
//	)
//	failOnError(err, "Failed to declare a queue")
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//
//	// body := c.Request.Body
//	var body InMessage
//	bodyBytes, err := ioutil.ReadAll(c.Request.Body)
//	if uErr := json.Unmarshal(bodyBytes, &body); uErr != nil {
//		failOnError(uErr, "Failed to unmarshall")
//		log.Println(uErr)
//		statusCode := http.StatusBadRequest
//		c.JSON(statusCode, map[string]interface{}{
//			"error":      err,
//			"status":     "failure",
//			"statusCode": statusCode,
//		})
//		return
//	}
//
//	fmt.Println(ctx, q)
//	err = ch.Publish(
//		"",     // exchange
//		q.Name, // routing key
//		false,  // mandatory
//		false,  // immediate
//		amqp.Publishing{
//			ContentType: "text/plain",
//			Body:        bodyBytes,
//		})
//	if err != nil {
//		statusCode := http.StatusBadRequest
//		c.JSON(statusCode, map[string]interface{}{
//			"error":      err,
//			"status":     "failure",
//			"statusCode": statusCode,
//		})
//		return
//	}
//	failOnError(err, "Failed to publish a message")
//	log.Printf(" [x] Sent %v\n", string(body.Message))
//
//	statusCode := http.StatusOK
//	c.JSON(statusCode, map[string]interface{}{
//		"message":    body.Message,
//		"status":     "success",
//		"statusCode": statusCode,
//	})
//	return
//}

func main() {
	r := gin.Default()
	r.GET("/ping", ping)
	r.POST("/sendToQueue", sendToQueue)
	r.POST("/work", work)
	//r.POST("/sendToQueueInterval", sendToQueueInterval)
	r.Run()
}

//func Run(_ *cli.Context) {
//
//	//startServer := time.Now()
//	// fmt.Println("Configuring Stackdriver")
//	//ctx := context.Background()
//	// closer, err := census.Configure(ctx, opts.Version, opts.Env)
//	// if err != nil {
//	// 	return err
//	// }
//	// defer closer.Close()
//	// fmt.Println("Done!")
//	// dc
//
//	httpServer := startDC()
//	defer httpServer.Shutdown(context.Background())
//
//	//m, httpServer := startDC(opts.DC.Endpoint3, opts.DC.Endpoint2, opts.DC.Endpoint, opts.DC.Addr, opts.DC.Port)
//	//defer httpServer.Shutdown(context.Background())
//
//	gin.SetMode(gin.ReleaseMode)
//	routes := gin.New()
//	routes.Use(
//		CORS(),
//		gin.Recovery(),
//		gin.Logger(),
//	)
//	r := routes.Group("/v1",
//		//jwe.RequireSession(opts.EncryptionKey, opts.MaxAge, "/v1/readiness", "/v1/auth_token", "/v1/traces/", "/v1/dkyc/", "/v1/object_detec"),
//		gin.Logger(),
//		//tracing,
//	)
//
//	//r := gin.Default()
//	r.GET("/ping", ping)
//	r.POST("/sendToQueue", sendToQueue)
//	r.POST("/work", work)
//	//r.POST("/sendToQueueInterval", sendToQueueInterval)
//	routes.Run()
//}
//
//func _main() {
//	app := cli.NewApp()
//	//app.Flags = ""
//	app.Action = Run
//
//	if err := app.Run(os.Args); err != nil {
//		log.Fatal(err)
//	}
//}
//
//func startDC() *http.Server {
//	router := gin.New()
//	fmt.Println("Port: ", "8080")
//	httpServer := &http.Server{
//		Addr:    ":" + "8080",
//		Handler: router,
//	}
//	go httpServer.ListenAndServe()
//
//	return httpServer
//}
