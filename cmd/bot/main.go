package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"sewing-ecosystem/internal/app"
	"sewing-ecosystem/internal/models"
	"sewing-ecosystem/internal/repo"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type botState struct {
	Step    string
	Name    string
	Phone   string
	Comment string
	Items   []models.OrderItem
}

func main() {
	cfg := app.LoadConfig()

	if cfg.BotToken == "" {
		log.Fatal("BOT_TOKEN is empty")
	}

	db, err := app.OpenDB(cfg.PostgresDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rp := repo.New(db)

	bot, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		log.Fatal(err)
	}

	states := map[int64]*botState{}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30
	updates := bot.GetUpdatesChan(u)

	for upd := range updates {
		if upd.Message == nil {
			continue
		}

		chatID := upd.Message.Chat.ID
		text := strings.TrimSpace(upd.Message.Text)

		s := states[chatID]
		if s != nil && !strings.HasPrefix(text, "/") {
			handleOrderFlow(bot, rp, chatID, text, s)
			if s.Step == "done" {
				delete(states, chatID)
			}
			continue
		}

		switch {
		case text == "/start" || text == "/help":
			msg := "" +
				"/products — каталог\n" +
				"/location — геометка компании\n" +
				"/status ORDER-000001 — статус заказа\n" +
				"/order — оформить заказ"
			bot.Send(tgbotapi.NewMessage(chatID, msg))

		case text == "/products":
			products, err := rp.ListProducts(context.Background())
			if err != nil {
				bot.Send(tgbotapi.NewMessage(chatID, "Не удалось загрузить каталог"))
				continue
			}

			var b strings.Builder
			b.WriteString("Каталог продукции:\n\n")
			for _, p := range products {
				b.WriteString(fmt.Sprintf("%d. %s\n", p.ID, p.Name))
				b.WriteString(fmt.Sprintf("   Категория: %s\n", p.Category))
				b.WriteString(fmt.Sprintf("   Цена: %.2f ₽\n\n", p.Price))
			}
			bot.Send(tgbotapi.NewMessage(chatID, b.String()))

		case text == "/location":
			company, err := rp.GetCompanyInfo(context.Background())
			if err != nil {
				bot.Send(tgbotapi.NewMessage(chatID, "Данные компании не найдены"))
				continue
			}

			loc := tgbotapi.NewLocation(chatID, company.Lat, company.Lon)
			bot.Send(loc)

			msg := fmt.Sprintf("%s\n%s\n%s", company.Name, company.Address, company.Phone)
			bot.Send(tgbotapi.NewMessage(chatID, msg))

		case strings.HasPrefix(text, "/status"):
			parts := strings.Fields(text)
			if len(parts) < 2 {
				bot.Send(tgbotapi.NewMessage(chatID, "Использование: /status ORDER-000001"))
				continue
			}

			status, err := rp.SearchOrderStatus(context.Background(), parts[1])
			if err != nil {
				bot.Send(tgbotapi.NewMessage(chatID, "Заказ не найден"))
				continue
			}

			bot.Send(tgbotapi.NewMessage(chatID, "Статус заказа "+parts[1]+": "+status))

		case text == "/order":
			states[chatID] = &botState{Step: "name"}
			bot.Send(tgbotapi.NewMessage(chatID, "Введите ФИО клиента"))

		default:
			bot.Send(tgbotapi.NewMessage(chatID, "Не понял команду. Напишите /help"))
		}
	}
}

func handleOrderFlow(bot *tgbotapi.BotAPI, rp *repo.Repository, chatID int64, text string, s *botState) {
	switch s.Step {
	case "name":
		s.Name = strings.TrimSpace(text)
		s.Step = "phone"
		bot.Send(tgbotapi.NewMessage(chatID, "Введите телефон"))

	case "phone":
		s.Phone = strings.TrimSpace(text)
		s.Step = "comment"
		bot.Send(tgbotapi.NewMessage(chatID, "Введите комментарий к заказу или '-' если без комментария"))

	case "comment":
		if strings.TrimSpace(text) == "-" {
			s.Comment = ""
		} else {
			s.Comment = strings.TrimSpace(text)
		}

		s.Step = "items"

		products, err := rp.ListProducts(context.Background())
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "Не удалось загрузить каталог"))
			return
		}

		var b strings.Builder
		b.WriteString("Введите позиции в формате ID:количество через запятую.\n")
		b.WriteString("Пример: 1:10,2:5\n\n")

		for _, p := range products {
			b.WriteString(fmt.Sprintf("%d. %s — %.2f ₽\n", p.ID, p.Name, p.Price))
		}

		bot.Send(tgbotapi.NewMessage(chatID, b.String()))

	case "items":
		parts := strings.Split(text, ",")

		var items []models.OrderItem
		var total float64

		for _, p := range parts {
			pair := strings.Split(strings.TrimSpace(p), ":")
			if len(pair) != 2 {
				continue
			}

			id, _ := strconv.ParseInt(strings.TrimSpace(pair[0]), 10, 64)
			qty, _ := strconv.Atoi(strings.TrimSpace(pair[1]))
			if qty <= 0 {
				continue
			}

			prod, err := rp.GetProduct(context.Background(), id)
			if err != nil {
				continue
			}

			subtotal := prod.Price * float64(qty)
			total += subtotal

			items = append(items, models.OrderItem{
				ProductID: id,
				Quantity:  qty,
				UnitPrice: prod.Price,
				Subtotal:  subtotal,
			})
		}

		if len(items) == 0 {
			bot.Send(tgbotapi.NewMessage(chatID, "Не удалось распознать позиции заказа. Пример: 1:10,2:5"))
			return
		}

		order, err := rp.CreateGuestOrder(
			context.Background(),
			s.Name,
			s.Phone,
			"Заказ через Telegram-бота. "+s.Comment,
			items,
		)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "Ошибка создания заказа"))
			return
		}

		s.Step = "done"

		msg := fmt.Sprintf(
			"Заказ создан.\nНомер: %s\nСумма: %.2f ₽\nПроверить статус: /status %s",
			order.OrderNumber,
			order.TotalAmount,
			order.OrderNumber,
		)
		bot.Send(tgbotapi.NewMessage(chatID, msg))
	}
}
