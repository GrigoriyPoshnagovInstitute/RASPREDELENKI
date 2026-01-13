package activity

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"paymentservice/pkg/payment/application/service"
)

func NewPaymentActivities(accountService service.AccountService) *PaymentActivities {
	return &PaymentActivities{accountService: accountService}
}

type PaymentActivities struct {
	accountService service.AccountService
}

func (a *PaymentActivities) ProcessPayment(ctx context.Context, userIDStr string, amount int64) (bool, error) {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return false, err
	}

	fmt.Printf("Attempting to charge user %s amount %d\n", userIDStr, amount)
	err = a.accountService.Charge(ctx, userID, amount)
	if err != nil {
		fmt.Printf("Charge failed: %v\n", err)
		return false, err
	}

	fmt.Printf("Charge success for user %s\n", userIDStr)
	return true, nil
}

func (a *PaymentActivities) RefundPayment(ctx context.Context, userIDStr string, amount int64) (bool, error) {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return false, err
	}

	fmt.Printf("Refund user %s amount %d\n", userIDStr, amount)
	err = a.accountService.Refund(ctx, userID, amount)
	if err != nil {
		return false, err
	}
	return true, nil
}
