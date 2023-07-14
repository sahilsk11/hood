// Code generated by MockGen. DO NOT EDIT.
// Source: trade_ingestion_service.go

// Package mock_trade_ingestion is a generated GoMock package.
package trade_ingestion

import (
	context "context"
	sql "database/sql"
	model "hood/internal/db/models/postgres/public/model"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockTradeIngestionService is a mock of TradeIngestionService interface.
type MockTradeIngestionService struct {
	ctrl     *gomock.Controller
	recorder *MockTradeIngestionServiceMockRecorder
}

// MockTradeIngestionServiceMockRecorder is the mock recorder for MockTradeIngestionService.
type MockTradeIngestionServiceMockRecorder struct {
	mock *MockTradeIngestionService
}

// NewMockTradeIngestionService creates a new mock instance.
func NewMockTradeIngestionService(ctrl *gomock.Controller) *MockTradeIngestionService {
	mock := &MockTradeIngestionService{ctrl: ctrl}
	mock.recorder = &MockTradeIngestionServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockTradeIngestionService) EXPECT() *MockTradeIngestionServiceMockRecorder {
	return m.recorder
}

// AddAssetSplit mocks base method.
func (m *MockTradeIngestionService) AddAssetSplit(ctx context.Context, tx *sql.Tx, split model.AssetSplit) (*model.AssetSplit, []model.AppliedAssetSplit, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddAssetSplit", ctx, tx, split)
	ret0, _ := ret[0].(*model.AssetSplit)
	ret1, _ := ret[1].([]model.AppliedAssetSplit)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// AddAssetSplit indicates an expected call of AddAssetSplit.
func (mr *MockTradeIngestionServiceMockRecorder) AddAssetSplit(ctx, tx, split interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddAssetSplit", reflect.TypeOf((*MockTradeIngestionService)(nil).AddAssetSplit), ctx, tx, split)
}

// ProcessBuyOrder mocks base method.
func (m *MockTradeIngestionService) ProcessBuyOrder(ctx context.Context, tx *sql.Tx, newTrade model.Trade) (*model.Trade, *model.OpenLot, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ProcessBuyOrder", ctx, tx, newTrade)
	ret0, _ := ret[0].(*model.Trade)
	ret1, _ := ret[1].(*model.OpenLot)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// ProcessBuyOrder indicates an expected call of ProcessBuyOrder.
func (mr *MockTradeIngestionServiceMockRecorder) ProcessBuyOrder(ctx, tx, newTrade interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ProcessBuyOrder", reflect.TypeOf((*MockTradeIngestionService)(nil).ProcessBuyOrder), ctx, tx, newTrade)
}

// ProcessSellOrder mocks base method.
func (m *MockTradeIngestionService) ProcessSellOrder(ctx context.Context, tx *sql.Tx, newTrade model.Trade) (*model.Trade, []*model.ClosedLot, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ProcessSellOrder", ctx, tx, newTrade)
	ret0, _ := ret[0].(*model.Trade)
	ret1, _ := ret[1].([]*model.ClosedLot)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// ProcessSellOrder indicates an expected call of ProcessSellOrder.
func (mr *MockTradeIngestionServiceMockRecorder) ProcessSellOrder(ctx, tx, newTrade interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ProcessSellOrder", reflect.TypeOf((*MockTradeIngestionService)(nil).ProcessSellOrder), ctx, tx, newTrade)
}

// ProcessTdaBuyOrder mocks base method.
func (m *MockTradeIngestionService) ProcessTdaBuyOrder(ctx context.Context, tx *sql.Tx, newTrade model.Trade, tdaTxId int64) (*model.Trade, *model.OpenLot, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ProcessTdaBuyOrder", ctx, tx, newTrade, tdaTxId)
	ret0, _ := ret[0].(*model.Trade)
	ret1, _ := ret[1].(*model.OpenLot)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// ProcessTdaBuyOrder indicates an expected call of ProcessTdaBuyOrder.
func (mr *MockTradeIngestionServiceMockRecorder) ProcessTdaBuyOrder(ctx, tx, newTrade, tdaTxId interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ProcessTdaBuyOrder", reflect.TypeOf((*MockTradeIngestionService)(nil).ProcessTdaBuyOrder), ctx, tx, newTrade, tdaTxId)
}
