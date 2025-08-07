package compliance

import (
	"context"
	"crypto-trading-strategies/pkg/types"
	"fmt"
	"time"
)

type ComplianceEngine struct {
	kycProvider   KYCProvider
	amlMonitor    AMLMonitor
	sanctionsDB   SanctionsDatabase
	riskScorer    RiskScorer
	reportManager ReportManager
}

type KYCProvider interface {
	VerifyIdentity(ctx context.Context, customer Customer) (*KYCResult, error)
	UpdateVerification(ctx context.Context, customerID string) error
	GetVerificationStatus(customerID string) (KYCStatus, error)
}

type AMLMonitor interface {
	MonitorTransaction(ctx context.Context, tx Transaction) (*AMLAlert, error)
	GenerateSAR(ctx context.Context, alert AMLAlert) (*SARReport, error)
	CheckSanctions(ctx context.Context, entity Entity) (bool, error)
}

type ComplianceCheck struct {
	CustomerID      string             `json:"customer_id"`
	TransactionID   string             `json:"transaction_id"`
	KYCStatus       KYCStatus          `json:"kyc_status"`
	AMLRisk         RiskLevel          `json:"aml_risk"`
	SanctionsHit    bool               `json:"sanctions_hit"`
	RiskScore       float64            `json:"risk_score"`
	RequiredActions []ComplianceAction `json:"required_actions"`
	Timestamp       time.Time          `json:"timestamp"`
}

// PerformComplianceCheck выполняет полную проверку соответствия
func (ce *ComplianceEngine) PerformComplianceCheck(
	ctx context.Context,
	customer Customer,
	transaction Transaction,
) (*ComplianceCheck, error) {

	check := &ComplianceCheck{
		CustomerID:    customer.ID,
		TransactionID: transaction.ID,
		Timestamp:     time.Now(),
	}

	// KYC верификация
	kycResult, err := ce.kycProvider.VerifyIdentity(ctx, customer)
	if err != nil {
		return nil, fmt.Errorf("KYC verification failed: %w", err)
	}
	check.KYCStatus = kycResult.Status

	// Проверка санкций
	sanctionsHit, err := ce.amlMonitor.CheckSanctions(ctx, Entity{
		Name:    customer.Name,
		Address: customer.Address,
		Country: customer.Country,
	})
	if err != nil {
		return nil, fmt.Errorf("sanctions check failed: %w", err)
	}
	check.SanctionsHit = sanctionsHit

	// AML мониторинг транзакции
	amlAlert, err := ce.amlMonitor.MonitorTransaction(ctx, transaction)
	if err != nil {
		return nil, fmt.Errorf("AML monitoring failed: %w", err)
	}

	if amlAlert != nil {
		check.AMLRisk = amlAlert.RiskLevel

		// Генерация SAR если необходимо
		if amlAlert.RiskLevel >= RiskLevelHigh {
			sar, err := ce.amlMonitor.GenerateSAR(ctx, *amlAlert)
			if err != nil {
				return nil, fmt.Errorf("SAR generation failed: %w", err)
			}

			check.RequiredActions = append(check.RequiredActions,
				ComplianceAction{
					Type:    "FILE_SAR",
					Details: sar.ID,
				})
		}
	}

	// Расчет общего риск-скора
	check.RiskScore = ce.riskScorer.CalculateRiskScore(RiskFactors{
		KYCStatus:         check.KYCStatus,
		AMLRisk:           check.AMLRisk,
		SanctionsHit:      check.SanctionsHit,
		TransactionAmount: transaction.Amount,
		CustomerHistory:   customer.TransactionHistory,
	})

	return check, nil
}

type RiskLevel int

const (
	RiskLevelLow RiskLevel = iota
	RiskLevelMedium
	RiskLevelHigh
	RiskLevelCritical
)

type KYCStatus int

const (
	KYCStatusPending KYCStatus = iota
	KYCStatusVerified
	KYCStatusRejected
	KYCStatusExpired
)
