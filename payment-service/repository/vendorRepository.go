package repository

import (
	"context"
	"payment-service/models"

	"gorm.io/gorm"
)

type VendorRepository struct {
	db *gorm.DB
}

func NewVendorRepository(db *gorm.DB) *VendorRepository {
	return &VendorRepository{db: db}
}

// Create new vendor account
func (r *VendorRepository) CreateVendorAccount(ctx context.Context, vendor *models.VendorAccount) error {
	return r.db.WithContext(ctx).Create(vendor).Error
}

// Get vendor by ID
func (r *VendorRepository) GetVendorByID(ctx context.Context, vendorID string) (*models.VendorAccount, error) {
	var vendor models.VendorAccount
	err := r.db.WithContext(ctx).Where("vendor_id = ?", vendorID).First(&vendor).Error
	if err != nil {
		return nil, err
	}
	return &vendor, nil
}

// Get vendor by Stripe account ID
func (r *VendorRepository) GetVendorByStripeAccountID(ctx context.Context, stripeAccountID string) (*models.VendorAccount, error) {
	var vendor models.VendorAccount
	err := r.db.WithContext(ctx).Where("stripe_account_id = ?", stripeAccountID).First(&vendor).Error
	if err != nil {
		return nil, err
	}
	return &vendor, nil
}

// Update vendor account
func (r *VendorRepository) UpdateVendorAccount(ctx context.Context, vendorID string, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&models.VendorAccount{}).Where("vendor_id = ?", vendorID).Updates(updates).Error
}

// Update vendor status
func (r *VendorRepository) UpdateVendorStatus(ctx context.Context, vendorID string, status string) error {
	return r.db.WithContext(ctx).Model(&models.VendorAccount{}).Where("vendor_id = ?", vendorID).Update("status", status).Error
}

// Update onboarding completion
func (r *VendorRepository) UpdateOnboardingStatus(ctx context.Context, vendorID string, completed bool, onboardingURL *string) error {
	updates := map[string]interface{}{
		"onboarding_completed": completed,
	}
	if onboardingURL != nil {
		updates["onboarding_url"] = *onboardingURL
	}
	return r.UpdateVendorAccount(ctx, vendorID, updates)
}

// Update Stripe capabilities
func (r *VendorRepository) UpdateStripeCapabilities(ctx context.Context, vendorID string, chargesEnabled, payoutsEnabled, detailsSubmitted bool) error {
	updates := map[string]interface{}{
		"charges_enabled":   chargesEnabled,
		"payouts_enabled":   payoutsEnabled,
		"details_submitted": detailsSubmitted,
	}
	return r.UpdateVendorAccount(ctx, vendorID, updates)
}

// Get all vendors with pagination
func (r *VendorRepository) GetVendors(ctx context.Context, offset, limit int) ([]*models.VendorAccount, error) {
	var vendors []*models.VendorAccount
	err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&vendors).Error
	return vendors, err
}

// Check if vendor exists
func (r *VendorRepository) VendorExists(ctx context.Context, vendorID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.VendorAccount{}).Where("vendor_id = ?", vendorID).Count(&count).Error
	return count > 0, err
}

// Get vendor's Stripe account ID for payments
func (r *VendorRepository) GetVendorStripeAccountID(ctx context.Context, vendorID string) (string, error) {
	var vendor models.VendorAccount
	err := r.db.WithContext(ctx).Select("stripe_account_id").Where("vendor_id = ? AND status = ?", vendorID, models.VendorAccountStatusActive).First(&vendor).Error
	if err != nil {
		return "", err
	}
	return vendor.StripeAccountID, nil
}

// CreateVendorPayout records a new payout transaction
func (r *VendorRepository) CreateVendorPayout(ctx context.Context, payout *models.VendorPayout) error {
	return r.db.WithContext(ctx).Create(payout).Error
}

// GetVendorPayouts retrieves payout history for a vendor
func (r *VendorRepository) GetVendorPayouts(ctx context.Context, vendorID string, limit, offset int) ([]*models.VendorPayout, error) {
	var payouts []*models.VendorPayout
	err := r.db.WithContext(ctx).
		Where("vendor_id = ?", vendorID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&payouts).Error
	return payouts, err
}

// UpdatePayoutStatus updates the status of a payout
func (r *VendorRepository) UpdatePayoutStatus(ctx context.Context, payoutID uint, status string) error {
	return r.db.WithContext(ctx).
		Model(&models.VendorPayout{}).
		Where("id = ?", payoutID).
		Update("status", status).Error
}
