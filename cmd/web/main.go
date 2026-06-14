package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimd "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog/v2"
	"sewing-ecosystem/internal/app"
	"sewing-ecosystem/internal/auth"
	"sewing-ecosystem/internal/handlers"
	"sewing-ecosystem/internal/middleware"
	"sewing-ecosystem/internal/repo"
	"sewing-ecosystem/internal/views"
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
	h := &handlers.Handler{Repo: rp, Views: views.New("web/templates"), Config: cfg}

	logger := httplog.NewLogger("sewing-web", httplog.Options{JSON: false})
	r := chi.NewRouter()
	r.Use(httplog.RequestLogger(logger))
	r.Use(chimd.RealIP, chimd.Recoverer, chimd.Timeout(60*time.Second))
	r.Use(middleware.Auth(cfg.Secret))

	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))
	r.Get("/", h.Home)
	r.Get("/catalog", h.Catalog)
	r.Get("/calculator", h.Calculator)
	r.Post("/calculator", h.Calculator)
	r.Get("/login", h.ShowLogin)
	r.Post("/login", h.Login)
	r.Get("/register", h.ShowRegister)
	r.Post("/register", h.Register)
	r.Post("/logout", h.Logout)

	r.Route("/dashboard", func(sr chi.Router) {
		sr.Use(middleware.RequireAuth)
		sr.Get("/", h.Dashboard)
		sr.Post("/orders", h.CreateOrder)
	})

	r.Route("/admin", func(sr chi.Router) {
		sr.Use(middleware.RequireAuth, middleware.RequireRole("admin"))
		sr.Get("/orders", h.AdminOrders)
		sr.Get("/orders/{orderNumber}", h.AdminOrder)
		sr.Post("/orders/{orderNumber}/status", h.AdminUpdateStatus)
		sr.Post("/orders/{orderNumber}/delete", h.AdminDeleteOrder)
		sr.Get("/users", h.AdminUsers)
		sr.Post("/users/{userID}/delete", h.AdminDeleteUser)
	})

	r.Get("/api/products", h.APIProducts)
	r.Get("/api/orders/{orderNumber}/status", h.APIOrderStatus)

	log.Printf("web running on %s", cfg.Addr)
	log.Fatal(http.ListenAndServe(cfg.Addr, r))
}

func seedAdmin(rp *repo.Repository, cfg app.Config) {
	hash, err := auth.HashPassword(cfg.DefaultAdminPassword)
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rp.EnsureAdmin(ctx, "Главный администратор", cfg.DefaultAdminEmail, "+70000000000", hash); err != nil {
		log.Fatal(err)
	}
}
