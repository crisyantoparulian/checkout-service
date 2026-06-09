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
	quantityBySKU, skus, err := buildCheckoutQuantities(req)
	if err != nil {
		return
	}

	rules, err := u.promotionRepository.GetActiveRulesByTargetSKUs(ctx, skus)
	if err != nil {
		return
	}

	skus = appendRewardSKUs(skus, rules)

	var (
		checkout          entity.Checkout
		checkoutItems     []entity.CheckoutItem
		appliedPromotions []appliedPromotion
	)

	err = u.txManager.WithTx(ctx, func(tx *sqlx.Tx) error {
		products, err := u.productRepository.GetBySKUsForUpdate(ctx, tx, skus)
		if err != nil {
			return err
		}

		productBySKU, err := validateCheckoutProducts(quantityBySKU, products)
		if err != nil {
			return err
		}

		appliedPromotions, discount, err := u.applyPromotions(quantityBySKU, productBySKU, rules)
		if err != nil {
			return err
		}

		checkoutItems, subtotal := buildCheckoutItems(req, productBySKU)

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

		if err := u.deductInventory(ctx, tx, checkout.UUID, quantityBySKU, productBySKU); err != nil {
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

func buildCheckoutQuantities(req CreateCheckoutRequest) (map[string]int, []string, error) {
	quantityBySKU := map[string]int{}
	skus := []string{}
	for _, item := range req.Items {
		if item.Quantity <= 0 {
			return nil, nil, apperror.New(http.StatusBadRequest, constants.CODE_INVALID_QUANTITY, errors.New("quantity must be greater than zero"))
		}
		if quantityBySKU[item.SKU] == 0 {
			skus = append(skus, item.SKU)
		}
		quantityBySKU[item.SKU] += item.Quantity
	}
	return quantityBySKU, skus, nil
}

func validateCheckoutProducts(quantityBySKU map[string]int, products []entity.Product) (map[string]entity.Product, error) {
	productBySKU := map[string]entity.Product{}
	for _, product := range products {
		productBySKU[product.SKU] = product
	}

	for sku, qty := range quantityBySKU {
		product, ok := productBySKU[sku]
		if !ok {
			return nil, apperror.New(http.StatusNotFound, constants.CODE_PRODUCT_NOT_FOUND, fmt.Errorf("product %s not found", sku))
		}
		if product.InventoryQty < qty {
			return nil, apperror.New(http.StatusConflict, constants.CODE_INSUFFICIENT_STOCK, fmt.Errorf("product %s only has %d items available", product.Name, product.InventoryQty))
		}
	}
	return productBySKU, nil
}

func buildCheckoutItems(req CreateCheckoutRequest, productBySKU map[string]entity.Product) ([]entity.CheckoutItem, int64) {
	checkoutItems := []entity.CheckoutItem{}
	var subtotal int64
	for _, item := range req.Items {
		product := productBySKU[item.SKU]
		totalPrice := product.PriceCents * int64(item.Quantity)
		subtotal += totalPrice
		checkoutItems = append(checkoutItems, entity.CheckoutItem{
			SKU:             product.SKU,
			ProductName:     product.Name,
			Quantity:        item.Quantity,
			UnitPriceCents:  product.PriceCents,
			TotalPriceCents: totalPrice,
		})
	}
	return checkoutItems, subtotal
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

func (u *usecase) deductInventory(ctx context.Context, tx *sqlx.Tx, checkoutUUID uuid.UUID, quantityBySKU map[string]int, productBySKU map[string]entity.Product) error {
	for sku, qty := range quantityBySKU {
		product := productBySKU[sku]
		if err := u.productRepository.DeductStock(ctx, tx, sku, qty); err != nil {
			return err
		}
		movement := entity.InventoryMovement{
			SKU:          sku,
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


func appendRewardSKUs(skus []string, rules []entity.PromotionRule) []string {
	seen := map[string]bool{}
	for _, s := range skus {
		seen[s] = true
	}
	for _, rule := range rules {
		if rule.RewardSKU.Valid && !seen[rule.RewardSKU.String] {
			skus = append(skus, rule.RewardSKU.String)
			seen[rule.RewardSKU.String] = true
		}
	}
	return skus
}
