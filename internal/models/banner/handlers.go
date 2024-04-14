package banner

import (
	"banner-service/internal/models/user"
	"banner-service/internal/utils"
	"banner-service/pkg/logging"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type Handler struct {
	logger     *logging.Logger
	repository Storage
	Cache      *CacheBanner
}

func NewHandler(repository Storage, logger *logging.Logger, cache *CacheBanner) *Handler {
	return &Handler{
		logger:     logger,
		repository: repository,
		Cache:      cache,
	}
}

type AdminChecker interface {
	CheckIfAdmin(r *http.Request) (bool, error)
}

type RealAdminChecker struct{}

func (h *Handler) GetUserBanner(w http.ResponseWriter, r *http.Request, adminChecker AdminChecker) {
	isAdmin, err := adminChecker.CheckIfAdmin(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	tagIDStr := r.URL.Query().Get("tag_id")
	featureIDStr := r.URL.Query().Get("feature_id")
	useLastRevisionStr := r.URL.Query().Get("use_last_revision")
	keyCache := fmt.Sprintf("%s-%s", tagIDStr, featureIDStr)

	tagID, err := strconv.Atoi(tagIDStr)
	if err != nil {
		http.Error(w, "Invalid tag_id parameter", http.StatusBadRequest)
		return
	}
	featureID, err := strconv.Atoi(featureIDStr)
	if err != nil {
		http.Error(w, "Invalid feature_id parameter", http.StatusBadRequest)
		return
	}
	useLastRevision, err := strconv.ParseBool(useLastRevisionStr)
	if err != nil {
		useLastRevision = false
	}

	ctx, cancel := getContextTimeout()
	defer cancel()

	var banner *Banner
	var ok bool

	if useLastRevision {
		banner, err = h.repository.GetBannerFromDB(ctx, tagID, featureID, useLastRevision)
	} else {
		banner, ok = h.Cache.GetBanner(ctx, keyCache)
		if !ok {
			banner, err = h.repository.GetBannerFromDB(ctx, tagID, featureID, useLastRevision)
		}
	}

	if err != nil {
		handleErrors(err, h.logger, w)
		return
	}

	if banner == nil || (!isAdmin && !banner.IsActive) {
		http.Error(w, "Banner not available", http.StatusNotFound)
		return
	}

	h.Cache.SetBanner(ctx, keyCache, banner)
	utils.RespondJSON(w, http.StatusOK, banner)
}

func handleErrors(err error, logger *logging.Logger, w http.ResponseWriter) {
	if errors.Is(err, context.DeadlineExceeded) {
		logger.Error(err)
		http.Error(w, "request timeout", http.StatusRequestTimeout)
	} else {
		logger.Error(err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}

//func checkIfAdmin(r *http.Request) (bool, error) {
//	// Получаем куку с именем "token" из запроса
//	cookie, err := r.Cookie("token")
//	if err != nil {
//		return false, errors.New("missing token cookie")
//	}
//
//	var claims user.CustomClaims
//
//	tokenString, err := url.QueryUnescape(cookie.Value)
//	if err != nil {
//		return false, errors.New("invalid token")
//	}
//
//	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
//		return []byte("mysecretkey"), nil
//	})
//	if err != nil {
//		if err == jwt.ErrSignatureInvalid {
//			return false, errors.New("invalid token signature")
//		}
//		return false, err
//	}
//
//	if !token.Valid {
//		return false, errors.New("token expired")
//	}
//
//	return claims.IsAdmin, nil
//}

func (h *Handler) GetBannerFilter(w http.ResponseWriter, r *http.Request) {
	tagIDStr := chi.URLParam(r, "tag_id")
	featureIDStr := chi.URLParam(r, "feature_id")
	limitStr := chi.URLParam(r, "limit")
	offsetStr := chi.URLParam(r, "offset")

	tagID, err := strconv.Atoi(tagIDStr)
	if err != nil {
		http.Error(w, "Invalid tag_id parameter", http.StatusBadRequest)
		return
	}
	featureID, err := strconv.Atoi(featureIDStr)
	if err != nil {
		http.Error(w, "Invalid feature_id parameter", http.StatusBadRequest)
		return
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		http.Error(w, "Invalid limit parameter", http.StatusBadRequest)
		return
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		http.Error(w, "Invalid offset parameter", http.StatusBadRequest)
		return
	}

	ctx, cancel := getContextTimeout()
	defer cancel()

	banners, err := h.repository.GetBannersByFiltering(ctx, featureID, tagID, limit, offset)
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

	utils.RespondJSON(w, http.StatusOK, banners)
}

func (h *Handler) CreateBannerHandler(w http.ResponseWriter, r *http.Request) {
	var banner Banner
	err := json.NewDecoder(r.Body).Decode(&banner)
	if err != nil {
		h.logger.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = validateBanner(banner)
	if err != nil {
		h.logger.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx, cancel := getContextTimeout()
	defer cancel()

	err = h.repository.CreateBanner(ctx, &banner)
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

	utils.RespondJSON(w, http.StatusCreated, banner)
}

func (h *Handler) UpdateBannerHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid tag_id parameter", http.StatusBadRequest)
		return
	}
	var banner Banner
	err = json.NewDecoder(r.Body).Decode(&banner)
	if err != nil {
		h.logger.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	banner.ID = id
	err = validateBanner(banner)
	if err != nil {
		h.logger.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx, cancel := getContextTimeout()
	defer cancel()

	err = h.repository.UpdateBanner(ctx, &banner)
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

	utils.RespondJSON(w, http.StatusCreated, banner)
}

func (h *Handler) DeleteBannerHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.logger.Error(err)
		http.Error(w, "Invalid banner ID parameter", http.StatusBadRequest)
		return
	}

	ctx, cancel := getContextTimeout()
	defer cancel()

	err = h.repository.DeleteBanner(ctx, id)
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

	response := map[string]string{"message": fmt.Sprintf("banner with ID (id %d) deleted", id)}
	utils.RespondJSON(w, http.StatusOK, response)
}

func validateBanner(banner Banner) error {
	validate := validator.New()
	err := validate.Struct(banner)
	if err != nil {
		return err
	}
	return nil
}

func getContextTimeout() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	return ctx, cancel
}

func (r *RealAdminChecker) CheckIfAdmin(req *http.Request) (bool, error) {
	// Получаем куку с именем "token" из запроса
	cookie, err := req.Cookie("token")
	if err != nil {
		return false, errors.New("missing token cookie")
	}

	var claims user.CustomClaims

	tokenString, err := url.QueryUnescape(cookie.Value)
	if err != nil {
		return false, errors.New("invalid token")
	}

	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte("mysecretkey"), nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return false, errors.New("invalid token signature")
		}
		return false, err
	}

	if !token.Valid {
		return false, errors.New("token expired")
	}

	return claims.IsAdmin, nil
}
