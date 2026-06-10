package checkout

import (
	"context"

	"github.com/crisyantoparulian/checkout-service/internal/app/repository"
	"github.com/crisyantoparulian/checkout-service/internal/pkg/transaction"
)

const (
	statusCompleted            = "COMPLETED"
	movementTypeCheckoutDeduct = "CHECKOUT_DEDUCT"
	timeFormat                 = "2006-01-02T15:04:05Z0700"
)

type CheckoutUsecase interface {
	Create(ctx context.Context, req CreateCheckoutRequest) (CheckoutResponse, error)
	GetDetail(ctx context.Context, checkoutID string) (CheckoutResponse, error)
}

type usecase struct {
	txManager           transaction.ManagerInterface
	checkoutRepository  repository.Checkout
	productRepository   repository.Product
	promotionRepository repository.Promotion
	inventoryRepository repository.Inventory
}

func NewUsecase() *usecase {
	return &usecase{}
}

func (u *usecase) SetTxManager(tm transaction.ManagerInterface) *usecase {
	u.txManager = tm
	return u
}

func (u *usecase) SetCheckoutRepository(repo repository.Checkout) *usecase {
	u.checkoutRepository = repo
	return u
}

func (u *usecase) SetProductRepository(repo repository.Product) *usecase {
	u.productRepository = repo
	return u
}

func (u *usecase) SetPromotionRepository(repo repository.Promotion) *usecase {
	u.promotionRepository = repo
	return u
}

func (u *usecase) SetInventoryRepository(repo repository.Inventory) *usecase {
	u.inventoryRepository = repo
	return u
}

func (u *usecase) Validate() CheckoutUsecase {
	if u.txManager == nil {
		panic("txManager is nil")
	}
	if u.checkoutRepository == nil {
		panic("checkoutRepository is nil")
	}
	if u.productRepository == nil {
		panic("productRepository is nil")
	}
	if u.promotionRepository == nil {
		panic("promotionRepository is nil")
	}
	if u.inventoryRepository == nil {
		panic("inventoryRepository is nil")
	}
	return u
}
