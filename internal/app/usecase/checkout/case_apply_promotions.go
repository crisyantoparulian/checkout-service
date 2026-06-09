package checkout

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/crisyantoparulian/checkout-service/internal/app/entity"
	"github.com/crisyantoparulian/checkout-service/internal/pkg/apperror"
	"github.com/crisyantoparulian/checkout-service/internal/pkg/constants"
)

func (u *usecase) applyPromotions(quantityBySKU map[string]int, productBySKU map[string]entity.Product, rules []entity.PromotionRule) ([]appliedPromotion, int64, error) {
	applied := []appliedPromotion{}
	var totalDiscount int64

	for _, rule := range rules {
		qty := quantityBySKU[rule.TargetSKU]
		if qty == 0 {
			continue
		}

		var discount int64
		switch rule.PromotionType {
		case promotionTypeFreeProduct:
			if !rule.MinQuantity.Valid || !rule.FreeQuantity.Valid || !rule.RewardSKU.Valid {
				return nil, 0, apperror.New(http.StatusInternalServerError, constants.CODE_INVALID_PROMOTION_RULE, errors.New("invalid free product rule"))
			}
			if qty >= int(rule.MinQuantity.Int64) && quantityBySKU[rule.RewardSKU.String] > 0 {
				rewardProduct := productBySKU[rule.RewardSKU.String]
				freeQty := qty / int(rule.MinQuantity.Int64) * int(rule.FreeQuantity.Int64)
				if quantityBySKU[rule.RewardSKU.String] < freeQty {
					freeQty = quantityBySKU[rule.RewardSKU.String]
				}
				discount = rewardProduct.PriceCents * int64(freeQty)
			}
		case promotionTypeBuyXPayY:
			if !rule.BuyQuantity.Valid || !rule.PayQuantity.Valid || rule.BuyQuantity.Int64 <= rule.PayQuantity.Int64 {
				return nil, 0, apperror.New(http.StatusInternalServerError, constants.CODE_INVALID_PROMOTION_RULE, errors.New("invalid buy x pay y rule"))
			}
			product := productBySKU[rule.TargetSKU]
			groupCount := qty / int(rule.BuyQuantity.Int64)
			freePerGroup := int(rule.BuyQuantity.Int64 - rule.PayQuantity.Int64)
			discount = product.PriceCents * int64(groupCount*freePerGroup)
		case promotionTypePercentage:
			if !rule.MinQuantity.Valid || !rule.DiscountPercentage.Valid {
				return nil, 0, apperror.New(http.StatusInternalServerError, constants.CODE_INVALID_PROMOTION_RULE, errors.New("invalid percentage rule"))
			}
			if qty >= int(rule.MinQuantity.Int64) {
				product := productBySKU[rule.TargetSKU]
				discount = int64(float64(product.PriceCents*int64(qty)) * rule.DiscountPercentage.Float64 / 100)
			}
		}

		if discount <= 0 {
			continue
		}
		snapshot, err := json.Marshal(rule)
		if err != nil {
			return nil, 0, err
		}
		applied = append(applied, appliedPromotion{
			PromotionRuleUUID: rule.UUID.String(),
			PromotionUUID:     rule.PromotionUUID.String(),
			PromotionCode:     rule.PromotionCode,
			PromotionName:     rule.PromotionName,
			PromotionType:     rule.PromotionType,
			DiscountCents:     discount,
			RuleSnapshot:      snapshot,
		})
		totalDiscount += discount
	}

	return applied, totalDiscount, nil
}
