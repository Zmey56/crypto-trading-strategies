package crosschain

import (
	"context"
	"crypto-trading-strategies/pkg/types"
	"fmt"
	"sync"
	"time"
)

type CrossChainArbitrageEngine struct {
	bridges      map[string]Bridge
	dexes        map[string]DEXClient
	flashLoaners map[string]FlashLoanProvider
	gasTracker   *GasTracker

	// Concurrent execution
	executor *CrossChainExecutor
	mutex    sync.RWMutex
}

type Bridge interface {
	Transfer(ctx context.Context, token string, amount float64,
		fromChain, toChain string) (*TransferReceipt, error)
	EstimateTime(fromChain, toChain string) time.Duration
	EstimateFee(token string, amount float64, fromChain, toChain string) (float64, error)
}

type ArbitrageOpportunity struct {
	TokenSymbol     string             `json:"token_symbol"`
	BuyChain        string             `json:"buy_chain"`
	SellChain       string             `json:"sell_chain"`
	BuyPrice        float64            `json:"buy_price"`
	SellPrice       float64            `json:"sell_price"`
	ProfitMargin    float64            `json:"profit_margin"`
	RequiredCapital float64            `json:"required_capital"`
	EstimatedProfit float64            `json:"estimated_profit"`
	Risks           []string           `json:"risks"`
	ExecutionTime   time.Duration      `json:"execution_time"`
	GasFees         map[string]float64 `json:"gas_fees"`
}

// ScanArbitrageOpportunities ищет возможности арбитража между сетями
func (ace *CrossChainArbitrageEngine) ScanArbitrageOpportunities(
	ctx context.Context,
	tokens []string,
) ([]ArbitrageOpportunity, error) {

	var opportunities []ArbitrageOpportunity
	var wg sync.WaitGroup
	opsChan := make(chan ArbitrageOpportunity, 100)

	// Параллельное сканирование всех пар цепочек
	for _, token := range tokens {
		for buyChain := range ace.dexes {
			for sellChain := range ace.dexes {
				if buyChain == sellChain {
					continue
				}

				wg.Add(1)
				go func(token, buy, sell string) {
					defer wg.Done()

					opp := ace.analyzeOpportunity(ctx, token, buy, sell)
					if opp.ProfitMargin > ace.getMinProfitThreshold() {
						opsChan <- opp
					}
				}(token, buyChain, sellChain)
			}
		}
	}

	// Закрытие канала после завершения всех горутин
	go func() {
		wg.Wait()
		close(opsChan)
	}()

	// Сбор результатов
	for opp := range opsChan {
		opportunities = append(opportunities, opp)
	}

	return ace.filterAndRankOpportunities(opportunities), nil
}

// ExecuteArbitrage выполняет cross-chain арбитраж с использованием flash loans
func (ace *CrossChainArbitrageEngine) ExecuteArbitrage(
	ctx context.Context,
	opportunity ArbitrageOpportunity,
) (*ArbitrageResult, error) {

	// Получение flash loan для начального капитала
	flashLoan, err := ace.flashLoaners[opportunity.BuyChain].RequestLoan(
		ctx,
		opportunity.TokenSymbol,
		opportunity.RequiredCapital,
	)
	if err != nil {
		return nil, fmt.Errorf("flash loan failed: %w", err)
	}

	// Выполнение арбитража в рамках одной транзакции
	result := &ArbitrageResult{
		OpportunityID: opportunity.ID,
		StartTime:     time.Now(),
	}

	// Шаг 1: Покупка токена на исходной сети
	buyTx, err := ace.dexes[opportunity.BuyChain].BuyToken(
		ctx,
		opportunity.TokenSymbol,
		opportunity.RequiredCapital,
	)
	if err != nil {
		return result, fmt.Errorf("buy failed: %w", err)
	}
	result.BuyTransaction = buyTx

	// Шаг 2: Bridge токенов в целевую сеть
	bridgeTx, err := ace.bridges[opportunity.BuyChain].Transfer(
		ctx,
		opportunity.TokenSymbol,
		buyTx.TokenAmount,
		opportunity.BuyChain,
		opportunity.SellChain,
	)
	if err != nil {
		return result, fmt.Errorf("bridge failed: %w", err)
	}
	result.BridgeTransaction = bridgeTx

	// Шаг 3: Продажа токена на целевой сети
	sellTx, err := ace.dexes[opportunity.SellChain].SellToken(
		ctx,
		opportunity.TokenSymbol,
		buyTx.TokenAmount,
	)
	if err != nil {
		return result, fmt.Errorf("sell failed: %w", err)
	}
	result.SellTransaction = sellTx

	// Шаг 4: Возврат flash loan
	repayment := flashLoan.Principal + flashLoan.Fee
	if sellTx.ReceivedAmount < repayment {
		return result, fmt.Errorf("insufficient funds to repay flash loan")
	}

	err = ace.flashLoaners[opportunity.BuyChain].RepayLoan(ctx, flashLoan)
	if err != nil {
		return result, fmt.Errorf("loan repayment failed: %w", err)
	}

	result.NetProfit = sellTx.ReceivedAmount - repayment
	result.EndTime = time.Now()
	result.Success = true

	return result, nil
}
