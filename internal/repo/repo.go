package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"sewing-ecosystem/internal/models"
)

type Repository struct{ DB *sql.DB }

func New(db *sql.DB) *Repository { return &Repository{DB: db} }

func (r *Repository) CreateUser(ctx context.Context, u *models.User) error {
	q := `INSERT INTO users (full_name,email,phone,password_hash,role)
	VALUES ($1,$2,$3,$4,$5) RETURNING id,created_at`
	return r.DB.QueryRowContext(
		ctx,
		q,
		strings.TrimSpace(u.FullName),
		strings.ToLower(strings.TrimSpace(u.Email)),
		strings.TrimSpace(u.Phone),
		u.PasswordHash,
		u.Role,
	).Scan(&u.ID, &u.CreatedAt)
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	u := &models.User{}
	q := `SELECT id, full_name, email, phone, password_hash, role, created_at
	      FROM users
	      WHERE email=$1`
	err := r.DB.QueryRowContext(ctx, q, strings.ToLower(strings.TrimSpace(email))).Scan(
		&u.ID, &u.FullName, &u.Email, &u.Phone, &u.PasswordHash, &u.Role, &u.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *Repository) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	u := &models.User{}
	q := `SELECT id, full_name, email, phone, password_hash, role, created_at
	      FROM users
	      WHERE id=$1`
	err := r.DB.QueryRowContext(ctx, q, id).Scan(
		&u.ID, &u.FullName, &u.Email, &u.Phone, &u.PasswordHash, &u.Role, &u.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *Repository) EnsureAdmin(ctx context.Context, fullName, email, phone, passwordHash string) error {
	_, err := r.GetUserByEmail(ctx, email)
	if err == nil {
		return nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	_, err = r.DB.ExecContext(
		ctx,
		`INSERT INTO users (full_name,email,phone,password_hash,role)
		 VALUES ($1,$2,$3,$4,'admin')`,
		strings.TrimSpace(fullName),
		strings.ToLower(strings.TrimSpace(email)),
		strings.TrimSpace(phone),
		passwordHash,
	)
	return err
}

func (r *Repository) ListProducts(ctx context.Context) ([]models.Product, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, sku, name, description, category, price, image_path, image_path_2, image_path_3, is_active
		FROM products
		WHERE is_active=true
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.Product
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(
			&p.ID, &p.SKU, &p.Name, &p.Description, &p.Category, &p.Price,
			&p.ImagePath, &p.ImagePath2, &p.ImagePath3, &p.IsActive,
		); err != nil {
			return nil, err
		}
		items = append(items, p)
	}
	return items, rows.Err()
}

func (r *Repository) GetProduct(ctx context.Context, id int64) (*models.Product, error) {
	var p models.Product
	err := r.DB.QueryRowContext(ctx, `
		SELECT id, sku, name, description, category, price, image_path, image_path_2, image_path_3, is_active
		FROM products
		WHERE id=$1 AND is_active=true
	`, id).Scan(
		&p.ID, &p.SKU, &p.Name, &p.Description, &p.Category, &p.Price,
		&p.ImagePath, &p.ImagePath2, &p.ImagePath3, &p.IsActive,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repository) QuoteItems(ctx context.Context, quantities map[int64]int) ([]models.OrderItemView, float64, error) {
	var result []models.OrderItemView
	var total float64

	for productID, qty := range quantities {
		if qty <= 0 {
			continue
		}

		p, err := r.GetProduct(ctx, productID)
		if err != nil {
			return nil, 0, err
		}

		subtotal := p.Price * float64(qty)
		result = append(result, models.OrderItemView{
			OrderItem: models.OrderItem{
				ProductID: productID,
				Quantity:  qty,
				UnitPrice: p.Price,
				Subtotal:  subtotal,
			},
			ProductName: p.Name,
		})
		total += subtotal
	}

	return result, total, nil
}

func (r *Repository) CreateOrder(ctx context.Context, userID int64, contactName, contactPhone, comment string, items []models.OrderItem) (*models.Order, error) {
	if len(items) == 0 {
		return nil, errors.New("empty order")
	}

	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var total float64
	for _, it := range items {
		total += it.Subtotal
	}

	order := &models.Order{}

	var row *sql.Row
	if userID > 0 {
		row = tx.QueryRowContext(ctx, `
			INSERT INTO orders (user_id, contact_name, contact_phone, status, total_amount, comment)
			VALUES ($1,$2,$3,'new',$4,$5)
			RETURNING id, order_number, created_at, updated_at
		`, userID, strings.TrimSpace(contactName), strings.TrimSpace(contactPhone), total, strings.TrimSpace(comment))
	} else {
		row = tx.QueryRowContext(ctx, `
			INSERT INTO orders (user_id, contact_name, contact_phone, status, total_amount, comment)
			VALUES (NULL,$1,$2,'new',$3,$4)
			RETURNING id, order_number, created_at, updated_at
		`, strings.TrimSpace(contactName), strings.TrimSpace(contactPhone), total, strings.TrimSpace(comment))
	}

	if err := row.Scan(&order.ID, &order.OrderNumber, &order.CreatedAt, &order.UpdatedAt); err != nil {
		return nil, err
	}

	for _, it := range items {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO order_items (order_id, product_id, quantity, unit_price, subtotal)
			VALUES ($1,$2,$3,$4,$5)
		`, order.ID, it.ProductID, it.Quantity, it.UnitPrice, it.Subtotal)
		if err != nil {
			return nil, err
		}
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO order_status_history (order_id, status, comment)
		VALUES ($1,'new','Заказ создан')
	`, order.ID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	order.UserID = userID
	order.ContactName = strings.TrimSpace(contactName)
	order.ContactPhone = strings.TrimSpace(contactPhone)
	order.Status = "new"
	order.TotalAmount = total
	order.Comment = strings.TrimSpace(comment)

	return order, nil
}

func (r *Repository) CreateGuestOrder(ctx context.Context, contactName, contactPhone, comment string, items []models.OrderItem) (*models.Order, error) {
	return r.CreateOrder(ctx, 0, contactName, contactPhone, comment, items)
}

func (r *Repository) ListOrdersForUser(ctx context.Context, userID int64) ([]models.Order, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, order_number, COALESCE(user_id,0), contact_name, contact_phone, status, total_amount, comment, created_at, updated_at
		FROM orders
		WHERE user_id=$1
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.Order
	for rows.Next() {
		var o models.Order
		if err := rows.Scan(
			&o.ID, &o.UserID, &o.OrderNumber, &o.ContactName, &o.ContactPhone,
			&o.Status, &o.TotalAmount, &o.Comment, &o.CreatedAt, &o.UpdatedAt,
		); err == nil {
			// intentionally wrong branch prevented? no
		}
	}
	_ = rows.Close()

	rows, err = r.DB.QueryContext(ctx, `
		SELECT id, order_number, COALESCE(user_id,0), contact_name, contact_phone, status, total_amount, comment, created_at, updated_at
		FROM orders
		WHERE user_id=$1
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items = nil
	for rows.Next() {
		var o models.Order
		if err := rows.Scan(
			&o.ID, &o.OrderNumber, &o.UserID, &o.ContactName, &o.ContactPhone,
			&o.Status, &o.TotalAmount, &o.Comment, &o.CreatedAt, &o.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, o)
	}
	return items, rows.Err()
}

func (r *Repository) ListAllOrders(ctx context.Context) ([]models.Order, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, order_number, COALESCE(user_id,0), contact_name, contact_phone, status, total_amount, comment, created_at, updated_at
		FROM orders
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.Order
	for rows.Next() {
		var o models.Order
		if err := rows.Scan(
			&o.ID, &o.OrderNumber, &o.UserID, &o.ContactName, &o.ContactPhone,
			&o.Status, &o.TotalAmount, &o.Comment, &o.CreatedAt, &o.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, o)
	}
	return items, rows.Err()
}

func (r *Repository) GetOrderWithItems(ctx context.Context, orderNumber string) (*models.OrderWithItems, error) {
	var o models.OrderWithItems

	err := r.DB.QueryRowContext(ctx, `
		SELECT id, order_number, COALESCE(user_id,0), contact_name, contact_phone, status, total_amount, comment, created_at, updated_at
		FROM orders
		WHERE order_number=$1
	`, strings.TrimSpace(orderNumber)).Scan(
		&o.ID, &o.OrderNumber, &o.UserID, &o.ContactName, &o.ContactPhone,
		&o.Status, &o.TotalAmount, &o.Comment, &o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	rows, err := r.DB.QueryContext(ctx, `
		SELECT oi.id, oi.order_id, oi.product_id, oi.quantity, oi.unit_price, oi.subtotal, p.name
		FROM order_items oi
		JOIN products p ON p.id = oi.product_id
		WHERE oi.order_id=$1
		ORDER BY oi.id
	`, o.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var it models.OrderItemView
		if err := rows.Scan(
			&it.ID, &it.OrderID, &it.ProductID, &it.Quantity, &it.UnitPrice, &it.Subtotal, &it.ProductName,
		); err != nil {
			return nil, err
		}
		o.Items = append(o.Items, it)
	}

	return &o, rows.Err()
}

func (r *Repository) UpdateOrderStatus(ctx context.Context, orderNumber, status, comment string) error {
	res, err := r.DB.ExecContext(ctx, `
		UPDATE orders
		SET status=$2, updated_at=now()
		WHERE order_number=$1
	`, strings.TrimSpace(orderNumber), strings.TrimSpace(status))
	if err != nil {
		return err
	}

	affected, _ := res.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows
	}

	var orderID int64
	if err := r.DB.QueryRowContext(ctx, `
		SELECT id FROM orders WHERE order_number=$1
	`, strings.TrimSpace(orderNumber)).Scan(&orderID); err != nil {
		return err
	}

	_, err = r.DB.ExecContext(ctx, `
		INSERT INTO order_status_history (order_id, status, comment)
		VALUES ($1,$2,$3)
	`, orderID, strings.TrimSpace(status), strings.TrimSpace(comment))
	return err
}

func (r *Repository) SearchOrderStatus(ctx context.Context, orderNumber string) (string, error) {
	var status string
	err := r.DB.QueryRowContext(ctx, `
		SELECT status
		FROM orders
		WHERE order_number=$1
	`, strings.TrimSpace(orderNumber)).Scan(&status)
	return status, err
}

func (r *Repository) GetCompanyInfo(ctx context.Context) (*models.CompanyInfo, error) {
	var c models.CompanyInfo
	err := r.DB.QueryRowContext(ctx, `
		SELECT id, name, address, phone, lat, lon
		FROM company_info
		ORDER BY id
		LIMIT 1
	`).Scan(&c.ID, &c.Name, &c.Address, &c.Phone, &c.Lat, &c.Lon)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *Repository) ListProductsPaged(ctx context.Context, limit, offset int) ([]models.Product, int, error) {
	var total int

	err := r.DB.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM products
		WHERE is_active = true
	`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, sku, name, description, category, price, image_path, image_path_2, image_path_3, is_active
		FROM products
		WHERE is_active = true
		ORDER BY id
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []models.Product
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(
			&p.ID, &p.SKU, &p.Name, &p.Description, &p.Category, &p.Price,
			&p.ImagePath, &p.ImagePath2, &p.ImagePath3, &p.IsActive,
		); err != nil {
			return nil, 0, err
		}
		items = append(items, p)
	}

	return items, total, rows.Err()
}

func (r *Repository) ListUsers(ctx context.Context) ([]models.User, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, full_name, email, phone, password_hash, role, created_at
		FROM users
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(
			&u.ID, &u.FullName, &u.Email, &u.Phone, &u.PasswordHash, &u.Role, &u.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, u)
	}
	return items, rows.Err()
}

func (r *Repository) DeleteOrder(ctx context.Context, orderNumber string) error {
	res, err := r.DB.ExecContext(ctx, `
		DELETE FROM orders
		WHERE order_number=$1
	`, strings.TrimSpace(orderNumber))
	if err != nil {
		return err
	}

	affected, _ := res.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *Repository) DeleteUser(ctx context.Context, userID int64) error {
	res, err := r.DB.ExecContext(ctx, `
		DELETE FROM users
		WHERE id=$1 AND role <> 'admin'
	`, userID)
	if err != nil {
		return err
	}

	affected, _ := res.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func BuildOrderNumber(id int64) string { return fmt.Sprintf("ORDER-%06d", id) }
