package main

import (
	"banner-service/internal/config"
	"banner-service/internal/models/banner"
	"banner-service/internal/models/banner/dbbanner"
	"banner-service/internal/models/user"
	"banner-service/internal/models/user/dbuser"
	"banner-service/pkg/db/postgresql"
	"banner-service/pkg/logging"
	"context"
	"time"
)

const configPath = "config.yml"

func main() {
	// Получаем логгер
	logger := logging.GetLogger()

	// Получаем конфигурацию приложения
	cfg := config.GetConfig(logger, configPath)

	// Инициализируем клиент PostgreSQL
	clientPostgreSQL, err := postgresql.NewClient(context.TODO(), 3, cfg.Storage)
	if err != nil {
		logger.Fatalf("%v", err)
	}
	defer clientPostgreSQL.Close() // Закрываем соединение с базой данных при завершении

	// Инициализируем и присваиваем обработчик пользователей из пакета user
	cfg.UserHandler = user.NewHandler(dbuser.NewUserRepository(clientPostgreSQL, logger), logger)

	// Инициализируем и присваиваем обработчик баннеров из пакета banner
	cfg.BannerHandler = banner.NewHandler(dbbanner.NewBannerRepository(clientPostgreSQL, logger), logger, banner.NewBannerCache(5*time.Minute))

	// Запускаем приложение
	err = cfg.Start(logger)
	if err != nil {
		logger.Fatalf("%v", err)
	}
}
