package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Person struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `json:"name"`
	Age       int    `json:"age"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func main() {
	// Buat koneksi NATS
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	// Buat koneksi ke database SQLite menggunakan GORM
	dsn := "gorm.db"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	db.AutoMigrate(&Person{})

	// Berlangganan pesan "create.person"
	nc.Subscribe("create.person", func(msg *nats.Msg) {
		var person Person
		err := json.Unmarshal(msg.Data, &person)
		if err != nil {
			log.Println("Error decoding message:", err)
			return
		}

		db.Create(&person)
		log.Println("Person created:", person)
	})

	// Berlangganan pesan "get.person"
	nc.Subscribe("get.person", func(msg *nats.Msg) {
		var person Person
		err := json.Unmarshal(msg.Data, &person)
		if err != nil {
			log.Println("Error decoding message:", err)
			return
		}

		var result Person
		db.First(&result, person.ID)
		data, _ := json.Marshal(result)
		nc.Publish(msg.Reply, data)
	})

	// Berlangganan pesan "update.person"
	nc.Subscribe("update.person", func(msg *nats.Msg) {
		var person Person
		err := json.Unmarshal(msg.Data, &person)
		if err != nil {
			log.Println("Error decoding message:", err)
			return
		}

		var existingPerson Person
		db.First(&existingPerson, person.ID)
		existingPerson.Name = person.Name
		existingPerson.Age = person.Age
		db.Save(&existingPerson)
		log.Println("Person updated:", existingPerson)
	})

	// Berlangganan pesan "delete.person"
	nc.Subscribe("delete.person", func(msg *nats.Msg) {
		var person Person
		err := json.Unmarshal(msg.Data, &person)
		if err != nil {
			log.Println("Error decoding message:", err)
			return
		}

		db.Delete(&person)
		log.Println("Person deleted:", person)
	})

	fmt.Println("Listening for messages...")
	select {}
}
