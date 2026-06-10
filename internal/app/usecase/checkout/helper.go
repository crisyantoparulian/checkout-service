package checkout

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/crisyantoparulian/checkout-service/internal/app/entity"
	"github.com/crisyantoparulian/checkout-service/internal/pkg/apperror"
	"github.com/crisyantoparulian/checkout-service/internal/pkg/constants"
)

func (u *usecase) applyPromotions(quantityByUUID map[string]int, productByUUID map[string]entity.Product, rules []entity.PromotionRule) ([]appliedPromotion, map[string]int, int64, error) {
	applied := []appliedPromotion{}
	rewardQuantityByUUID := map[string]int{}
	var totalDiscount int64

	for _, rule := range rules {
		qty := quantityByUUID[rule.TargetProductUUID.String()]
		if qty == 0 {
			continue
		}

		var discount int64
		switch rule.PromotionType {
		case constants.PromotionTypeFreeProduct:
			if !rule.MinQuantity.Valid || !rule.FreeQuantity.Valid || !rule.RewardProductUUID.Valid {
				return nil, nil, 0, apperror.New(http.StatusInternalServerError, constants.CODE_INVALID_PROMOTION_RULE, errors.New("invalid free product rule"))
			}
			if qty >= int(rule.MinQuantity.Int64) {
				rewardUUID := rule.RewardProductUUID.UUID.String()
				rewardProduct := productByUUID[rewardUUID]
				freeQty := qty / int(rule.MinQuantity.Int64) * int(rule.FreeQuantity.Int64)
				requestedRewardQty := quantityByUUID[rewardUUID]
				if requestedRewardQty < freeQty {
					autoRewardQty := freeQty - requestedRewardQty
					availableRewardQty := rewardProduct.InventoryQty - requestedRewardQty
					if availableRewardQty < autoRewardQty {
						autoRewardQty = availableRewardQty
					}
					if autoRewardQty > 0 {
						rewardQuantityByUUID[rewardUUID] += autoRewardQty
						requestedRewardQty += autoRewardQty
					}
				}
				if requestedRewardQty < freeQty {
					freeQty = requestedRewardQty
				}
				if freeQty > 0 {
					discount = rewardProduct.PriceCents * int64(freeQty)
				}
			}
		case constants.PromotionTypeBuyXPayY:
			if !rule.BuyQuantity.Valid || !rule.PayQuantity.Valid || rule.BuyQuantity.Int64 <= rule.PayQuantity.Int64 {
				return nil, nil, 0, apperror.New(http.StatusInternalServerError, constants.CODE_INVALID_PROMOTION_RULE, errors.New("invalid buy x pay y rule"))
			}
			product := productByUUID[rule.TargetProductUUID.String()]
			groupCount := qty / int(rule.BuyQuantity.Int64)
			freePerGroup := int(rule.BuyQuantity.Int64 - rule.PayQuantity.Int64)
			discount = product.PriceCents * int64(groupCount*freePerGroup)
		case constants.PromotionTypePercentageDiscount:
			if !rule.MinQuantity.Valid || !rule.DiscountPercentage.Valid {
				return nil, nil, 0, apperror.New(http.StatusInternalServerError, constants.CODE_INVALID_PROMOTION_RULE, errors.New("invalid percentage rule"))
			}
			if qty >= int(rule.MinQuantity.Int64) {
				product := productByUUID[rule.TargetProductUUID.String()]
				discount = int64(float64(product.PriceCents*int64(qty)) * rule.DiscountPercentage.Float64 / 100)
			}
		}

		if discount <= 0 {
			continue
		}
		snapshot, err := json.Marshal(rule)
		if err != nil {
			return nil, nil, 0, err
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

	return applied, rewardQuantityByUUID, totalDiscount, nil
}
