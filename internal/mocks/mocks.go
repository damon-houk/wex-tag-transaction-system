// internal/mocks/mocks.go
package mocks

import (
	"context"
	"time"

	"github.com/damon-houk/wex-tag-transaction-system/internal/domain/entity"
	"github.com/stretchr/testify/mock"
)

// MockTransactionRepository mocks the TransactionRepository interface
type MockTransactionRepository struct {
	mock.Mock
}

func (m *MockTransactionRepository) Store(ctx context.Context, tx *entity.Transaction) (string, error) {
	args := m.Called(ctx, tx)
	return args.String(0), args.Error(1)
}

func (m *MockTransactionRepository) FindByID(ctx context.Context, id string) (*entity.Transaction, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Transaction), args.Error(1)
}

// MockExchangeRateRepository mocks the ExchangeRateRepository interface
type MockExchangeRateRepository struct {
	mock.Mock
}

func (m *MockExchangeRateRepository) FindRate(ctx context.Context, currency string, date time.Time) (*entity.ExchangeRate, error) {
	args := m.Called(ctx, currency, date)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.ExchangeRate), args.Error(1)
}

func (m *MockExchangeRateRepository) StoreRate(ctx context.Context, rate *entity.ExchangeRate) error {
	args := m.Called(ctx, rate)
	return args.Error(0)
}

// MockExchangeRateProvider mocks the exchange rate provider interface
type MockExchangeRateProvider struct {
	mock.Mock
}

func (m *MockExchangeRateProvider) FetchExchangeRate(ctx context.Context, currency string, date time.Time) (*entity.ExchangeRate, error) {
	args := m.Called(ctx, currency, date)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.ExchangeRate), args.Error(1)
}

// MockLogger mocks the logger interface
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Debug(msg string, fields map[string]interface{}) {
	m.Called(msg, fields)
}

func (m *MockLogger) Info(msg string, fields map[string]interface{}) {
	m.Called(msg, fields)
}

func (m *MockLogger) Warn(msg string, fields map[string]interface{}) {
	m.Called(msg, fields)
}

func (m *MockLogger) Error(msg string, fields map[string]interface{}) {
	m.Called(msg, fields)
}

func (m *MockLogger) Fatal(msg string, fields map[string]interface{}) {
	m.Called(msg, fields)
}

func (m *MockLogger) WithField(key string, value interface{}) interface{} {
	args := m.Called(key, value)
	return args.Get(0)
}

func (m *MockLogger) WithFields(fields map[string]interface{}) interface{} {
	args := m.Called(fields)
	return args.Get(0)
}
