package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"payment-service/models"
	"payment-service/repository"
)

type VendorService struct {
	VendorRepo     *repository.VendorRepository
	PaymentService *PaymentService
	Producer       *KafkaProducer
}

func NewVendorService(vendorRepo *repository.VendorRepository, paymentService *PaymentService) *VendorService {
	return &VendorService{
		VendorRepo:     vendorRepo,
		PaymentService: paymentService,
	}
}

// Register new vendor with Stripe Connect account
func (s *VendorService) RegisterVendor(ctx context.Context, req VendorRegistrationRequest) (*VendorRegistrationResponse, error) {
	// Validate input
	if req.VendorID == "" || req.Email == "" || req.Country == "" {
		return nil, errors.New("vendor_id, email, and country are required")
	}

	// Check if vendor already exists
	exists, err := s.VendorRepo.VendorExists(ctx, req.VendorID)
	if err != nil {
		return nil, fmt.Errorf("failed to check vendor existence: %w", err)
	}
	if exists {
		return nil, errors.New("vendor already exists")
	}

	// Create Stripe Connect account
	businessProfile := make(map[string]string)
	if req.BusinessName != "" {
		businessProfile["name"] = req.BusinessName
	}
	if req.BusinessURL != "" {
		businessProfile["url"] = req.BusinessURL
	}
	if req.BusinessMCC != "" {
		businessProfile["mcc"] = req.BusinessMCC
	}

	stripeAccount, err := s.PaymentService.CreateStripeConnectAccount(
		ctx,
		req.VendorID,
		req.Email,
		req.Country,
		businessProfile,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Stripe Connect account: %w", err)
	}

	// Create vendor record in database
	vendor := &models.VendorAccount{
		VendorID:        req.VendorID,
		StripeAccountID: stripeAccount.ID,
		Email:           req.Email,
		Country:         req.Country,
		BusinessName:    &req.BusinessName,
		BusinessURL:     &req.BusinessURL,
		BusinessType:    &req.BusinessType,
		Status:          models.VendorAccountStatusPending,
	}

	// Only set bank info if provided (for direct bank transfers)
	if req.BankName != "" {
		vendor.BankName = &req.BankName
	}
	if req.BankAccountName != "" {
		vendor.BankAccountName = &req.BankAccountName
	}
	if req.BankAccountNumber != "" {
		vendor.BankAccountNumber = &req.BankAccountNumber
	}
	if req.BankRoutingNumber != "" {
		vendor.BankRoutingNumber = &req.BankRoutingNumber
	}
	if req.SwiftCode != "" {
		vendor.SwiftCode = &req.SwiftCode
	}

	if err := s.VendorRepo.CreateVendorAccount(ctx, vendor); err != nil {
		// TODO: Consider rollback Stripe account creation
		return nil, fmt.Errorf("failed to create vendor record: %w", err)
	}

	// Create onboarding link
	baseURL := os.Getenv("FRONTEND_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:3000" // Default for development
	}

	returnURL := fmt.Sprintf("%s/vendor/onboarding/success?vendor_id=%s", baseURL, req.VendorID)
	refreshURL := fmt.Sprintf("%s/vendor/onboarding/refresh?vendor_id=%s", baseURL, req.VendorID)

	accountLink, err := s.PaymentService.CreateAccountLink(
		ctx,
		stripeAccount.ID,
		refreshURL,
		returnURL,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create onboarding link: %w", err)
	}

	// Update with onboarding URL
	s.VendorRepo.UpdateOnboardingStatus(ctx, req.VendorID, false, &accountLink.URL)

	// Send vendor registration event
	if s.Producer != nil {
		event := map[string]interface{}{
			"vendor_id":         req.VendorID,
			"stripe_account_id": stripeAccount.ID,
			"status":            models.VendorAccountStatusPending,
			"event":             "vendor_registered",
		}
		s.Producer.SendVendorAccountUpdate(ctx, event)
	}

	return &VendorRegistrationResponse{
		VendorID:        req.VendorID,
		StripeAccountID: stripeAccount.ID,
		OnboardingURL:   accountLink.URL,
		Status:          models.VendorAccountStatusPending,
		Message:         "Vendor registered successfully. Please complete onboarding.",
	}, nil
}

