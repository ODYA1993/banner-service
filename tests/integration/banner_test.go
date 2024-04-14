package banner_test

import (
	"banner-service/internal/config"
	"banner-service/internal/models/banner"
	"banner-service/internal/models/banner/dbbanner"
	"banner-service/pkg/db/postgresql"
	"banner-service/pkg/logging"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

var configPath = "../../configureTestDB.yml"

type fakeAdminChecker struct{}

func (f *fakeAdminChecker) CheckIfAdmin(req *http.Request) (bool, error) {
	return true, nil // Всегда возвращаем true
}

func TestGetUserBanner(t *testing.T) {
	adminChecker := &fakeAdminChecker{}

	// Создаем фейковый HTTP-запрос
	req := httptest.NewRequest("GET", "/user_banner?tag_id=1&feature_id=1&use_last_revision=true", nil)

	// Создаем логгер
	logger := logging.GetLogger()

	// Получаем конфигурацию приложения из файла
	cfg := config.GetConfig(logger, configPath)

	testDB, err := postgresql.NewClient(context.TODO(), 3, cfg.Storage)
	if err != nil {
		logger.Fatalf("%v", err)
	}
	defer testDB.Close()

	// Инициализируем репозиторий баннеров с тестовой базой данных
	bannerRepo := dbbanner.NewBannerRepository(testDB, logger)

	cache := banner.NewBannerCache(5 * time.Minute)

	// Создаем экземпляр хэндлера баннеров
	bannerHandler := banner.NewHandler(bannerRepo, logger, cache)

	// Создаем фейковый HTTP-запрос
	w := httptest.NewRecorder()

	// Вызываем функцию обработчика
	bannerHandler.GetUserBanner(w, req, adminChecker)

	// Проверяем код статуса ответа
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d; got %d", http.StatusOK, w.Code)
	}

	// Проверяем, что баннер был получен
	expectedBanner := &banner.Banner{
		ID:       1,
		Title:    "Banner 1",
		Text:     "This is the text of Banner 1",
		URL:      "https://example.com/banner1",
		IsActive: true,
		FeatureID: banner.Feature{
			ID:   1,
			Name: "Feature 1",
		},
		Tags: []banner.Tag{
			{ID: 1, Name: "Tag 1"},
		},
	}
	if !equalBanners(expectedBanner, w.Body.String()) {
		t.Errorf("expected banner %+v; got %+v", expectedBanner, w.Body.String())
	}
}

// Функция для сравнения ожидаемого баннера и полученного из JSON
func equalBanners(expected *banner.Banner, actual string) bool {
	actualBanner := new(banner.Banner)
	if err := json.Unmarshal([]byte(actual), actualBanner); err != nil {
		return false
	}

	return expected.ID == actualBanner.ID &&
		expected.Title == actualBanner.Title &&
		expected.Text == actualBanner.Text &&
		expected.URL == actualBanner.URL &&
		expected.IsActive == actualBanner.IsActive &&
		expected.FeatureID.ID == actualBanner.FeatureID.ID &&
		expected.FeatureID.Name == actualBanner.FeatureID.Name &&
		reflect.DeepEqual(expected.Tags, actualBanner.Tags)
}
