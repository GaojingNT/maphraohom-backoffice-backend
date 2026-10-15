package bill_module

import (
	"testing"

	"maphraohom.app/maphraohom-backoffice/internal/exception"
	"maphraohom.app/maphraohom-backoffice/src/models"
	"maphraohom.app/maphraohom-backoffice/src/modules/bill_module/dtos"
)

var testProducts = map[int]models.Product{
	1: {BaseModel: models.BaseModel{ID: 1}, Name: "เนื้อพับ 1.5 ชั้น", Unit: "กก."},
	2: {BaseModel: models.BaseModel{ID: 2}, Name: "น้ำมะพร้าว", Unit: "ขวด"},
}

func hasFieldError(errs []exception.ParameterError, field string) bool {
	for _, e := range errs {
		if e.FailedField == field {
			return true
		}
	}
	return false
}

func TestValidateBillInput_Valid(t *testing.T) {
	items := []dtos.CreateBillItem{
		{ProductID: 1, Quantity: d("20.5"), Price: d("80")},
		{ProductID: 2, Quantity: d("10"), Price: d("25")},
	}

	errs := validateBillInput(models.BillTypeReceipt, d("0"), d("50"), items, testProducts)
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got %+v", errs)
	}
}

func TestValidateBillInput_InvalidType(t *testing.T) {
	items := []dtos.CreateBillItem{{ProductID: 1, Quantity: d("1"), Price: d("1")}}

	errs := validateBillInput("refund", d("0"), d("0"), items, testProducts)
	if !hasFieldError(errs, "type") {
		t.Fatalf("expected a \"type\" field error, got %+v", errs)
	}
}

func TestValidateBillInput_NoItems(t *testing.T) {
	errs := validateBillInput(models.BillTypeReceipt, d("0"), d("0"), nil, testProducts)
	if !hasFieldError(errs, "items") {
		t.Fatalf("expected an \"items\" field error, got %+v", errs)
	}
}

func TestValidateBillInput_NegativeDiscountAndShipping(t *testing.T) {
	items := []dtos.CreateBillItem{{ProductID: 1, Quantity: d("1"), Price: d("1")}}

	errs := validateBillInput(models.BillTypeReceipt, d("-1"), d("-5"), items, testProducts)
	if !hasFieldError(errs, "discount") {
		t.Errorf("expected a \"discount\" field error, got %+v", errs)
	}
	if !hasFieldError(errs, "shippingFee") {
		t.Errorf("expected a \"shippingFee\" field error, got %+v", errs)
	}
}

func TestValidateBillInput_UnknownProduct(t *testing.T) {
	items := []dtos.CreateBillItem{{ProductID: 999, Quantity: d("1"), Price: d("1")}}

	errs := validateBillInput(models.BillTypeReceipt, d("0"), d("0"), items, testProducts)
	if !hasFieldError(errs, "items[0].productId") {
		t.Fatalf("expected an \"items[0].productId\" field error, got %+v", errs)
	}
}

func TestValidateBillInput_NonPositiveQuantityAndPrice(t *testing.T) {
	items := []dtos.CreateBillItem{
		{ProductID: 1, Quantity: d("0"), Price: d("0")},
		{ProductID: 1, Quantity: d("-5"), Price: d("10")},
	}

	errs := validateBillInput(models.BillTypeReceipt, d("0"), d("0"), items, testProducts)
	if !hasFieldError(errs, "items[0].quantity") {
		t.Errorf("expected \"items[0].quantity\" error, got %+v", errs)
	}
	if !hasFieldError(errs, "items[0].price") {
		t.Errorf("expected \"items[0].price\" error, got %+v", errs)
	}
	if !hasFieldError(errs, "items[1].quantity") {
		t.Errorf("expected \"items[1].quantity\" error, got %+v", errs)
	}
}

func TestValidateBillInput_BottleMustBeWholeNumber(t *testing.T) {
	items := []dtos.CreateBillItem{
		{ProductID: 2, Quantity: d("10.5"), Price: d("25")}, // ขวด — fractional, invalid
	}

	errs := validateBillInput(models.BillTypeReceipt, d("0"), d("0"), items, testProducts)
	if !hasFieldError(errs, "items[0].quantity") {
		t.Fatalf("expected \"items[0].quantity\" error for fractional bottle count, got %+v", errs)
	}
}

