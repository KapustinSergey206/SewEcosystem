package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"sewing-ecosystem/internal/app"
	"sewing-ecosystem/internal/auth"
	"sewing-ecosystem/internal/middleware"
	"sewing-ecosystem/internal/models"
	"sewing-ecosystem/internal/repo"
	"sewing-ecosystem/internal/views"

	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
)

var str = "user=postgres password=73237323Qwa dbname=Sewing sslmode=disable"
var db, err = sql.Open("postgres", str)

type Handler struct {
	Repo   *repo.Repository
	Views  *views.Renderer
	Config app.Config
}

type TemplateData map[string]any

func (h *Handler) base(r *http.Request, title string) TemplateData {
	td := TemplateData{
		"Title":   title,
		"AppName": h.Config.AppName,
	}

	if u, ok := middleware.CurrentUser(r); ok {
		td["CurrentUser"] = u
	}

	if c, err := r.Cookie("flash"); err == nil && c.Value != "" {
		if decoded, err := url.QueryUnescape(c.Value); err == nil {
			td["Flash"] = decoded
		} else {
			td["Flash"] = c.Value
		}
	}

	return td
}

func setFlash(w http.ResponseWriter, msg string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "flash",
		Value:    url.QueryEscape(msg),
		Path:     "/",
		MaxAge:   10,
		HttpOnly: true,
	})
}

func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	products, _ := h.Repo.ListProducts(r.Context())
	td := h.base(r, "Главная")
	td["Products"] = products
	h.Views.Render(w, "home.html", td)
}

