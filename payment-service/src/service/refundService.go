package service

import (
	"errors"
	"payment-service/models"
	"payment-service/repository"

	"github.com/google/uuid"
)

type RefundService struct {
	Repo *repository.PaymentRepository
}

func NewRefundService(repo *repository.PaymentRepository) *RefundService {
	return &RefundService{}
}

func (s *RefundService) ProcessRefund(req models.RefundRequest) (*models.RefundResponse, error) {
	if s.Repo == nil {
		return nil, errors.New("repository not configured")
	}

	refund := &models.Refund{
		RefundID: uuid.NewString(),
		OrderID: req.OrderID,
		Amount: req.Amount,
		Status: "pending",
		Reason: req.Reason,
	}

	if err := s.Repo.CreateRefundRequest(refund); err != nil {
		return nil, err 
	}

	providerRefID := "prov-ref-" + uuid.NewString()[:8]
	status := "succeeded"

	if err := s.Repo.UpdateRefundResult(refund.RefundID, status, &providerRefID); err != nil {
		return nil, err 
	}

	refund.Status = status
	return &models.RefundResponse{
		RefundID: refund.RefundID,
		Status: status,
		Message: "Refund processed successfully",
	}, nil
}