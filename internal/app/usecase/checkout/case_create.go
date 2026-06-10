package checkout

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/crisyantoparulian/checkout-service/internal/app/entity"
	"github.com/crisyantoparulian/checkout-service/internal/pkg/apperror"
	"github.com/crisyantoparulian/checkout-service/internal/pkg/constants"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

func (u *usecase) Create(ctx context.Context, req CreateCheckoutRequest) (resp CheckoutResponse, err error) {
	quantityByUUID, productUUIDs, err := buildCheckoutQuantities(req)
	if err != nil {
		return
	}

	// Fetch promotion rules by target product UUIDs
	rules, err := u.promotionRepository.GetActiveRulesByTargetProductUUIDs(ctx, productUUIDs)
	if err != nil {
		return
	}

	// Collect all UUIDs to lock: cart products + reward products
	allUUIDs := appendLockedUUIDs(productUUIDs, rules)

	var (
		checkout          entity.Checkout
		checkoutItems     []entity.CheckoutItem
		appliedPromotions []appliedPromotion
	)

	err = u.txManager.WithTx(ctx, func(tx *sqlx.Tx) error {
		// Single lock query for all products (cart + reward)
		lockedProducts, err := u.productRepository.GetByUUIDsForUpdate(ctx, tx, allUUIDs)
		if err != nil {
			return err
		}

		// Build map and validate stock with locked data
		productByUUID, err := validateAndRefreshProducts(quantityByUUID, lockedProducts, nil)
		if err != nil {
			return err
		}

		var discount int64
		var rewardQuantityByUUID map[string]int
		appliedPromotions, rewardQuantityByUUID, discount, err = u.applyPromotions(quantityByUUID, productByUUID, rules)
		if err != nil {
			return err
		}

		var subtotal int64
		checkoutItems, subtotal = buildCheckoutItems(req, productByUUID)
		checkoutItems, subtotal = appendRewardCheckoutItems(checkoutItems, subtotal, rewardQuantityByUUID, productByUUID)
		deductQuantityByUUID := mergeQuantities(quantityByUUID, rewardQuantityByUUID)

		checkout = entity.Checkout{
			Status:        statusCompleted,
			SubtotalCents: subtotal,
			DiscountCents: discount,
			TotalCents:    subtotal - discount,
		}
		if err := u.checkoutRepository.Create(ctx, tx, &checkout); err != nil {
			return err
		}

		if err := u.createCheckoutItems(ctx, tx, checkout.UUID, checkoutItems); err != nil {
			return err
		}

		if err := u.createCheckoutPromotions(ctx, tx, checkout.UUID, appliedPromotions); err != nil {
			return err
		}

		if err := u.deductInventory(ctx, tx, checkout.UUID, deductQuantityByUUID, productByUUID); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return
	}

	resp = mapCheckoutResponse(checkout, checkoutItems, appliedPromotions)
	return
}

func buildCheckoutQuantities(req CreateCheckoutRequest) (map[string]int, []uuid.UUID, error) {
	quantityByUUID := map[string]int{}
	var productUUIDs []uuid.UUID
	for _, item := range req.Items {
		if item.Quantity <= 0 {
			return nil, nil, apperror.New(http.StatusBadRequest, constants.CODE_INVALID_QUANTITY, errors.New("quantity must be greater than zero"))
		}
		if quantityByUUID[item.ProductUUID] == 0 {
			productUUID, err := uuid.Parse(item.ProductUUID)
			if err != nil {
				return nil, nil, apperror.New(http.StatusBadRequest, constants.CODE_PRODUCT_NOT_FOUND, fmt.Errorf("invalid product_uuid: %s", item.ProductUUID))
			}
			productUUIDs = append(productUUIDs, productUUID)
		}
		quantityByUUID[item.ProductUUID] += item.Quantity
	}
	return quantityByUUID, productUUIDs, nil
}

// validateAndRefreshProducts updates (or creates) a product map from the given products
// and validates that requested quantities are available in stock.
// If existingMap is nil, a new map is created. If provided, it is updated in-place.
func validateAndRefreshProducts(quantityByUUID map[string]int, products []entity.Product, existingMap map[string]entity.Product) (map[string]entity.Product, error) {
	productByUUID := existingMap
	if productByUUID == nil {
		productByUUID = map[string]entity.Product{}
	}

	for _, product := range products {
		productByUUID[product.UUID.String()] = product
	}

	for uuidStr, qty := range quantityByUUID {
		product, ok := productByUUID[uuidStr]
		if !ok {
			return nil, apperror.New(http.StatusNotFound, constants.CODE_PRODUCT_NOT_FOUND, fmt.Errorf("product %s not found", uuidStr))
		}
		if product.InventoryQty < qty {
			return nil, apperror.New(http.StatusConflict, constants.CODE_INSUFFICIENT_STOCK, fmt.Errorf("product %s only has %d items available", product.Name, product.InventoryQty))
		}
	}
	return productByUUID, nil
}

func buildCheckoutItems(req CreateCheckoutRequest, productByUUID map[string]entity.Product) ([]entity.CheckoutItem, int64) {
	checkoutItems := []entity.CheckoutItem{}
	var subtotal int64
	for _, item := range req.Items {
		product := productByUUID[item.ProductUUID]
		totalPrice := product.PriceCents * int64(item.Quantity)
		subtotal += totalPrice
		checkoutItems = append(checkoutItems, entity.CheckoutItem{
			ProductUUID:     product.UUID,
			SKU:             product.SKU,
			ProductName:     product.Name,
			Quantity:        item.Quantity,
			UnitPriceCents:  product.PriceCents,
			TotalPriceCents: totalPrice,
		})
	}
	return checkoutItems, subtotal
}

func appendRewardCheckoutItems(checkoutItems []entity.CheckoutItem, subtotal int64, rewardQuantityByUUID map[string]int, productByUUID map[string]entity.Product) ([]entity.CheckoutItem, int64) {
	for uuidStr, qty := range rewardQuantityByUUID {
		if qty <= 0 {
			continue
		}
		product := productByUUID[uuidStr]
		totalPrice := product.PriceCents * int64(qty)
		subtotal += totalPrice
		checkoutItems = append(checkoutItems, entity.CheckoutItem{
			ProductUUID:     product.UUID,
			SKU:             product.SKU,
			ProductName:     product.Name,
			Quantity:        qty,
			UnitPriceCents:  product.PriceCents,
			TotalPriceCents: totalPrice,
		})
	}
	return checkoutItems, subtotal
}

func mergeQuantities(base map[string]int, extra map[string]int) map[string]int {
	merged := map[string]int{}
	for uuidStr, qty := range base {
		merged[uuidStr] = qty
	}
	for uuidStr, qty := range extra {
		merged[uuidStr] += qty
	}
	return merged
}

func (u *usecase) createCheckoutItems(ctx context.Context, tx *sqlx.Tx, checkoutUUID uuid.UUID, checkoutItems []entity.CheckoutItem) error {
	for i := range checkoutItems {
		checkoutItems[i].CheckoutUUID = checkoutUUID
		if err := u.checkoutRepository.CreateItem(ctx, tx, &checkoutItems[i]); err != nil {
			return err
		}
	}
	return nil
}

func (u *usecase) createCheckoutPromotions(ctx context.Context, tx *sqlx.Tx, checkoutUUID uuid.UUID, promotions []appliedPromotion) error {
	for _, promotion := range promotions {
		promotionRuleUUID, _ := uuid.Parse(promotion.PromotionRuleUUID)
		promotionUUID, _ := uuid.Parse(promotion.PromotionUUID)
		checkoutPromotion := entity.CheckoutPromotion{
			CheckoutUUID:      checkoutUUID,
			PromotionUUID:     uuid.NullUUID{UUID: promotionUUID, Valid: true},
			PromotionRuleUUID: uuid.NullUUID{UUID: promotionRuleUUID, Valid: true},
			PromotionCode:     promotion.PromotionCode,
			PromotionName:     promotion.PromotionName,
			PromotionType:     promotion.PromotionType,
			DiscountCents:     promotion.DiscountCents,
			RuleSnapshot:      promotion.RuleSnapshot,
		}
		if err := u.checkoutRepository.CreatePromotion(ctx, tx, &checkoutPromotion); err != nil {
			return err
		}
	}
	return nil
}

func (u *usecase) deductInventory(ctx context.Context, tx *sqlx.Tx, checkoutUUID uuid.UUID, quantityByUUID map[string]int, productByUUID map[string]entity.Product) error {
	for uuidStr, qty := range quantityByUUID {
		product := productByUUID[uuidStr]
		if err := u.productRepository.DeductStock(ctx, tx, product.UUID, qty); err != nil {
			return err
		}
		movement := entity.InventoryMovement{
			ProductUUID:  product.UUID,
			SKU:          product.SKU,
			CheckoutUUID: uuid.NullUUID{UUID: checkoutUUID, Valid: true},
			MovementType: movementTypeCheckoutDeduct,
			Quantity:     -qty,
			StockBefore:  product.InventoryQty,
			StockAfter:   product.InventoryQty - qty,
			Note:         sql.NullString{String: "Checkout completed", Valid: true},
		}
		if err := u.inventoryRepository.CreateMovement(ctx, tx, &movement); err != nil {
			return err
		}
	}
	return nil
}

func appendLockedUUIDs(productUUIDs []uuid.UUID, rules []entity.PromotionRule) []uuid.UUID {
	seen := map[uuid.UUID]bool{}
	var allUUIDs []uuid.UUID
	for _, id := range productUUIDs {
		allUUIDs = append(allUUIDs, id)
		seen[id] = true
	}
	for _, rule := range rules {
		if rule.RewardProductUUID.Valid && !seen[rule.RewardProductUUID.UUID] {
			allUUIDs = append(allUUIDs, rule.RewardProductUUID.UUID)
			seen[rule.RewardProductUUID.UUID] = true
		}
	}
	return allUUIDs
}
