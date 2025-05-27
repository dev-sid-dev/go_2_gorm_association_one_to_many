package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type User struct {
	gorm.Model
	ID         uint64 `gorm:"primaryKey"`
	Username   string `gorm:"size:64"`
	Password   string `gorm:"size:255"`
	Notes      []Note
	CreditCard *CreditCard
}

type Note struct {
	gorm.Model
	ID      uint64 `gorm:"primaryKey"`
	Name    string `gorm:"size:255"`
	Content string `gorm:"type:text"`
	UserID  uint64 `gorm:"index"`
	User    User
}

type CreditCard struct {
	gorm.Model
	Number string
	UserID uint64
	User   User
}

var DB *gorm.DB

func connectDatabase() {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	// DSN: ajuste conforme seu ambiente PostgreSQL
	dsn := "host=localhost user=postgres_develop password=123456 dbname=postgres_develop port=5435 sslmode=disable TimeZone=America/Sao_Paulo"
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: newLogger})

	if err != nil {
		panic("Failed to connect to database!")
	}

	DB = database
}

func dbMigrate() {
	DB.AutoMigrate(&Note{}, &User{}, &CreditCard{})
}

func main() {
	connectDatabase()
	dbMigrate()

	var note Note
	DB.Preload("User").First(&note)
	fmt.Printf("User from a note: %s\n", note.User.Username)

	fmt.Println("\n----------------")

	var user User
	DB.Preload("Notes").Preload("CreditCard").Where("username = ?", "user1@codeheim").First(&user)

	fmt.Println("Notes from a user:")
	for _, element := range user.Notes {
		fmt.Printf("%s - %s\n", element.Name, element.Content)
	}
	fmt.Println("\n----------------")

	fmt.Printf("Credit Card from a user: %s\n", user.CreditCard.Number)
}
