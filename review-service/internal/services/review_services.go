package services

import (
	"context"
	"review-service/internal/models"
	"review-service/internal/repository"
	logger "review-service/log"

	"github.com/google/uuid"
)

type ReviewService interface {
	Save(ctx context.Context, review *models.Review) error
	GetByProductID(ctx context.Context, productID string, limit int, lastKey string) ([]models.Review, string, error)
	GetByID(ctx context.Context, id string) (*models.Review, error)
	AddtoSumPending(ctx context.Context, pending models.SumReviewPending) error
}

type reviewServiceImpl struct {
	repo repository.ReviewRepository
}

func NewReviewService(repo repository.ReviewRepository) ReviewService {
	return &reviewServiceImpl{repo: repo}
}

func (s *reviewServiceImpl) Save(ctx context.Context, review *models.Review) error {
	if review.ID == "" {
		review.ID = uuid.New().String()
	}
	if err := s.repo.Create(ctx, *review); err != nil {
		return err 
	}

	pending := models.SumReviewPending{
		ProductID: review.ProductID,
		ReviewID:  review.ID,
		Rating:    review.Rating,
	}
	if err := s.repo.AddtoSumPending(ctx, pending); err != nil {
		logger.Error("Failed to add to sum pending")
		return err 
	}
	return nil 
}

func (s *reviewServiceImpl) GetByProductID(ctx context.Context, productID string, limit int, lastKey string) ([]models.Review, string, error) {
	return s.repo.GetByProductID(ctx, productID, limit, lastKey)
}

func (s *reviewServiceImpl) GetByID(ctx context.Context, id string) (*models.Review, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *reviewServiceImpl) AddtoSumPending(ctx context.Context, pending models.SumReviewPending) error {
	return s.repo.AddtoSumPending(ctx, pending)
}