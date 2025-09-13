package repository

import (
	"payment-service/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PaymentRepository struct {
	DB *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) *PaymentRepository {
	return &PaymentRepository{DB: db}
}

func (r *PaymentRepository) Migrate() error {
	return r.DB.AutoMigrate(&models.Payment{}, &models.Refund{}, &models.Transaction{})
}

func (r *PaymentRepository) SavePayment(payment *models.Payment) error {
	return r.DB.Create(payment).Error
}

func (r *PaymentRepository) GetByOrderID(orderID string) (*models.Payment, error) {
	var p models.Payment
	if err := r.DB.Where("order_id = ?", orderID).First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PaymentRepository) CreateTransaction(txn *models.Transaction) error {
	return r.DB.Create(txn).Error
}

func (r *PaymentRepository) GetTransactionByID(transactionID string) (*models.Transaction, error) {
	var t models.Transaction
	if err := r.DB.Where("transaction_id = ?", transactionID).First(&t).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *PaymentRepository) MarkAuthByOrderID(orderID, providerID string) error {
	// create or update payment and transaction
	return r.DB.Transaction(func(tx *gorm.DB) error {
		// upsert payment
		var p models.Payment
		if err := tx.Where("order_id = ?", orderID).First(&p).Error; err != nil {
			// create
			p = models.Payment{OrderID: orderID, Status: "authorized", ProviderID: &providerID, TransactionID: providerID}
			if err := tx.Create(&p).Error; err != nil {
				return err
			}
		} else {
			if err := tx.Model(&models.Payment{}).Where("order_id = ?", orderID).Updates(map[string]interface{}{"status": "authorized", "provider_id": providerID, "transaction_id": providerID}).Error; err != nil {
				return err
			}
		}

		// create transaction record
		txn := models.Transaction{
			TransactionID: providerID,
			OrderID:       orderID,
			ProviderID:    &providerID,
			Amount:        p.Amount,
			Currency:      p.Currency,
			Status:        "authorized",
		}
		if err := tx.Create(&txn).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *PaymentRepository) MarkCapturedByOrderID(orderID, providerID string) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Payment{}).Where("order_id = ?", orderID).Updates(map[string]interface{}{"status": "captured", "transaction_id": providerID}).Error; err != nil {
			return err
		}
		// update transaction status
		if err := tx.Model(&models.Transaction{}).Where("order_id = ?", orderID).Where("transaction_id = ?", providerID).Update("status", "captured").Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *PaymentRepository) MarkFailedByOrderID(orderID, providerID, reason string) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Payment{}).Where("order_id = ?", orderID).Updates(map[string]interface{}{"status": "failed", "transaction_id": providerID}).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.Transaction{}).Where("order_id = ?", orderID).Where("transaction_id = ?", providerID).Update("status", "failed").Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *PaymentRepository) UpdateStatus(orderID, status string, providerID, transactionID *string) error {
	updates := map[string]interface{}{"status": status}
	if providerID != nil {
		updates["provider_id"] = *providerID
	}
	if transactionID != nil {
		updates["transaction_id"] = *transactionID
	}
	return r.DB.Model(&models.Payment{}).Where("order_id = ?", orderID).Updates(updates).Error
}

func (r *PaymentRepository) CreateRefundRequest(refund *models.Refund) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		// Lock payment row
		var p models.Payment
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("order_id = ?", refund.OrderID).First(&p).Error; err != nil {
			return err
		}

		// Create refund record
		if err := tx.Create(refund).Error; err != nil {
			return err
		}

		// Update payment status
		if err := tx.Model(&models.Payment{}).Where("order_id = ?", refund.OrderID).Update("status", "refund_pending").Error; err != nil {
			return err
		}

		return nil
	})
}
func (r *PaymentRepository) UpdateRefundResult(refundID, status string, providerRefID *string) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		var ref models.Refund
		if err := tx.Where("refund_id = ?", refundID).First(&ref).Error; err != nil {
			return err
		}

		if ref.Status == "succeeded" || ref.Status == "failed" {
			return nil
		}

		updates := map[string]interface{}{"status": status}
		if providerRefID != nil {
			updates["provider_ref_id"] = *providerRefID
		}

		if err := tx.Model(&models.Refund{}).Where("refund_id = ?", refundID).Updates(updates).Error; err != nil {
			return err
		}

		if status == "succeeded" {
			if err := tx.Model(&models.Payment{}).Where("order_id = ?", ref.OrderID).Updates(map[string]interface{}{
				"status":         "refunded",
				"transaction_id": providerRefID,
			}).Error; err != nil {
				return err
			}
		} else if status == "failed" {
			if err := tx.Model(&models.Payment{}).Where("order_id = ?", ref.OrderID).Update("status", "refund_failed").Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *PaymentRepository) GetRefundByRefundID(refundID string) (*models.Refund, error) {
	var ref models.Refund
	if err := r.DB.Where("refund_id = ?", refundID).First(&ref).Error; err != nil {
		return nil, err
	}

	return &ref, nil
}
