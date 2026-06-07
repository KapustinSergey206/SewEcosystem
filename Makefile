.PHONY: run web ais bot deps

deps:
	go mod tidy

web:
	go run ./cmd/web

ais:
	go run ./cmd/ais

bot:
	go run ./cmd/bot

run: web
