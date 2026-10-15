package bill_module

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"maphraohom.app/maphraohom-backoffice/internal/exception"
	"maphraohom.app/maphraohom-backoffice/src/models"
	"maphraohom.app/maphraohom-backoffice/src/modules/bill_module/dtos"
)

// bottleUnit is the one unit that must be billed in whole numbers.
const bottleUnit = "ขวด"

// maxProductNameLength matches tbl_products.name's size.
const maxProductNameLength = 255

// validateBillInput checks every field shared by create and update requests
// against business rules that need more than a struct tag: bill type
// whitelist, non-negative discount/shipping fee, at least one item, and per
// item quantity/price/product checks (including "ขวด" must be a whole
// number). products maps a referenced productId to its catalog row; a
// missing key means the id doesn't exist or was soft-deleted. A receipt bill
// may only use ProductScopeAll products; a typed-in item (productId 0) is
// allowed on a payment bill only, with a name and one of
// models.PaymentItemUnits.
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
			if !models.IsValidPaymentItemUnit(item.Unit) {
				errs = append(errs, exception.ParameterError{FailedField: field + ".unit", Tag: "oneof=" + strings.Join(models.PaymentItemUnits, " "), Value: item.Unit})
			}
			unit = item.Unit
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
			unit = product.Unit
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

// decimalLike is the subset of decimal.Decimal's API validateBillInput
// needs — declared as an interface purely so its unit tests don't have to
// import shopspring/decimal just to build fixtures.
type decimalLike interface {
	IsNegative() bool
	String() string
}