func (h *Handler) Catalog(w http.ResponseWriter, r *http.Request) {
	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	const perPage = 8
	offset := (page - 1) * perPage

	products, total, err := h.Repo.ListProductsPaged(r.Context(), perPage, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	totalPages := total / perPage
	if total%perPage != 0 {
		totalPages++
	}
	if totalPages == 0 {
		totalPages = 1
	}

	td := h.base(r, "Каталог")
	td["Products"] = products
	td["CurrentPage"] = page
	td["TotalPages"] = totalPages
	td["HasPrev"] = page > 1
	td["HasNext"] = page < totalPages
	td["PrevPage"] = page - 1
	td["NextPage"] = page + 1

	var pages []int
	for i := 1; i <= totalPages; i++ {
		pages = append(pages, i)
	}
	td["Pages"] = pages

	h.Views.Render(w, "catalog.html", td)
}

func (h *Handler) Calculator(w http.ResponseWriter, r *http.Request) {
	products, _ := h.Repo.ListProducts(r.Context())
	td := h.base(r, "Калькулятор")
	td["Products"] = products

	if r.Method == http.MethodPost {
		_ = r.ParseForm()

		quantities := map[int64]int{}
		for _, p := range products {
			qty, _ := strconv.Atoi(r.FormValue(fmt.Sprintf("qty_%d", p.ID)))
			if qty > 0 {
				quantities[p.ID] = qty
			}
		}

		items, total, err := h.Repo.QuoteItems(r.Context(), quantities)
		if err == nil {
			td["QuoteItems"] = items
			td["Total"] = total
		}
	}

	h.Views.Render(w, "calculator.html", td)
}

func (h *Handler) ShowLogin(w http.ResponseWriter, r *http.Request) {
	h.Views.Render(w, "login.html", h.base(r, "Вход"))
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()

	u, err := h.Repo.GetUserByEmail(r.Context(), strings.TrimSpace(r.FormValue("email")))
	if err != nil || auth.CheckPassword(u.PasswordHash, r.FormValue("password")) != nil {
		setFlash(w, "Неверный email или пароль")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	token, _ := auth.GenerateToken(h.Config.Secret, u.ID, u.Role)

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		HttpOnly: true,
		Path:     "/",
		MaxAge:   7 * 24 * 3600,
	})

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (h *Handler) ShowRegister(w http.ResponseWriter, r *http.Request) {
	h.Views.Render(w, "register.html", h.base(r, "Регистрация"))
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()

	em := strings.TrimSpace(r.FormValue("email"))
	res := r.FormValue("password")
	uP := models.PassNoHash{
		Login:    em,
		Password: res,
	}
	db.Exec("Insert Into passnohash Values($1,$2)", uP.Login, uP.Password)

	passwordHash, err := auth.HashPassword(r.FormValue("password"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	u := &models.User{
		FullName:     strings.TrimSpace(r.FormValue("full_name")),
		Email:        em,
		Phone:        strings.TrimSpace(r.FormValue("phone")),
		PasswordHash: passwordHash,
		Role:         "client",
	}

	if err := h.Repo.CreateUser(r.Context(), u); err != nil {
		setFlash(w, "Не удалось зарегистрироваться. Возможно, email уже используется.")
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}

	token, _ := auth.GenerateToken(h.Config.Secret, u.ID, u.Role)

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		HttpOnly: true,
		Path:     "/",
		MaxAge:   7 * 24 * 3600,
	})

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		HttpOnly: true,
		Path:     "/",
		MaxAge:   -1,
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	u, _ := middleware.CurrentUser(r)

	if u.Role == "admin" {
		http.Redirect(w, r, "/admin/orders", http.StatusSeeOther)
		return
	}

	orders, _ := h.Repo.ListOrdersForUser(r.Context(), u.UserID)
	products, _ := h.Repo.ListProducts(r.Context())

	td := h.base(r, "Личный кабинет")
	td["Orders"] = orders
	td["Products"] = products

	h.Views.Render(w, "dashboard_client.html", td)
}

func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	u, _ := middleware.CurrentUser(r)
	products, _ := h.Repo.ListProducts(r.Context())

	_ = r.ParseForm()

	var items []models.OrderItem
	for _, p := range products {
		qty, _ := strconv.Atoi(r.FormValue(fmt.Sprintf("qty_%d", p.ID)))
		if qty > 0 {
			items = append(items, models.OrderItem{
				ProductID: p.ID,
				Quantity:  qty,
				UnitPrice: p.Price,
				Subtotal:  p.Price * float64(qty),
			})
		}
	}

	if len(items) == 0 {
		setFlash(w, "Добавьте хотя бы одну позицию в заказ")
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	order, err := h.Repo.CreateOrder(
		r.Context(),
		u.UserID,
		r.FormValue("contact_name"),
		r.FormValue("contact_phone"),
		r.FormValue("comment"),
		items,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	setFlash(w, "Заказ успешно создан: "+order.OrderNumber)
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (h *Handler) AdminOrders(w http.ResponseWriter, r *http.Request) {
	orders, _ := h.Repo.ListAllOrders(r.Context())
	td := h.base(r, "Управление заказами")
	td["Orders"] = orders
	h.Views.Render(w, "dashboard_admin.html", td)
}

func (h *Handler) AdminOrder(w http.ResponseWriter, r *http.Request) {
	order, err := h.Repo.GetOrderWithItems(r.Context(), chi.URLParam(r, "orderNumber"))
	if err != nil {
		http.Error(w, "Заказ не найден", http.StatusNotFound)
		return
	}

	td := h.base(r, "Заказ")
	td["Order"] = order

	h.Views.Render(w, "order_admin.html", td)
}

func (h *Handler) AdminUpdateStatus(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()

	status := NormalizeStatus(r.FormValue("status"))
	comment := strings.TrimSpace(r.FormValue("comment"))

	err := h.Repo.UpdateOrderStatus(r.Context(), chi.URLParam(r, "orderNumber"), status, comment)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	setFlash(w, "Статус заказа обновлён")
	http.Redirect(w, r, "/admin/orders/"+chi.URLParam(r, "orderNumber"), http.StatusSeeOther)
}

func (h *Handler) APIProducts(w http.ResponseWriter, r *http.Request) {
	products, err := h.Repo.ListProducts(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write([]byte("["))

	for i, p := range products {
		if i > 0 {
			_, _ = w.Write([]byte(","))
		}
		_, _ = w.Write([]byte(fmt.Sprintf(`{"id":%d,"name":%q,"price":%.2f}`, p.ID, p.Name, p.Price)))
	}

	w.Write([]byte("]"))
}

func (h *Handler) APIOrderStatus(w http.ResponseWriter, r *http.Request) {
	status, err := h.Repo.SearchOrderStatus(r.Context(), chi.URLParam(r, "orderNumber"))
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, _ = w.Write([]byte(fmt.Sprintf(`{"order_number":%q,"status":%q}`, chi.URLParam(r, "orderNumber"), status)))
}

func NormalizeStatus(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))

	switch s {
	case "new", "processing", "ready", "shipped", "completed", "cancelled":
		return s
	default:
		return "processing"
	}
}
