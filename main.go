package main

import (
	"cart-order-service/config"
	cartHandler "cart-order-service/handlers/cart"
	"cart-order-service/repository/cart"
	"cart-order-service/repository/order"
	"cart-order-service/routes"
	cartUsecase "cart-order-service/usecase/cart"
	"database/sql"
	"fmt"
	"os"
	"time"

	orderHandler "cart-order-service/handlers/order"
	orderUseCase "cart-order-service/usecase/order"

	"github.com/go-playground/validator"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Get the current date
	currentDate := time.Now().Format("2006-01-02")

	// Create a directory for the logs if it doesn't exist
	logDir := "logs"
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		err := os.Mkdir(logDir, 0755)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to create log directory")
		}
	}

	// Create a file to write the log messages to
	logFilePath := fmt.Sprintf("%s/%s.log", logDir, currentDate)
	f, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to open log file")
	}
	defer f.Close()

	// Create a multi-level writer that writes to both the console and the file
	writers := zerolog.MultiLevelWriter(os.Stdout, f)

	// Create a logger instance with the multi-level writer
	logger := zerolog.New(writers).With().Timestamp().Caller().Logger()

	cfg, err := config.LoadConfig()
	if err != nil {
		return
	}

	sqlDb, err := config.ConnectToDatabase(config.Connection{
		Host:     cfg.DBHost,
		Port:     cfg.DBPort,
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
		DBName:   cfg.DBName,
	})
	if err != nil {
		return
	}
	defer sqlDb.Close()

	validator := validator.New()

	routes := setupRoutes(sqlDb, validator, logger)
	routes.Run(cfg.AppPort)
}

func setupRoutes(db *sql.DB, validator *validator.Validate, logger zerolog.Logger) *routes.Routes {

	cartRepository := cart.NewStore(db, logger)
	cartUseCase := cartUsecase.NewCart(cartRepository, logger)
	cartHandler := cartHandler.NewHandler(cartUseCase, logger)

	orderRepository := order.NewStore(db, logger)
	orderUseCase := orderUseCase.NewOrder(orderRepository, logger)
	orderHandler := orderHandler.NewHandler(orderUseCase, validator, logger)

	return &routes.Routes{
		Cart:  cartHandler,
		Order: orderHandler,
	}
}