// Get vendor information
func (s *VendorService) GetVendor(ctx context.Context, vendorID string) (*models.VendorAccount, error) {
	return s.VendorRepo.GetVendorByID(ctx, vendorID)
}

// Update vendor status based on Stripe webhook
func (s *VendorService) UpdateVendorFromStripe(ctx context.Context, stripeAccountID string, accountData map[string]interface{}) error {
	vendor, err := s.VendorRepo.GetVendorByStripeAccountID(ctx, stripeAccountID)
	if err != nil {
		return fmt.Errorf("vendor not found for Stripe account %s: %w", stripeAccountID, err)
	}

	// Extract relevant fields from Stripe account data
	updates := make(map[string]interface{})

	if chargesEnabled, ok := accountData["charges_enabled"].(bool); ok {
		updates["charges_enabled"] = chargesEnabled
	}
	if payoutsEnabled, ok := accountData["payouts_enabled"].(bool); ok {
		updates["payouts_enabled"] = payoutsEnabled
	}
	if detailsSubmitted, ok := accountData["details_submitted"].(bool); ok {
		updates["details_submitted"] = detailsSubmitted
	}

	// Determine status based on capabilities
	if chargesEnabled, _ := accountData["charges_enabled"].(bool); chargesEnabled {
		if payoutsEnabled, _ := accountData["payouts_enabled"].(bool); payoutsEnabled {
			updates["status"] = models.VendorAccountStatusActive
			updates["onboarding_completed"] = true
		} else {
			updates["status"] = models.VendorAccountStatusRestricted
		}
	}

	if len(updates) > 0 {
		err = s.VendorRepo.UpdateVendorAccount(ctx, vendor.VendorID, updates)
		if err != nil {
			return fmt.Errorf("failed to update vendor: %w", err)
		}

		// Send update event
		if s.Producer != nil {
			event := map[string]interface{}{
				"vendor_id":         vendor.VendorID,
				"stripe_account_id": stripeAccountID,
				"status":            updates["status"],
				"event":             "vendor_updated",
			}
			s.Producer.SendVendorAccountUpdate(ctx, event)
		}
	}

	return nil
}

// Create new onboarding link for existing vendor
func (s *VendorService) CreateOnboardingLink(ctx context.Context, vendorID string) (string, error) {
	vendor, err := s.VendorRepo.GetVendorByID(ctx, vendorID)
	if err != nil {
		return "", fmt.Errorf("vendor not found: %w", err)
	}

	baseURL := os.Getenv("FRONTEND_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:3000" // Default for development
	}

	returnURL := fmt.Sprintf("%s/vendor/onboarding/success?vendor_id=%s", baseURL, vendorID)
	refreshURL := fmt.Sprintf("%s/vendor/onboarding/refresh?vendor_id=%s", baseURL, vendorID)

	accountLink, err := s.PaymentService.CreateAccountLink(
		ctx,
		vendor.StripeAccountID,
		refreshURL,
		returnURL,
	)
	if err != nil {
		return "", fmt.Errorf("failed to create onboarding link: %w", err)
	}

	// Update onboarding URL in database
	s.VendorRepo.UpdateOnboardingStatus(ctx, vendorID, false, &accountLink.URL)

	return accountLink.URL, nil
}

