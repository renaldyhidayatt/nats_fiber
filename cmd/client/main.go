package main

import (
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/nats-io/nats.go"
)

type Person struct {
	ID        uint   `json:"id"`
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

	app := fiber.New()

	// Mendapatkan data seseorang berdasarkan ID
	app.Get("/person/:id", func(c *fiber.Ctx) error {
		personID := c.Params("id")

		msg, err := nc.Request("get.person", []byte(personID), time.Second)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Error fetching person")
		}

		var person Person
		err = json.Unmarshal(msg.Data, &person)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Error decoding response")
		}

		return c.JSON(person)
	})

	// Menambahkan data seseorang
	app.Post("/person", func(c *fiber.Ctx) error {
		var person Person
		if err := c.BodyParser(&person); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid input")
		}

		personJSON, err := json.Marshal(person)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Error encoding response")
		}

		nc.Publish("create.person", personJSON)
		return c.JSON(person)
	})

	// Mengubah data seseorang berdasarkan ID
	app.Put("/person/:id", func(c *fiber.Ctx) error {
		personID := c.Params("id")

		// Konversi personID dari string ke uint
		id, err := strconv.ParseUint(personID, 10, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid input")
		}

		var person Person
		if err := c.BodyParser(&person); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid input")
		}

		person.ID = uint(id)
		personJSON, err := json.Marshal(person)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Error encoding response")
		}

		nc.Publish("update.person", personJSON)
		return c.JSON(person)
	})

	// Menghapus data seseorang berdasarkan ID
	app.Delete("/person/:id", func(c *fiber.Ctx) error {
		personID := c.Params("id")

		personJSON, err := json.Marshal(map[string]interface{}{
			"id": personID,
		})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Error encoding response")
		}

		nc.Publish("delete.person", personJSON)
		return c.SendString("Person deleted")
	})

	app.Listen(":8080")
}