func TestValidateBillInput_BottleWholeNumberIsValid(t *testing.T) {
	items := []dtos.CreateBillItem{
		{ProductID: 2, Quantity: d("10"), Price: d("25")},
	}

	errs := validateBillInput(models.BillTypeReceipt, d("0"), d("0"), items, testProducts)
	if len(errs) != 0 {
		t.Fatalf("expected no errors for a whole bottle count, got %+v", errs)
	}
}

func TestValidateBillInput_KilogramFractionIsValid(t *testing.T) {
	items := []dtos.CreateBillItem{
		{ProductID: 1, Quantity: d("20.555"), Price: d("80")}, // กก. — fractional is fine
	}

	errs := validateBillInput(models.BillTypeReceipt, d("0"), d("0"), items, testProducts)
	if len(errs) != 0 {
		t.Fatalf("expected no errors for a fractional กก. quantity, got %+v", errs)
	}
}

func TestValidateBillInput_MultipleRowErrorsAreAllReported(t *testing.T) {
	items := []dtos.CreateBillItem{
		{ProductID: 1, Quantity: d("1"), Price: d("1")},   // valid
		{ProductID: 2, Quantity: d("1.5"), Price: d("0")}, // bad quantity (bottle) and price
		{ProductID: 999, Quantity: d("1"), Price: d("1")}, // unknown product
	}

	errs := validateBillInput(models.BillTypeReceipt, d("0"), d("0"), items, testProducts)

	for _, field := range []string{"items[1].quantity", "items[1].price", "items[2].productId"} {
		if !hasFieldError(errs, field) {
			t.Errorf("expected error for %q, got %+v", field, errs)
		}
	}
	if hasFieldError(errs, "items[0].quantity") || hasFieldError(errs, "items[0].price") {
		t.Errorf("row 0 is valid and should not have an error, got %+v", errs)
	}
}

var scopedProducts = map[int]models.Product{
	1: {BaseModel: models.BaseModel{ID: 1}, Name: "เนื้อพับ 1.5 ชั้น", Unit: "กก.", BillScope: models.ProductScopeAll},
	3: {BaseModel: models.BaseModel{ID: 3}, Name: "ค่าขนส่ง", Unit: "ครั้ง", BillScope: models.ProductScopePayment},
}

// A payment bill can mix catalog products, payment-only products and
// typed-in items.
func TestValidateBillInput_PaymentTypedItems(t *testing.T) {
	items := []dtos.CreateBillItem{
		{ProductID: 1, Quantity: d("2"), Price: d("80")},
		{ProductID: 3, Quantity: d("1"), Price: d("300")},
		{ProductName: "  ค่า  น้ำแข็ง ", Unit: "ถุง", Quantity: d("4"), Price: d("20")},
	}

	if errs := validateBillInput(models.BillTypePayment, d("0"), d("0"), items, scopedProducts); len(errs) != 0 {
		t.Fatalf("expected no errors, got %+v", errs)
	}
}

func TestValidateBillInput_TypedItemRules(t *testing.T) {
	items := []dtos.CreateBillItem{
		{ProductName: "   ", Unit: "ถุง", Quantity: d("1"), Price: d("1")},
		{ProductName: "ค่าแรง", Unit: "วัน", Quantity: d("1"), Price: d("1")},
		{ProductName: "น้ำดื่ม", Unit: "ขวด", Quantity: d("1.5"), Price: d("1")},
	}

	errs := validateBillInput(models.BillTypePayment, d("0"), d("0"), items, scopedProducts)
	for _, field := range []string{"items[0].productName", "items[1].unit", "items[2].quantity"} {
		if !hasFieldError(errs, field) {
			t.Errorf("expected a %q field error, got %+v", field, errs)
		}
	}
}

// Receipt bills only take "all" products: no typed-in items, no
// payment-only products.
func TestValidateBillInput_ReceiptRejectsPaymentItems(t *testing.T) {
	items := []dtos.CreateBillItem{
		{ProductID: 1, Quantity: d("1"), Price: d("1")},
		{ProductID: 3, Quantity: d("1"), Price: d("1")},
		{ProductName: "ค่าขนส่ง", Unit: "ครั้ง", Quantity: d("1"), Price: d("1")},
	}

	errs := validateBillInput(models.BillTypeReceipt, d("0"), d("0"), items, scopedProducts)
	if hasFieldError(errs, "items[0].productId") {
		t.Errorf("catalog product rejected on a receipt bill: %+v", errs)
	}
	for _, field := range []string{"items[1].productId", "items[2].productId"} {
		if !hasFieldError(errs, field) {
			t.Errorf("expected a %q field error, got %+v", field, errs)
		}
	}
}
