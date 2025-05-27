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
	Username string `gorm:"size:64"`
	Password string `gorm:"size:255"`
	Notes    []Note // Um para muitos
	//CreditCard CreditCard // 👈 relação 1:1 // 👈 valor direto (não pode ser nulo)
	CreditCard *CreditCard // 👈 ponteiro (permite valor nulo)
}

//func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
//	fmt.Println("🔔 BeforeCreate: validando/ajustando antes de inserir o usuário...")
//	u.Username = "safe_" + u.Username // prefixo automático
//	return
//}
//
//func (u *User) AfterCreate(tx *gorm.DB) (err error) {
//	fmt.Println("✅ AfterCreate: usuário inserido com ID =", u.ID)
//	return
//}

type Note struct {
	gorm.Model
	Name    string `gorm:"size:255"`
	Content string `gorm:"type:text"`
	UserID  uint   `gorm:"index"` // 👈 Chave estrangeira (vai para o N)
	User    User   // 👈 opcional, útil se quiser navegar reversamente
}

type CreditCard struct {
	gorm.Model
	Number string
	UserID uint
	User   User // 👈 opcional, permite navegar reversamente
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
		{
			Username: "alice@example.com",
			Password: "123456",
			CreditCard: &CreditCard{
				Number: "4111-2222-3333-0001",
			},
		},
		{
			Username: "bob@example.com",
			Password: "654321",
			CreditCard: &CreditCard{
				Number: "4111-2222-3333-0002",
			},
		},
		{
			Username: "carol@example.com",
			Password: "abcdef",
		},
	}

	for _, user := range users {
		// Cria o usuário e seu cartão de crédito junto
		DB.Create(&user)

		// Cria 3 notas por usuário
		for i := 1; i <= 3; i++ {
			note := Note{
				Name:    fmt.Sprintf("Nota %d de %s", i, user.Username),
				Content: fmt.Sprintf("Conteúdo fictício %d para teste.", i),
				UserID:  user.ID,
			}
			DB.Create(&note)
		}
	}

	fmt.Println("Dados de exemplo inseridos com sucesso.")
}

func loadWithPreload() {
	var user User
	err := DB.Preload("Notes").Preload("CreditCard").First(&user, "username = ?", "carol@example.com").Error
	if err != nil {
		log.Fatal("Erro ao carregar usuário:", err)
	}

	fmt.Println("Usuário:", user.Username)

	fmt.Println("\nNotas:")
	for _, note := range user.Notes {
		fmt.Printf("- %s: %s\n", note.Name, note.Content)
	}

	if user.CreditCard != nil {
		fmt.Println("\nCartão de Crédito:", user.CreditCard.Number)
	} else {
		fmt.Println("\nCartão de Crédito: nenhum")
	}
}

func rawSQLExample() {
	type Result struct {
		Username  string
		NoteCount int
	}

	var results []Result

	sql := `
		SELECT u.username, COUNT(n.id) as note_count
		FROM users u
		JOIN notes n ON u.id = n.user_id
		GROUP BY u.username
		HAVING COUNT(n.id) > ?
	`

	if err := DB.Raw(sql, 2).Scan(&results).Error; err != nil {
		log.Fatal("Erro ao executar Raw SQL:", err)
	}

	fmt.Println("\n🔍 Usuários com mais de 2 notas:")
	for _, r := range results {
		fmt.Printf("- %s: %d notas\n", r.Username, r.NoteCount)
	}
}

func execSQLExample() {
	res := DB.Exec(`UPDATE users SET password = ? WHERE username = ?`, "novaSenha123", "alice@example.com")

	if res.Error != nil {
		log.Fatal("Erro ao executar Exec SQL:", res.Error)
	}

	fmt.Printf("\n🔧 Senhas atualizadas: %d\n", res.RowsAffected)
}

func joinQueryExample() {
	type Result struct {
		Username     string
		NoteName     string
		CreditNumber *string
	}

	var results []Result

	err := DB.Table("users").
		Joins("LEFT JOIN notes ON notes.user_id = users.id").
		Joins("LEFT JOIN credit_cards ON credit_cards.user_id = users.id").
		Select("users.username, notes.name AS note_name, credit_cards.number AS credit_number").
		Where("users.username = ?", "alice@example.com").
		Scan(&results).Error

	if err != nil {
		log.Fatal("Erro na consulta com JOIN:", err)
	}

	fmt.Println("\n📄 Resultado da consulta com JOIN:")
	for _, r := range results {
		card := "nenhum"
		if r.CreditNumber != nil {
			card = *r.CreditNumber
		}
		fmt.Printf("- Usuário: %s | Nota: %s | Cartão: %s\n", r.Username, r.NoteName, card)
	}
}

func testHooksExample() {
	user := User{
		Username: "daniel@example.com",
		Password: "segredo123",
	}

	if err := DB.Create(&user).Error; err != nil {
		log.Fatal("Erro ao criar usuário:", err)
	}
}

func transactionExample() {
	err := DB.Transaction(func(tx *gorm.DB) error {
		user := User{Username: "transacao@example.com", Password: "123"}
		if err := tx.Create(&user).Error; err != nil {
			return err // rollback automático
		}

		note := Note{Name: "Nota com transação", Content: "Conteúdo seguro", UserID: user.ID}
		if err := tx.Create(&note).Error; err != nil {
			return err // rollback automático
		}

		fmt.Println("✅ Transação executada com sucesso")
		return nil // commit
	})

	if err != nil {
		log.Fatal("Erro na transação:", err)
	}
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

	loadWithPreload()
	rawSQLExample()
	execSQLExample()
	joinQueryExample()
	//	testHooksExample()
	transactionExample()

}
