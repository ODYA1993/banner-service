package config

import (
	"banner-service/internal/models/banner"
	"banner-service/internal/models/user"
	"banner-service/pkg/logging"
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/ilyakaznacheev/cleanenv"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type Config struct {
	Port          string        `yaml:"port" env:"PORT" env-default:"8080"`
	IsDebug       *bool         `yaml:"is_debug" env:"IS_DEBUG" env-default:"false"`
	Storage       StorageConfig `yaml:"storage"`
	UserHandler   *user.Handler
	BannerHandler *banner.Handler
}

type StorageConfig struct {
	Host     string `yaml:"host" env:"STORAGE_HOST"`
	Port     string `yaml:"port" env:"STORAGE_PORT"`
	DBname   string `yaml:"dbname" env:"STORAGE_DBNAME"`
	Username string `yaml:"username" env:"STORAGE_USERNAME"`
	Password string `yaml:"password" env:"STORAGE_PASSWORD"`
}

var cfg *Config
var once sync.Once

func GetConfig(logger *logging.Logger, configPath string) *Config {
	once.Do(func() {
		logger.Info("read application configuration")
		cfg = &Config{}
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			logger.Fatal("configuration file not found")
		}
		if err := cleanenv.ReadConfig(configPath, cfg); err != nil {
			help, _ := cleanenv.GetDescription(cfg, nil)
			logger.Info(help)
			logger.Fatal(err)
		}

	})
	return cfg
}

func (c *Config) Start(logger *logging.Logger) error {
	logger.Infof("starting server on port: %s", c.Port)

	logger.Info("configure router")
	router := c.ConfigureRouter()

	listener, err := net.Listen("tcp", ":"+c.Port)
	if err != nil {
		return err
	}
	server := &http.Server{
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	// Создаем канал для получения сигналов завершения работы приложения
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, syscall.SIGINT, syscall.SIGTERM)

	// Запускаем сервер в горутине
	go func() {
		if err := server.Serve(listener); err != http.ErrServerClosed {
			logger.Fatalf("server stopped with error: %v", err)
		}
	}()

	// Ждем сигнала завершения работы приложения
	<-shutdownChan

	// Начинаем завершение работы сервера
	logger.Info("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Errorf("server shutdown with error: %v", err)
	}

	// Закрываем ресурсы приложения
	logger.Info("closing application resources...")
	c.CloseCache()

	logger.Info("application stopped")
	return nil
}

func (c *Config) CloseCache() {
	// Закрываем кэш
	if c.BannerHandler.Cache != nil {
		c.BannerHandler.Cache.Close()
	}
}

func (c *Config) ConfigureRouter() chi.Router {
	router := chi.NewRouter()

	router.Use(middleware.Logger)

	router.Post("/register", c.UserHandler.Register)
	router.Post("/login", c.UserHandler.Login)
	router.Get("/users", c.UserHandler.GetUsers)
	router.Get("/user/{id}", c.UserHandler.GetUserByID)
	router.Delete("/delete/{id}", c.UserHandler.DeleteUser)

	router.Get("/user_banner", func(w http.ResponseWriter, r *http.Request) {
		c.BannerHandler.GetUserBanner(w, r, &banner.RealAdminChecker{}) // Здесь мы передаем fakeAdminChecker
	})
	router.Get("/banner/{feature_id}/{tag_id}/{limit}/{offset}", jwtMiddleware(c.BannerHandler.GetBannerFilter))
	router.Post("/banner", jwtMiddleware(c.BannerHandler.CreateBannerHandler))
	router.Put("/banner/{id}", jwtMiddleware(c.BannerHandler.UpdateBannerHandler))
	router.Delete("/delete-banner/{id}", jwtMiddleware(c.BannerHandler.DeleteBannerHandler))
	return router
}
