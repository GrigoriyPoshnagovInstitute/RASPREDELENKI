package transport

import (
	"context"

	"github.com/google/uuid"

	"paymentservice/api/server/paymentinternal"
	appmodel "paymentservice/pkg/payment/application/model"
	"paymentservice/pkg/payment/application/query"
	"paymentservice/pkg/payment/application/service"
)

func NewPaymentInternalAPI(
	accountQueryService query.AccountQueryService,
	accountService service.AccountService,
) paymentinternal.PaymentInternalServiceServer {
	return &paymentInternalAPI{
		accountQueryService: accountQueryService,
		accountService:      accountService,
	}
}

type paymentInternalAPI struct {
	accountQueryService query.AccountQueryService
	accountService      service.AccountService

	paymentinternal.UnimplementedPaymentInternalServiceServer
}

func (p *paymentInternalAPI) StoreUserBalance(ctx context.Context, request *paymentinternal.StoreUserBalanceRequest) (*paymentinternal.StoreUserBalanceResponse, error) {
	userID, err := uuid.Parse(request.Balance.UserID)
	if err != nil {
		return nil, err
	}

	err = p.accountService.StoreUserBalance(ctx, appmodel.UserBalance{
		UserID:  userID,
		Balance: request.Balance.Balance,
	})
	if err != nil {
		return nil, err
	}

	return &paymentinternal.StoreUserBalanceResponse{
		UserID: userID.String(),
	}, nil
}

func (p *paymentInternalAPI) FindUserBalance(ctx context.Context, request *paymentinternal.FindUserBalanceRequest) (*paymentinternal.FindUserBalanceResponse, error) {
	userID, err := uuid.Parse(request.UserID)
	if err != nil {
		return nil, err
	}
	balance, err := p.accountQueryService.FindUserBalance(ctx, userID)
	if err != nil {
		return nil, err
	}
	if balance == nil {
		return &paymentinternal.FindUserBalanceResponse{}, nil
	}
	return &paymentinternal.FindUserBalanceResponse{
		Balance: &paymentinternal.UserBalance{
			UserID:  balance.UserID.String(),
			Balance: balance.Balance,
		},
	}, nil
}
