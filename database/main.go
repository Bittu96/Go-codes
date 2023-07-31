package postgresdb

import (
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"log"
	"time"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "mrudulkatla"
	password = "postgres"
	dbname   = "postgres"
)

// add table schemas below ---------------------------------------------------------------------------------------------

// task logs table
type taskLogs struct {
	ID         int       `gorm:"type:int;primary_key;auto_increment"`
	Record     string    `gorm:"type:varchar(200)"`
	RowCreated time.Time `gorm:"type:timestamp;DEFAULT:current_timestamp"`
}

// task records table
type taskRecords struct {
	//gorm.Model
	WorkerId int       `json:"worker_id" gorm:"int(8)"`
	TaskId   int       `json:"task_id" gorm:"int(8)"`
	Score    int       `json:"score" gorm:"int(16)"`
	Duration time.Time `json:"duration" gorm:"timestamp"`
	Status   string    `json:"status" gorm:"varchar(16)"`
	Message  string    `json:"message" gorm:"varchar(256)"`
}

// AppConfig Table 1
type AppConfig struct {
	AppID     string    `json:"app_id"  gorm:"type:varchar(32)"`
	Status    bool      `json:"status" gorm:"type:bool"`
	CreatedAt time.Time `json:"created_at" gorm:"type:timestamp"`
	UpdatedAt time.Time `json:"updated_at" gorm:"type:timestamp"`
}

// APIModelConfig Table 2
type APIModelConfig struct {
	SID         int32       `json:"sid" gorm:"type:int"`
	AppID       string      `json:"app_id" gorm:"type:varchar(32)"`
	API         string      `json:"api" gorm:"type:varchar(64)"`
	ModelConfig interface{} `json:"model_config" gorm:"type:jsonb"`
	EnvConfig   interface{} `json:"env_config" gorm:"type:jsonb"`
	Status      bool        `json:"status" gorm:"type:bool"`
	CreatedAt   time.Time   `json:"created_at" gorm:"type:timestamp"`
	UpdatedAt   time.Time   `json:"updated_at" gorm:"type:timestamp"`
}
type ModelName string
type ModelConfig struct {
	Status string `json:"status"`
	Accept bool   `json:"accept"`
}
type EnvConfig struct {
	DeliveryMode    string `json:"delivery_mode"`
	BirdmanRedirect bool   `json:"birdman_redirect"`
}

// add table schemas above ---------------------------------------------------------------------------------------------

type DB interface {
	GetConnection() *gorm.DB
	CreateTable(dbConn *gorm.DB)
	GetAllRecords(dbConn *gorm.DB)
	CreateRecord(dbConn *gorm.DB)
}

func GetConnection() *gorm.DB {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := gorm.Open("postgres", psqlInfo)

	if err != nil {
		log.Println(err)
		//dummyDB := gorm.DB{}
		//return &dummyDB
		panic("failed to connect database")
	}

	log.Println("DB Connection established...")
	return db
}

func CreateTable(dbConn *gorm.DB) {
	resp := dbConn.AutoMigrate(&APIModelConfig{})
	log.Println("automigrate", resp)
}

func GetAllRecords(dbConn *gorm.DB) {
	var resp []taskRecords
	err := dbConn.Find(&resp)
	log.Println("find", resp)
	//fmt.Println(err)
	log.Printf("%+v", resp)

	if err == nil {
		log.Println("table found")
	} else {
		log.Println(err)
	}
}

func CreateRecord(dbConn *gorm.DB) {
	//var taskRecord = taskLogs{Record: "test record", RowCreated: time.Now()}
	//var taskRecord = taskRecords{WorkerId: 1, TaskId: 2, Score: 10, Duration: time.Now(), Message: "test record"}
	//dummyAppConfigData := AppConfig{"dkyc", true, time.Time{}, time.Time{}}

	dummyModelConfig, _ := json.Marshal(map[ModelName]ModelConfig{
		"BlurCheck":            ModelConfig{"enabled", true},
		"EyeCheck":             ModelConfig{"disabled", true},
		"WhiteBackgroundCheck": ModelConfig{"enabled", false},
	})
	dummyEnvConfig, _ := json.Marshal(EnvConfig{
		BirdmanRedirect: false,
		DeliveryMode:    "SDK",
	})
	dummyApiConfigData := APIModelConfig{1, "dkyc", "check_liveness", string(dummyModelConfig), string(dummyEnvConfig), true, time.Time{}, time.Time{}}

	err := dbConn.Create(&dummyApiConfigData).Error
	if err != nil {
		log.Println("log not created")
		log.Println(err)
	} else {
		log.Println("log created")
		log.Println(err)
	}
}

func main() {
	dbConn := GetConnection()
	log.Println(dbConn)
	//CreateTable(dbConn)
	//GetAllRecords(dbConn)
	CreateRecord(dbConn)
}
