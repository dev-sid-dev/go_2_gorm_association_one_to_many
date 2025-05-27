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
	ID          uint64 `gorm:"primaryKey"`
	Username    string `gorm:"size:64"`
	Password    string `gorm:"size:255"`
	Notes       []Note // Um para muitos
	CreditCards []CreditCard
}

type Note struct {
	gorm.Model
	ID      uint64 `gorm:"primaryKey"`
	Name    string `gorm:"size:255"`
	Content string `gorm:"type:text"`
	UserID  uint64 `gorm:"index"` // 👈 Chave estrangeira (vai para o N)
}

type CreditCard struct {
	gorm.Model
	Number string
	UserID uint64
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

	dsn := "host=localhost user=postgres_develop password=123456 dbname=postgres_develop port=5435 sslmode=disable TimeZone=America/Sao_Paulo"
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: newLogger})

	if err != nil {
		panic("Failed to connect to database!")
	}

	DB = database
}

func dbMigrate() {
	// Corrigir a ordem de criação das tabelas: User deve vir primeiro
	err := DB.AutoMigrate(&User{}, &Note{}, &CreditCard{})
	if err != nil {
		panic("Erro ao executar migração: " + err.Error())
	}
}

func seedDatabase() {
	var count int64
	DB.Model(&User{}).Count(&count)
	if count > 0 {
		fmt.Println("Banco já possui dados. Pulação ignorada.")
		return
	}

	users := []User{
		{Username: "alice@example.com", Password: "123456"},
		{Username: "bob@example.com", Password: "654321"},
		{Username: "carol@example.com", Password: "abcdef"},
	}

	for idx, user := range users {
		DB.Create(&user)

		// Criando 3 notas para cada usuário
		for i := 1; i <= 3; i++ {
			note := Note{
				Name:    fmt.Sprintf("Nota %d de %s", i, user.Username),
				Content: fmt.Sprintf("Conteúdo fictício %d para teste.", i),
				UserID:  user.ID,
			}
			DB.Create(&note)
		}

		// Criando múltiplos cartões de crédito por usuário
		var cardCount int
		switch idx {
		case 0, 1:
			cardCount = 2 // alice e bob
		case 2:
			cardCount = 3 // carol
		}

		for i := 1; i <= cardCount; i++ {
			card := CreditCard{
				Number: fmt.Sprintf("4111-2222-3333-%04d", user.ID*10+uint64(i)),
				UserID: user.ID,
			}
			DB.Create(&card)
		}
	}
	fmt.Println("Dados de exemplo inseridos com sucesso.")
}

func main() {
	connectDatabase()
	dbMigrate()
	seedDatabase()

	var note Note
	DB.First(&note)
	var user User
	DB.Where("id = ?", note.UserID).First(&user)
	fmt.Printf("User from a note: %s\n", user.Username)

	fmt.Println("\n----------------")

	var notes []Note
	DB.Where("user_id = ?", user.ID).Find(&notes)

	fmt.Println("Notes from a user:")
	for _, element := range notes {
		fmt.Printf("%s - %s\n", element.Name, element.Content)
	}
	fmt.Println("\n----------------")

	var cc CreditCard
	DB.Where("user_id = ?", user.ID).First(&cc)
	fmt.Printf("Credit Card from a user: %s\n", cc.Number)
}
