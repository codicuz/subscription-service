package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"github.com/codicuz/subscription-service/internal/model"
	"github.com/codicuz/subscription-service/internal/repository"
	"github.com/codicuz/subscription-service/internal/validator"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type SubscriptionHandler struct {
	repo   *repository.SubscriptionRepository
	logger *slog.Logger
}

func NewSubscriptionHandler(repo *repository.SubscriptionRepository, logger *slog.Logger) *SubscriptionHandler {
	return &SubscriptionHandler{
		repo:   repo,
		logger: logger,
	}
}

func (h *SubscriptionHandler) RegisterRoutes(r chi.Router) {
	r.Route("/subscriptions", func(r chi.Router) {
		r.Post("/", h.Create)
		r.Get("/", h.List)
		r.Get("/{id}", h.GetByID)
		r.Put("/{id}", h.Update)
		r.Delete("/{id}", h.Delete)
		r.Get("/report", h.GetReport)
	})
}

func (h *SubscriptionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateSubscriptionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("ошибка декодирования запроса", "error", err)
		respondWithError(w, http.StatusBadRequest, "неверный формат запроса")
		return
	}

	if err := validator.ValidateCreateSubscription(&req); err != nil {
		h.logger.Warn("ошибка валидации", "error", err)
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	startDate, err := model.ParseDate(req.StartDate)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "неверный формат start_date, ожидается MM-YYYY")
		return
	}

	var endDate *model.CustomDate
	if req.EndDate != "" {
		parsedEndDate, err := model.ParseDate(req.EndDate)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "неверный формат end_date, ожидается MM-YYYY")
			return
		}
		endDate = &parsedEndDate
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "неверный формат user_id, ожидается UUID")
		return
	}

	sub := &model.Subscription{
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      userID,
		StartDate:   startDate,
		EndDate:     endDate,
	}

	created, err := h.repo.Create(r.Context(), sub)
	if err != nil {
		h.logger.Error("ошибка создания подписки", "error", err)
		respondWithError(w, http.StatusInternalServerError, "ошибка при создании подписки")
		return
	}

	h.logger.Info("подписка создана", "id", created.ID)
	respondWithJSON(w, http.StatusCreated, created)
}

func (h *SubscriptionHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "неверный формат ID, ожидается UUID")
		return
	}

	sub, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		h.logger.Error("ошибка получения подписки", "error", err)
		respondWithError(w, http.StatusInternalServerError, "ошибка при получении подписки")
		return
	}

	if sub == nil {
		respondWithError(w, http.StatusNotFound, "подписка не найдена")
		return
	}

	respondWithJSON(w, http.StatusOK, sub)
}

func (h *SubscriptionHandler) List(w http.ResponseWriter, r *http.Request) {
	limit := 10
	offset := 0

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsedLimit, err := strconv.Atoi(l); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	if o := r.URL.Query().Get("offset"); o != "" {
		if parsedOffset, err := strconv.Atoi(o); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	subscriptions, err := h.repo.List(r.Context(), limit, offset)
	if err != nil {
		h.logger.Error("ошибка получения списка подписок", "error", err)
		respondWithError(w, http.StatusInternalServerError, "ошибка при получении списка")
		return
	}

	if subscriptions == nil {
		subscriptions = []*model.Subscription{}
	}

	respondWithJSON(w, http.StatusOK, subscriptions)
}

func (h *SubscriptionHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "неверный формат ID, ожидается UUID")
		return
	}

	var req model.UpdateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "неверный формат запроса")
		return
	}

	updated, err := h.repo.Update(r.Context(), id, &req)
	if err != nil {
		if err.Error() == "подписка не найдена" {
			respondWithError(w, http.StatusNotFound, err.Error())
			return
		}
		h.logger.Error("ошибка обновления подписки", "error", err)
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if updated == nil {
		respondWithError(w, http.StatusNotFound, "подписка не найдена")
		return
	}

	respondWithJSON(w, http.StatusOK, updated)
}

func (h *SubscriptionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "неверный формат ID, ожидается UUID")
		return
	}

	if err := h.repo.Delete(r.Context(), id); err != nil {
		if err.Error() == "подписка не найдена" {
			respondWithError(w, http.StatusNotFound, "подписка не найдена")
			return
		}
		h.logger.Error("ошибка удаления подписки", "error", err)
		respondWithError(w, http.StatusInternalServerError, "ошибка при удалении")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *SubscriptionHandler) GetReport(w http.ResponseWriter, r *http.Request) {
	var req model.SubscriptionReportRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "неверный формат запроса")
		return
	}

	if err := validator.ValidateReportRequest(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	report, err := h.repo.GetTotalCost(r.Context(), &req)
	if err != nil {
		h.logger.Error("ошибка получения отчета", "error", err)
		respondWithError(w, http.StatusInternalServerError, "ошибка при получении отчета")
		return
	}

	respondWithJSON(w, http.StatusOK, report)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		slog.Error("ошибка кодирования ответа", "error", err)
	}
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}