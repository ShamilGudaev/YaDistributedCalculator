package db

import (
	"backend/parser"
	"fmt"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Expression struct {
	ID        uint64 `gorm:"primaryKey"`
	Text      string `gorm:"not null"`
	Result    *float64
	AgentID   *string
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

type Agent struct {
	ID          string    `gorm:"primaryKey"`
	LastSeen    time.Time `gorm:"not null"`
	DeletedAt   *time.Time
	Expressions []Expression
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}

type ExecutionTime struct {
	Operator   parser.Operator `gorm:"primaryKey;autoIncrement:false"`
	DurationMS uint32          `gorm:"not null"`
}

var DB *gorm.DB

func OpenDB() {
	password, err := os.ReadFile("/run/secrets/db_password")
	if err != nil {
		panic(err.Error())
	}

	dsn := fmt.Sprintf("host=db user=postgres password=%s dbname=postgres port=5432 sslmode=disable", string(password))

	DB, err = gorm.Open(postgres.Open(dsn))

	for i := 0; i < 50; i++ {
		if err == nil {
			break
		}

		time.Sleep(1000 * time.Millisecond)
		DB, err = gorm.Open(postgres.Open(dsn))
	}

	if err != nil {
		panic(err.Error())
	}

	DB.AutoMigrate(&Agent{}, &Expression{}, &ExecutionTime{})

	DB.Clauses(clause.OnConflict{
		DoNothing: true,
	}).Create([]*ExecutionTime{
		{Operator: parser.OpMul, DurationMS: 5000},
		{Operator: parser.OpDiv, DurationMS: 5000},
		{Operator: parser.OpAdd, DurationMS: 5000},
		{Operator: parser.OpSub, DurationMS: 5000},
	})
}
