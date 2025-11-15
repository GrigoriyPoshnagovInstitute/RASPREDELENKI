package transport

import (
	"context"

	"github.com/google/uuid"

	"productservice/api/server/productinternal"
	appmodel "productservice/pkg/product/application/model"
	"productservice/pkg/product/application/query"
	"productservice/pkg/product/application/service"
)

func NewProductInternalAPI(
	productQueryService query.ProductQueryService,
	productService service.ProductService,
) productinternal.ProductInternalServiceServer {
	return &productInternalAPI{
		productQueryService: productQueryService,
		productService:      productService,
	}
}

type productInternalAPI struct {
	productQueryService query.ProductQueryService
	productService      service.ProductService

	productinternal.UnimplementedProductInternalServiceServer
}

func (p *productInternalAPI) StoreProduct(ctx context.Context, request *productinternal.StoreProductRequest) (*productinternal.StoreProductResponse, error) {
	var (
		productID uuid.UUID
		err       error
	)
	if request.Product.ProductID != "" {
		productID, err = uuid.Parse(request.Product.ProductID)
		if err != nil {
			return nil, err
		}
	}

	productID, err = p.productService.StoreProduct(ctx, appmodel.Product{
		ProductID:   productID,
		Name:        request.Product.Name,
		Price:       request.Product.Price,
		Description: request.Product.Description,
	})
	if err != nil {
		return nil, err
	}

	return &productinternal.StoreProductResponse{
		ProductID: productID.String(),
	}, nil
}

func (p *productInternalAPI) FindProduct(ctx context.Context, request *productinternal.FindProductRequest) (*productinternal.FindProductResponse, error) {
	productID, err := uuid.Parse(request.ProductID)
	if err != nil {
		return nil, err
	}
	product, err := p.productQueryService.FindProduct(ctx, productID)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return &productinternal.FindProductResponse{}, nil
	}
	return &productinternal.FindProductResponse{
		Product: &productinternal.Product{
			ProductID:   product.ProductID.String(),
			Name:        product.Name,
			Price:       product.Price,
			Description: product.Description,
		},
	}, nil
}
