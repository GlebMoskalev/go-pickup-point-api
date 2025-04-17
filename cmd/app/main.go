package main

import "github.com/GlebMoskalev/go-pickup-point-api/internal/app"

const configPath = "config/config.yaml"

// @title Pickup Point API
// @version 1.0.0
// @description Сервис для управления ПВЗ и приемкой товаров

// @schemes http

// @securityDefinitions.apikey JWT
// @in header
// @name Authorization

func main() {
	app.Run(configPath)
}
