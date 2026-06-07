package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"sewing-ecosystem/internal/app"
	"sewing-ecosystem/internal/auth"
	"sewing-ecosystem/internal/models"
	"sewing-ecosystem/internal/repo"
)

func main() {
	cfg := app.LoadConfig()

	db, err := app.OpenDB(cfg.PostgresDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rp := repo.New(db)
	seedAdmin(rp, cfg)

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("\n--- АИС швейного предприятия ---")
		fmt.Println("1. Список изделий")
		fmt.Println("2. Все заказы")
		fmt.Println("3. Создать заказ")
		fmt.Println("4. Открыть заказ по номеру")
		fmt.Println("5. Изменить статус заказа")
		fmt.Println("6. Данные компании")
		fmt.Println("7. Выход")
		fmt.Print("Выберите пункт: ")

		switch readLine(reader) {
		case "1":
			listProducts(rp)
		case "2":
			listOrders(rp)
		case "3":
			createOrder(rp, reader)
		case "4":
			showOrder(rp, reader)
		case "5":
			updateStatus(rp, reader)
		case "6":
			showCompany(rp)
		case "7":
			return
		default:
			fmt.Println("Неверный пункт")
		}
	}
}

func seedAdmin(rp *repo.Repository, cfg app.Config) {
	hash, _ := auth.HashPassword(cfg.DefaultAdminPassword)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rp.EnsureAdmin(ctx, "Главный администратор", cfg.DefaultAdminEmail, "+70000000000", hash); err != nil {
		log.Println("admin seed warning:", err)
	}
}

func listProducts(rp *repo.Repository) {
	items, err := rp.ListProducts(context.Background())
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Println("\n--- Каталог продукции ---")
	for _, p := range items {
		fmt.Printf("[%d] %s | %s | %.2f ₽\n", p.ID, p.Name, p.Category, p.Price)
	}
}

func listOrders(rp *repo.Repository) {
	items, err := rp.ListAllOrders(context.Background())
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Println("\n--- Список заказов ---")
	if len(items) == 0 {
		fmt.Println("Заказов пока нет")
		return
	}

	for _, o := range items {
		fmt.Printf("%s | %s | %s | %.2f ₽ | %s\n",
			o.OrderNumber, o.ContactName, o.Status, o.TotalAmount, o.CreatedAt.Format("02.01.2006 15:04"))
	}
}

func createOrder(rp *repo.Repository, reader *bufio.Reader) {
	fmt.Println("\n--- Создание заказа ---")

	fmt.Print("ФИО: ")
	name := readLine(reader)

	fmt.Print("Телефон: ")
	phone := readLine(reader)

	fmt.Print("Комментарий: ")
	comment := readLine(reader)

	products, err := rp.ListProducts(context.Background())
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Println("\nДоступные изделия:")
	for _, p := range products {
		fmt.Printf("%d) %s — %.2f ₽\n", p.ID, p.Name, p.Price)
	}

	var items []models.OrderItem
	for {
		fmt.Print("ID изделия (0 = закончить): ")
		idStr := readLine(reader)

		id, _ := strconv.ParseInt(idStr, 10, 64)
		if id == 0 {
			break
		}

		product, err := rp.GetProduct(context.Background(), id)
		if err != nil {
			fmt.Println("Изделие не найдено")
			continue
		}

		fmt.Print("Количество: ")
		qty, _ := strconv.Atoi(readLine(reader))
		if qty <= 0 {
			fmt.Println("Количество должно быть больше 0")
			continue
		}

		items = append(items, models.OrderItem{
			ProductID: product.ID,
			Quantity:  qty,
			UnitPrice: product.Price,
			Subtotal:  product.Price * float64(qty),
		})
	}

	if len(items) == 0 {
		fmt.Println("Заказ пустой")
		return
	}

	order, err := rp.CreateGuestOrder(context.Background(), name, phone, comment, items)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Printf("Создан заказ: %s | сумма %.2f ₽\n", order.OrderNumber, order.TotalAmount)
}

func showOrder(rp *repo.Repository, reader *bufio.Reader) {
	fmt.Print("Введите номер заказа: ")
	number := readLine(reader)

	order, err := rp.GetOrderWithItems(context.Background(), number)
	if err != nil {
		log.Println("Заказ не найден")
		return
	}

	fmt.Println("\n--- Карточка заказа ---")
	fmt.Println("Номер:", order.OrderNumber)
	fmt.Println("Клиент:", order.ContactName)
	fmt.Println("Телефон:", order.ContactPhone)
	fmt.Println("Статус:", order.Status)
	fmt.Printf("Сумма: %.2f ₽\n", order.TotalAmount)
	fmt.Println("Комментарий:", order.Comment)
	fmt.Println("Состав:")

	for _, it := range order.Items {
		fmt.Printf("- %s | %d шт. | %.2f ₽ | %.2f ₽\n",
			it.ProductName, it.Quantity, it.UnitPrice, it.Subtotal)
	}
}

func updateStatus(rp *repo.Repository, reader *bufio.Reader) {
	fmt.Print("Номер заказа: ")
	number := readLine(reader)

	fmt.Print("Новый статус (new/processing/ready/shipped/completed/cancelled): ")
	status := strings.TrimSpace(readLine(reader))

	fmt.Print("Комментарий: ")
	comment := readLine(reader)

	if err := rp.UpdateOrderStatus(context.Background(), number, status, comment); err != nil {
		log.Println(err)
		return
	}

	fmt.Println("Статус обновлён")
}

func showCompany(rp *repo.Repository) {
	company, err := rp.GetCompanyInfo(context.Background())
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Println("\n--- Данные компании ---")
	fmt.Println("Название:", company.Name)
	fmt.Println("Адрес:", company.Address)
	fmt.Println("Телефон:", company.Phone)
	fmt.Printf("Координаты: %.6f, %.6f\n", company.Lat, company.Lon)
}

func readLine(r *bufio.Reader) string {
	s, _ := r.ReadString('\n')
	return strings.TrimSpace(s)
}
