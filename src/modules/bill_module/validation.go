package bill_module

import (
	"fmt"
	"strconv"
	"unicode/utf8"

	"maphraohom.app/maphraohom-backoffice/internal/exception"
	"maphraohom.app/maphraohom-backoffice/src/models"
	"maphraohom.app/maphraohom-backoffice/src/modules/bill_module/dtos"
)

// bottleUnit is the one unit that must be billed in whole numbers.
const bottleUnit = "ขวด"

// maxProductNameLength matches tbl_products.name's size.
const maxProductNameLength = 255

// itemUnit is the unit a bill item is billed in. On a payment bill the
// unit the user set wins (normalized like a product name), falling back to
// the product's own; a receipt bill always uses the product's unit.
// product is nil for a typed-in item.
func itemUnit(billType string, item dtos.CreateBillItem, product *models.Product) string {
	if billType == models.BillTypePayment {
		if unit := models.NormalizeProductName(item.Unit); unit != "" {
			return unit
		}
	}
	if product != nil {
		return product.Unit
	}
	return ""
}

// validateBillInput checks every field shared by create and update requests
// against business rules that need more than a struct tag: bill type
// whitelist, non-negative discount/shipping fee, at least one item, and per
// item quantity/price/product checks (including "ขวด" must be a whole
// number). products maps a referenced productId to its catalog row; a
// missing key means the id doesn't exist or was soft-deleted. A receipt bill
// may only use ProductScopeAll products; a typed-in item (productId 0) is
// allowed on a payment bill only, with a name and a unit. On a payment bill
// every item's unit can be set (see itemUnit); it must be non-empty and at
// most models.MaxUnitLength characters.
//
// Every error names the exact field/row it came from (e.g. "items[1].quantity")
// so the client can point the user at what to fix. Returns nil when the
// input is valid.
func validateBillInput(billType string, discount, shippingFee decimalLike, items []dtos.CreateBillItem, products map[int]models.Product) []exception.ParameterError {
	var errs []exception.ParameterError

	if !models.IsValidBillType(billType) {
		errs = append(errs, exception.ParameterError{FailedField: "type", Tag: "oneof=receipt payment", Value: billType})
	}

	if discount.IsNegative() {
		errs = append(errs, exception.ParameterError{FailedField: "discount", Tag: "gte=0", Value: discount.String()})
	}
	if shippingFee.IsNegative() {
		errs = append(errs, exception.ParameterError{FailedField: "shippingFee", Tag: "gte=0", Value: shippingFee.String()})
	}

	if len(items) == 0 {
		errs = append(errs, exception.ParameterError{FailedField: "items", Tag: "min=1", Value: "0"})
		return errs
	}

	for i, item := range items {
		field := fmt.Sprintf("items[%d]", i)

		var unit string
		if item.ProductID == 0 {
			if billType != models.BillTypePayment {
				errs = append(errs, exception.ParameterError{FailedField: field + ".productId", Tag: "required", Value: "0"})
				continue
			}
			name := models.NormalizeProductName(item.ProductName)
			if name == "" {
				errs = append(errs, exception.ParameterError{FailedField: field + ".productName", Tag: "required", Value: ""})
			} else if utf8.RuneCountInString(name) > maxProductNameLength {
				errs = append(errs, exception.ParameterError{FailedField: field + ".productName", Tag: fmt.Sprintf("max=%d", maxProductNameLength), Value: name})
			}
			unit = itemUnit(billType, item, nil)
		} else {
			product, ok := products[item.ProductID]
			if !ok {
				errs = append(errs, exception.ParameterError{FailedField: field + ".productId", Tag: "exists", Value: fmt.Sprintf("%d", item.ProductID)})
				continue
			}
			if billType == models.BillTypeReceipt && product.BillScope == models.ProductScopePayment {
				errs = append(errs, exception.ParameterError{FailedField: field + ".productId", Tag: "billScope=all", Value: fmt.Sprintf("%d", item.ProductID)})
				continue
			}
			unit = itemUnit(billType, item, &product)
		}

		if unit == "" {
			errs = append(errs, exception.ParameterError{FailedField: field + ".unit", Tag: "required", Value: ""})
		} else if utf8.RuneCountInString(unit) > models.MaxUnitLength {
			errs = append(errs, exception.ParameterError{FailedField: field + ".unit", Tag: fmt.Sprintf("max=%d", models.MaxUnitLength), Value: unit})
		}

		if !item.Quantity.IsPositive() {
			errs = append(errs, exception.ParameterError{FailedField: field + ".quantity", Tag: "gt=0", Value: item.Quantity.String()})
		} else if unit == bottleUnit && !item.Quantity.IsInteger() {
			errs = append(errs, exception.ParameterError{FailedField: field + ".quantity", Tag: "integer", Value: item.Quantity.String()})
		}

		if !item.Price.IsPositive() {
			errs = append(errs, exception.ParameterError{FailedField: field + ".price", Tag: "gt=0", Value: item.Price.String()})
		}
	}

	return errs
}

// PlatformFilterNone is the GET /bills?platformTypeId= value for "bills
// without a platform".
const PlatformFilterNone = "none"

// validateBillPlatform checks the platform rules that need no database:
// only a receipt bill may carry a platform, and an id must be positive.
// Whether the id exists is checked by the service. nil (no platform) is
// always valid.
func validateBillPlatform(billType string, platformTypeID *int) []exception.ParameterError {
	if platformTypeID == nil {
		return nil
	}
	value := fmt.Sprintf("%d", *platformTypeID)
	if billType != models.BillTypeReceipt {
		return []exception.ParameterError{{FailedField: "platformTypeId", Tag: "receiptOnly", Value: value}}
	}
	if *platformTypeID <= 0 {
		return []exception.ParameterError{{FailedField: "platformTypeId", Tag: "gt=0", Value: value}}
	}
	return nil
}

// validPlatformFilter reports whether a GET /bills platformTypeId query
// value is usable: empty (no filter), "none", or a positive integer.
func validPlatformFilter(value string) bool {
	if value == "" || value == PlatformFilterNone {
		return true
	}
	id, err := strconv.Atoi(value)
	return err == nil && id > 0
}

// decimalLike is the subset of decimal.Decimal's API validateBillInput
// needs — declared as an interface purely so its unit tests don't have to
// import shopspring/decimal just to build fixtures.
type decimalLike interface {
	IsNegative() bool
	String() string
}
