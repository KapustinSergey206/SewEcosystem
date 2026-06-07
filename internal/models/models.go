package models

import "time"

type User struct {
	ID           int64
	FullName     string
	Email        string
	Phone        string
	PasswordHash string
	Role         string
	CreatedAt    time.Time
}

type Product struct {
	ID          int64
	SKU         string
	Name        string
	Description string
	Category    string
	Price       float64
	ImagePath   string
	IsActive    bool
}

type Order struct {
	ID           int64
	OrderNumber  string
	UserID       int64
	ContactName  string
	ContactPhone string
	Status       string
	TotalAmount  float64
	Comment      string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type OrderItem struct {
	ID        int64
	OrderID   int64
	ProductID int64
	Quantity  int
	UnitPrice float64
	Subtotal  float64
}

type OrderWithItems struct {
	Order
	Items []OrderItemView
}

type OrderItemView struct {
	OrderItem
	ProductName string
}

type CompanyInfo struct {
	ID      int64
	Name    string
	Address string
	Phone   string
	Lat     float64
	Lon     float64
}

type PassNoHash struct {
	Login    string
	Password string
}
