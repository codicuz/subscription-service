package validator

import (
	"fmt"
	"strings"
	"github.com/codicuz/subscription-service/internal/model"
	"time"

	"github.com/google/uuid"
)

func ValidateCreateSubscription(req *model.CreateSubscriptionRequest) error {
	if strings.TrimSpace(req.ServiceName) == "" {
		return fmt.Errorf("service_name не может быть пустым")
	}

	if req.Price <= 0 {
		return fmt.Errorf("price должен быть положительным числом")
	}

	if _, err := uuid.Parse(req.UserID); err != nil {
		return fmt.Errorf("неверный формат user_id, ожидается UUID")
	}

	startDate, err := model.ParseDate(req.StartDate)
	if err != nil {
		return fmt.Errorf("неверный формат start_date, ожидается MM-YYYY")
	}

	if req.EndDate != "" {
		endDate, err := model.ParseDate(req.EndDate)
		if err != nil {
			return fmt.Errorf("неверный формат end_date, ожидается MM-YYYY")
		}

		startTime := time.Date(startDate.Year, time.Month(startDate.Month), 1, 0, 0, 0, 0, time.UTC)
		endTime := time.Date(endDate.Year, time.Month(endDate.Month), 1, 0, 0, 0, 0, time.UTC)

		if endTime.Before(startTime) {
			return fmt.Errorf("end_date не может быть раньше start_date")
		}
	}

	return nil
}

func ValidateReportRequest(req *model.SubscriptionReportRequest) error {
	if _, err := model.ParseDate(req.StartDate); err != nil {
		return fmt.Errorf("неверный формат start_date, ожидается MM-YYYY")
	}

	if _, err := model.ParseDate(req.EndDate); err != nil {
		return fmt.Errorf("неверный формат end_date, ожидается MM-YYYY")
	}

	if req.UserID != "" {
		if _, err := uuid.Parse(req.UserID); err != nil {
			return fmt.Errorf("неверный формат user_id, ожидается UUID")
		}
	}

	return nil
}
