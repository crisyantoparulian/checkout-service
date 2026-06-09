package checkout

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	"github.com/crisyantoparulian/checkout-service/internal/pkg/apperror"
	"github.com/crisyantoparulian/checkout-service/internal/pkg/constants"
	"github.com/google/uuid"
)

func (u *usecase) GetDetail(ctx context.Context, checkoutID string) (resp CheckoutResponse, err error) {
	checkoutUUID, err := uuid.Parse(checkoutID)
	if err != nil {
		return
	}

	checkout, err := u.checkoutRepository.GetByUUID(ctx, checkoutUUID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = apperror.New(http.StatusNotFound, constants.CODE_CHECKOUT_NOT_FOUND, errors.New("checkout not found"))
		}
		return
	}

	items, err := u.checkoutRepository.GetItems(ctx, checkoutUUID)
	if err != nil {
		return
	}

	promotions, err := u.checkoutRepository.GetPromotions(ctx, checkoutUUID)
	if err != nil {
		return
	}

	appliedPromotions := []appliedPromotion{}
	for _, promotion := range promotions {
		appliedPromotions = append(appliedPromotions, appliedPromotion{
			PromotionCode: promotion.PromotionCode,
			PromotionName: promotion.PromotionName,
			PromotionType: promotion.PromotionType,
			DiscountCents: promotion.DiscountCents,
		})
	}

	resp = mapCheckoutResponse(checkout, items, appliedPromotions)
	resp.CreatedAt = checkout.CreatedAt.Format(timeFormat)
	return
}
