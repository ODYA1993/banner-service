package user

import (
	"banner-service/internal/utils"
	"banner-service/pkg/logging"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"time"
)

const contextTimeOut = time.Second * 10

type Handler struct {
	logger     *logging.Logger
	repository Storage
}

func NewHandler(repository Storage, logger *logging.Logger) *Handler {
	return &Handler{
		logger:     logger,
		repository: repository,
	}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var user User

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		h.logger.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = validateUserRegister(user)
	if err != nil {
		h.logger.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		h.logger.Error(err)
		h.logger.Error("failed to generate password hash", err.Error())
		return
	}

	user.Password = string(hashedPassword)

	ctx, cancel := context.WithTimeout(context.Background(), contextTimeOut)
	defer cancel()

	err = h.repository.Create(ctx, &user)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			h.logger.Error(err)
			http.Error(w, "request timeout", http.StatusRequestTimeout)
			return
		}
		h.logger.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.RespondJSON(w, http.StatusCreated, user)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var userCredentials User
	err := json.NewDecoder(r.Body).Decode(&userCredentials)
	if err != nil {
		h.logger.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = validateUser(userCredentials)
	if err != nil {
		h.logger.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), contextTimeOut)
	defer cancel()

	user, err := h.repository.FindOneByEmail(ctx, userCredentials.Email)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			h.logger.Error(err)
			http.Error(w, "request timeout", http.StatusRequestTimeout)
			return
		}
		h.logger.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(userCredentials.Password))
	if err != nil {
		h.logger.Error(err)
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	token, err := newToken(user.ID, user.IsAdmin, time.Hour*24*30)
	if err != nil {
		h.logger.Error(err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	cookie := &http.Cookie{
		Name:     "token",
		Value:    token,
		HttpOnly: true,
		Secure:   false, // Установите этот флаг в true, если ваш сайт использует HTTPS
		Path:     "/",
		MaxAge:   24 * 60 * 60, // Срок действия куки (24 часа)
	}

	http.SetCookie(w, cookie)

	utils.RespondJSON(w, http.StatusOK, user)
}

func (h *Handler) GetUsers(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), contextTimeOut)
	defer cancel()

	users, err := h.repository.FindAll(ctx)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			h.logger.Error(err.Error())
			http.Error(w, "request timeout", http.StatusRequestTimeout)
			return
		}
		h.logger.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.RespondJSON(w, http.StatusOK, users)
}

func (h *Handler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	ctx, cancel := context.WithTimeout(context.Background(), contextTimeOut)
	defer cancel()

	user, err := h.repository.FindOne(ctx, id)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			h.logger.Error(err.Error())
			http.Error(w, "request timeout", http.StatusRequestTimeout)
			return
		}
		h.logger.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.RespondJSON(w, http.StatusOK, user)
}

func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	ctx, cancel := context.WithTimeout(context.Background(), contextTimeOut)
	defer cancel()

	rowsAffected, err := h.repository.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			h.logger.Error(err.Error())
			http.Error(w, "request timeout", http.StatusRequestTimeout)
			return
		}
		h.logger.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		http.Error(w, "user does not exist", http.StatusNotFound)
		return
	}

	response := map[string]string{"message": fmt.Sprintf("user with ID (id %s) deleted", id)}
	utils.RespondJSON(w, http.StatusOK, response)
}

func validateUser(user User) error {
	validate := validator.New()
	err := validate.Struct(user)
	if err != nil {
		return err
	}
	return nil
}

func validateUserRegister(user User) error {
	if user.Name == "" || user.Email == "" || user.Password == "" {
		return errors.New("missing required fields")
	}
	return nil
}

type CustomClaims struct {
	jwt.RegisteredClaims
	ID      int  `json:"id,omitempty"`
	IsAdmin bool `json:"is_admin,omitempty"`
}

func (c *CustomClaims) Valid() error {
	return nil
}

func newToken(userID int, isAdmin bool, expirationTime time.Duration) (string, error) {
	claims := CustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expirationTime)),
		},
		ID:      userID,
		IsAdmin: isAdmin,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	secretKey := []byte("mysecretkey")
	return token.SignedString(secretKey)
}