// Check vendor payout methods available
func (s *VendorService) GetAvailablePayoutMethods(ctx context.Context, vendorID string) (*PayoutMethodsResponse, error) {
	vendor, err := s.VendorRepo.GetVendorByID(ctx, vendorID)
	if err != nil {
		return nil, fmt.Errorf("vendor not found: %w", err)
	}

	methods := &PayoutMethodsResponse{
		VendorID: vendorID,
		Status:   vendor.Status,
		Methods:  []PayoutMethod{},
	}

	// Stripe Connect method (available if onboarding completed)
	if vendor.OnboardingCompleted && vendor.PayoutsEnabled {
		methods.Methods = append(methods.Methods, PayoutMethod{
			Type:        "stripe_connect",
			Available:   true,
			Description: "Automatic payouts via Stripe Connect",
			Fee:         "Stripe standard fees apply",
		})
	} else {
		methods.Methods = append(methods.Methods, PayoutMethod{
			Type:        "stripe_connect",
			Available:   false,
			Description: "Complete Stripe onboarding to enable",
			Fee:         "Stripe standard fees apply",
		})
	}

	// Direct bank transfer method (available if bank info provided)
	hasBankInfo := vendor.BankAccountNumber != nil && *vendor.BankAccountNumber != ""
	methods.Methods = append(methods.Methods, PayoutMethod{
		Type:        "bank_transfer",
		Available:   hasBankInfo && vendor.Status == models.VendorAccountStatusActive,
		Description: "Direct bank transfer to your account",
		Fee:         "Platform processing fee may apply",
	})

	return methods, nil
}

// Update vendor bank account information
func (s *VendorService) UpdateVendorBankAccount(ctx context.Context, vendorID string, req UpdateBankAccountRequest) error {
	// Validate vendor exists
	vendor, err := s.VendorRepo.GetVendorByID(ctx, vendorID)
	if err != nil {
		return fmt.Errorf("vendor not found: %w", err)
	}

	// Prepare updates
	updates := map[string]interface{}{}
	if req.BankName != "" {
		updates["bank_name"] = req.BankName
	}
	if req.BankAccountName != "" {
		updates["bank_account_name"] = req.BankAccountName
	}
	if req.BankAccountNumber != "" {
		updates["bank_account_number"] = req.BankAccountNumber
	}
	if req.BankRoutingNumber != "" {
		updates["bank_routing_number"] = req.BankRoutingNumber
	}
	if req.SwiftCode != "" {
		updates["swift_code"] = req.SwiftCode
	}

	if len(updates) == 0 {
		return errors.New("no updates provided")
	}

	// Update database
	err = s.VendorRepo.UpdateVendorAccount(ctx, vendorID, updates)
	if err != nil {
		return fmt.Errorf("failed to update bank account: %w", err)
	}

	// Send update event
	if s.Producer != nil {
		event := map[string]interface{}{
			"vendor_id":         vendorID,
			"stripe_account_id": vendor.StripeAccountID,
			"event":             "bank_account_updated",
			"updates":           updates,
		}
		s.Producer.SendVendorAccountUpdate(ctx, event)
	}

	return nil
}

// Get vendor bank account for transfers
func (s *VendorService) GetVendorBankAccount(ctx context.Context, vendorID string) (*VendorBankAccount, error) {
	vendor, err := s.VendorRepo.GetVendorByID(ctx, vendorID)
	if err != nil {
		return nil, fmt.Errorf("vendor not found: %w", err)
	}

	// Only return bank info if vendor is active
	if vendor.Status != models.VendorAccountStatusActive {
		return nil, errors.New("vendor account is not active")
	}

	bankAccount := &VendorBankAccount{
		VendorID:          vendor.VendorID,
		BankName:          safeStringValue(vendor.BankName),
		BankAccountName:   safeStringValue(vendor.BankAccountName),
		BankAccountNumber: safeStringValue(vendor.BankAccountNumber),
		BankRoutingNumber: safeStringValue(vendor.BankRoutingNumber),
		SwiftCode:         safeStringValue(vendor.SwiftCode),
		Country:           vendor.Country,
		Status:            vendor.Status,
	}

	return bankAccount, nil
}

