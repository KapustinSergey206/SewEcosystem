package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"sewing-ecosystem/internal/app"

	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/SevereCloud/vksdk/v2/events"
	longpoll "github.com/SevereCloud/vksdk/v2/longpoll-bot"
)

type BotState struct {
	Step  string
	Name  string
	Phone string
}

var states = map[int]*BotState{}

func main() {

	cfg := app.LoadConfig()

	db, err := app.OpenDB(cfg.PostgresDSN)
	if err != nil {
		log.Fatal(err)
	}

	vk := api.NewVK(cfg.BotToken)

	lp, err := longpoll.NewLongPollCommunity(vk)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("VK bot started")

	lp.MessageNew(func(ctx context.Context, obj events.MessageNewObject) {

		userID := obj.Message.PeerID
		text := strings.TrimSpace(obj.Message.Text)

		if _, ok := states[userID]; !ok {
			states[userID] = &BotState{}
		}

		state := states[userID]

		switch text {

		case "Начать":

			send(vk, userID,
				"👋 Добро пожаловать\n\n"+
					"Команды:\n"+
					"Каталог\n"+
					"Заказ\n"+
					"Статус\n"+
					"Адрес")

		case "Каталог":

			rows, err := db.Query(
				`SELECT id, name, price FROM products ORDER BY id`,
			)

			if err != nil {
				send(vk, userID, "Ошибка каталога")
			}

			defer rows.Close()

			msg := "📦 Каталог:\n\n"

			for rows.Next() {

				var id int
				var name string
				var price float64

				err := rows.Scan(
					&id,
					&name,
					&price,
				)

				if err != nil {
					continue
				}

				msg += fmt.Sprintf(
					"%d. %s — %.2f ₽\n",
					id,
					name,
					price,
				)
			}

			send(vk, userID, msg)

		case "Заказ":

			state.Step = "name"

			send(vk, userID,
				"Введите ФИО")

		case "Статус":

			state.Step = "status"

			send(vk, userID,
				"Введите ID заказа")

		case "Адрес":

			send(vk, userID,
				fmt.Sprintf(
					"%s\n%s",
					cfg.CompanyName,
					cfg.CompanyAddress,
				))

		default:

			switch state.Step {

			case "name":

				state.Name = text
				state.Step = "phone"

				send(vk, userID,
					"Введите телефон")

			case "phone":

				state.Phone = text
				state.Step = "product"

				send(vk, userID,
					"Введите:\nID_товара количество\n\nПример:\n1 5")

			case "product":

				parts := strings.Split(text, " ")

				if len(parts) != 2 {

					send(vk, userID,
						"Неверный формат")

				}

				productID, err := strconv.Atoi(parts[0])
				if err != nil {

					send(vk, userID,
						"Ошибка ID")

				}

				qty, err := strconv.Atoi(parts[1])
				if err != nil {

					send(vk, userID,
						"Ошибка количества")

				}

				var productName string
				var price float64

				err = db.QueryRow(
					`SELECT name, price
					 FROM products
					 WHERE id=$1`,
					productID,
				).Scan(
					&productName,
					&price,
				)

				if err != nil {

					send(vk, userID,
						"Товар не найден")

				}

				total := float64(qty) * price

				var orderID int

				err = db.QueryRow(
					`
					INSERT INTO orders
					(contact_name, contact_phone, status, total_amount)
					VALUES ($1,$2,$3,$4)
					`,
					state.Name,
					state.Phone,
					"new",
					total,
				).Scan(&orderID)

				// if err != nil {

				// 	log.Println(err)

				// 	send(vk, userID,
				// 		"Ошибка создания заказа")

				// }

				_, err = db.Exec(
					`
					INSERT INTO order_items
					(order_id, product_id, quantity, unit_price)
					VALUES ($1,$2,$3,$4)
					`,
					orderID,
					productID,
					qty,
					price,
				)

				var orederNum string

				row, err := db.Query("Select order_number From orders")
				for row.Next() {
					row.Scan(&orederNum)
				}

				if err != nil {

					log.Println(err)

					send(vk, userID,
						"Ошибка добавления товара")

				}

				send(vk, userID,
					fmt.Sprintf(
						"✅ Заказ #%v создан\n\n"+
							"📦 %s\n"+
							"🔢 %d шт.\n"+
							"💰 %.2f ₽",
						orederNum,
						productName,
						qty,
						total,
					),
				)

				states[userID] = &BotState{}

			case "status":

				orderID := text
				if err != nil {

					send(vk, userID,
						"Введите корректный ID")

				}

				var status string

				err = db.QueryRow(
					`
					SELECT status
					FROM orders
					WHERE order_number=$1
					`,
					orderID,
				).Scan(&status)

				if err != nil {

					send(vk, userID,
						"Заказ не найден")

				}

				send(vk, userID,
					fmt.Sprintf(
						"📄 Заказ #%v\nСтатус: %s",
						orderID,
						status,
					),
				)
			}
		}
	})

	log.Fatal(lp.Run())
}

func send(vk *api.VK, peerID int, message string) {

	_, err := vk.MessagesSend(api.Params{
		"peer_id":   peerID,
		"message":   message,
		"random_id": 0,
	})

	if err != nil {
		log.Println(err)
	}
}
