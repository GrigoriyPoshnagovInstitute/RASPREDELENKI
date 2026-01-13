package activity

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"productservice/pkg/product/application/query"
)

func NewProductActivities(productQueryService query.ProductQueryService) *ProductActivities {
	return &ProductActivities{productQueryService: productQueryService}
}

type ProductActivities struct {
	productQueryService query.ProductQueryService
}

type OrderItem struct {
	ProductID string
	Quantity  int
}

func (a *ProductActivities) ReserveProducts(ctx context.Context, items []OrderItem) (bool, error) {
	fmt.Printf("Checking availability for %d items\n", len(items))

	for _, item := range items {
		pID, err := uuid.Parse(item.ProductID)
		if err != nil {
			return false, fmt.Errorf("invalid product id: %s", item.ProductID)
		}

		product, err := a.productQueryService.FindProduct(ctx, pID)
		if err != nil {
			return false, err
		}
		if product == nil {
			return false, fmt.Errorf("product not found: %s", item.ProductID)
		}

		fmt.Printf("Product confirmed: %s\n", product.Name)
	}

	return true, nil
}

func (a *ProductActivities) ReleaseProducts(_ context.Context, items []OrderItem) (bool, error) {
	// In a real system with 'stock' column, this would increment stock back.
	// Since we don't have stock, we just log.
	fmt.Printf("Releasing products reservation: %+v\n", items)
	return true, nil
}