// Helper function to safely get string value from pointer
func safeStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// Request models
type VendorRegistrationRequest struct {
	VendorID     string `json:"vendor_id"`
	Email        string `json:"email"`
	Country      string `json:"country" validate:"required,len=2"` // ISO country code
	BusinessName string `json:"business_name"`
	BusinessURL  string `json:"business_url"`
	BusinessType string `json:"business_type"`
	BusinessMCC  string `json:"business_mcc"` // Merchant Category Code

	// Bank Account Information (OPTIONAL - only for direct bank transfers)
	BankName          string `json:"bank_name,omitempty"`
	BankAccountName   string `json:"bank_account_name,omitempty"`
	BankAccountNumber string `json:"bank_account_number,omitempty"`
	BankRoutingNumber string `json:"bank_routing_number,omitempty"`
	SwiftCode         string `json:"swift_code,omitempty"`
}

type VendorRegistrationResponse struct {
	VendorID        string `json:"vendor_id"`
	StripeAccountID string `json:"stripe_account_id"`
	OnboardingURL   string `json:"onboarding_url"`
	Status          string `json:"status"`
	Message         string `json:"message"`
}

type UpdateBankAccountRequest struct {
	BankName          string `json:"bank_name"`
	BankAccountName   string `json:"bank_account_name"`
	BankAccountNumber string `json:"bank_account_number"`
	BankRoutingNumber string `json:"bank_routing_number,omitempty"`
	SwiftCode         string `json:"swift_code,omitempty"`
}

type VendorBankAccount struct {
	VendorID          string `json:"vendor_id"`
	BankName          string `json:"bank_name"`
	BankAccountName   string `json:"bank_account_name"`
	BankAccountNumber string `json:"bank_account_number"`
	BankRoutingNumber string `json:"bank_routing_number,omitempty"`
	SwiftCode         string `json:"swift_code,omitempty"`
	Country           string `json:"country"`
	Status            string `json:"status"`
}

type PayoutMethodsResponse struct {
	VendorID string         `json:"vendor_id"`
	Status   string         `json:"status"`
	Methods  []PayoutMethod `json:"methods"`
}

type PayoutMethod struct {
	Type        string `json:"type"` // stripe_connect, bank_transfer
	Available   bool   `json:"available"`
	Description string `json:"description"`
	Fee         string `json:"fee"`
}

type VendorAccountUpdateEvent struct {
	VendorID        string `json:"vendor_id"`
	StripeAccountID string `json:"stripe_account_id"`
	Status          string `json:"status"`
	Event           string `json:"event"` // vendor_registered, vendor_updated, onboarding_completed
}

// CreateVendorPayout creates a payout to vendor's Stripe Connect account
func (s *VendorService) CreateVendorPayout(ctx context.Context, vendorID string, orderID string, amount float64) error {
	// Get vendor account
	vendor, err := s.VendorRepo.GetVendorByID(ctx, vendorID)
	if err != nil {
		return fmt.Errorf("vendor not found: %w", err)
	}

	// Check if vendor onboarding is completed
	if !vendor.OnboardingCompleted || !vendor.ChargesEnabled || !vendor.PayoutsEnabled {
		return errors.New("vendor onboarding not completed or payouts not enabled")
	}

	// Create payout via Stripe
	amountInCents := int64(amount * 100)
	payoutID, err := s.PaymentService.CreateStripePayout(ctx, vendor.StripeAccountID, amountInCents, "usd")
	if err != nil {
		return fmt.Errorf("failed to create Stripe payout: %w", err)
	}

	// Record payout in database
	payout := &models.VendorPayout{
		VendorID:       vendorID,
		OrderID:        &orderID,
		Amount:         amount,
		Currency:       "usd",
		Status:         models.PayoutStatusPending,
		PayoutMethod:   "stripe_connect",
		StripePayoutID: &payoutID,
		Description:    fmt.Sprintf("Payout for order %s", orderID),
	}

	if err := s.VendorRepo.CreateVendorPayout(ctx, payout); err != nil {
		return fmt.Errorf("failed to record payout: %w", err)
	}

	return nil
}
