package bill_module

import (
	"context"
	"fmt"
	"mime/multipart"
	"time"

	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"maphraohom.app/maphraohom-backoffice/internal/exception"
	"maphraohom.app/maphraohom-backoffice/pkg/database/paginator"
	"maphraohom.app/maphraohom-backoffice/src/models"
	"maphraohom.app/maphraohom-backoffice/src/modules/bill_module/dtos"
	"maphraohom.app/maphraohom-backoffice/src/modules/bill_module/responses"
)

// billSlipPath is the MinIO/local-storage folder bill slips are kept under,
// partitioned by day so no single folder grows unbounded.
const billSlipPath = "slip"

// maxSlipFileSize is the 10MB cap on a slip upload.
const maxSlipFileSize = 10 * 1024 * 1024

// slipExtByMIME maps a magic-byte-detected MIME type to the file extension
// its object key is stored under — only these three are accepted.
var slipExtByMIME = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

func (s Service) GetBills(ctx context.Context, paginate *paginator.Pagination) (*paginator.Pagination, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "GetBillsService", trace.WithAttributes(attribute.String("service", "GetBills")))

	result, err := s.billRepository().GetBillPaginate(ctx, paginate)

	// Convert the result to a collection of responses
	if result != nil && result.Data != nil {
		result.Data = responses.BillListItem{}.Collection(result.Data.([]models.Bill))
	}

	s.tracer.TraceEnd(childSpan)

	return result, err
}

func (s Service) GetBill(ctx context.Context, id int) (*responses.BillDetailResponse, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "GetBillService", trace.WithAttributes(attribute.String("service", "GetBill")))

	bill, err := s.billRepository().GetBillByID(ctx, id)

	s.tracer.TraceEnd(childSpan)

	if err != nil {
		return nil, err
	}

	return new(responses.BillDetailResponse).Make(bill), nil
}

// resolveAndPriceItems validates dto-level item input against the product
// catalog and returns the priced CreateBillItemInput slice plus its total
// subtotal-sum decimal errs (if any) name the exact field/row that's wrong.
func (s Service) resolveAndPriceItems(ctx context.Context, items []dtos.CreateBillItem, products map[int]models.Product) []CreateBillItemInput {
	priced := make([]CreateBillItemInput, 0, len(items))
	for _, item := range items {
		product, ok := products[item.ProductID]
		if !ok {
			// Already reported by validateBillInput — skip so pricing
			// doesn't panic on a zero-value Product.
			continue
		}
		priced = append(priced, CreateBillItemInput{
			ProductID: item.ProductID,
			Unit:      product.Unit,
			Quantity:  item.Quantity,
			Price:     item.Price,
			Subtotal:  computeSubtotal(item.Quantity, item.Price),
		})
	}
	return priced
}

// validateAndPrice runs every check shared by create/update (type, item
// rules, product/customer existence, and the total >= 0 rule), returning the
// priced items and total on success, or the list of field errors otherwise.
func (s Service) validateAndPrice(ctx context.Context, billType string, customerID *int, discount decimal.Decimal, shippingFee decimal.Decimal, items []dtos.CreateBillItem) ([]CreateBillItemInput, decimal.Decimal, []exception.ParameterError, error) {
	productIDs := make([]int, 0, len(items))
	seen := make(map[int]bool, len(items))
	for _, item := range items {
		if !seen[item.ProductID] {
			seen[item.ProductID] = true
			productIDs = append(productIDs, item.ProductID)
		}
	}

	products, err := s.billRepository().GetProductsByID(ctx, productIDs)
	if err != nil {
		return nil, decimal.Zero, nil, err
	}

	fieldErrors := validateBillInput(billType, discount, shippingFee, items, products)

	if customerID != nil {
		exists, err := s.billRepository().CustomerExists(ctx, *customerID)
		if err != nil {
			return nil, decimal.Zero, nil, err
		}
		if !exists {
			fieldErrors = append(fieldErrors, exception.ParameterError{FailedField: "customerId", Tag: "exists", Value: fmt.Sprintf("%d", *customerID)})
		}
	}

	if len(fieldErrors) > 0 {
		return nil, decimal.Zero, fieldErrors, nil
	}

	pricedItems := s.resolveAndPriceItems(ctx, items, products)

	subtotals := make([]decimal.Decimal, 0, len(pricedItems))
	for _, item := range pricedItems {
		subtotals = append(subtotals, item.Subtotal)
	}
	total := computeTotal(subtotals, discount, shippingFee)

	if total.IsNegative() {
		fieldErrors = append(fieldErrors, exception.ParameterError{FailedField: "total", Tag: "gte=0", Value: total.String()})
		return nil, decimal.Zero, fieldErrors, nil
	}

	return pricedItems, total, nil, nil
}

// CreateBill validates, prices, and creates a bill. Returns field errors
// (for a 400 response) when the request itself is invalid; err is reserved
// for infrastructure failures.
func (s Service) CreateBill(ctx context.Context, dto *dtos.CreateBill) (*responses.BillDetailResponse, []exception.ParameterError, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "CreateBillService", trace.WithAttributes(attribute.String("service", "CreateBill")))
	defer s.tracer.TraceEnd(childSpan)

	billType := dto.Type
	if billType == "" {
		billType = models.BillTypeReceipt
	}

	pricedItems, total, fieldErrors, err := s.validateAndPrice(ctx, billType, dto.CustomerID, dto.Discount, dto.ShippingFee, dto.Items)
	if err != nil || len(fieldErrors) > 0 {
		return nil, fieldErrors, err
	}

	bill, err := s.billRepository().CreateBill(ctx, CreateBillInput{
		Type:            billType,
		StoreID:         dto.StoreID,
		CustomerID:      dto.CustomerID,
		CustomerName:    dto.CustomerName,
		CustomerAddress: dto.CustomerAddress,
		CustomerPhone:   dto.CustomerPhone,
		Discount:        dto.Discount,
		ShippingFee:     dto.ShippingFee,
		Total:           total,
		Items:           pricedItems,
	})
	if err != nil {
		return nil, nil, err
	}

	return new(responses.BillDetailResponse).Make(bill), nil, nil
}

// UpdateBill validates, prices, and replaces a bill's editable fields and
// items. Returns field errors (400) for invalid input, including when
// storeId/type don't match the existing bill.
func (s Service) UpdateBill(ctx context.Context, id int, dto *dtos.UpdateBill) (*responses.BillDetailResponse, []exception.ParameterError, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "UpdateBillService", trace.WithAttributes(attribute.String("service", "UpdateBill")))
	defer s.tracer.TraceEnd(childSpan)

	billType := dto.Type
	if billType == "" {
		billType = models.BillTypeReceipt
	}

	pricedItems, total, fieldErrors, err := s.validateAndPrice(ctx, billType, dto.CustomerID, dto.Discount, dto.ShippingFee, dto.Items)
	if err != nil || len(fieldErrors) > 0 {
		return nil, fieldErrors, err
	}

	bill, err := s.billRepository().UpdateBill(ctx, id, UpdateBillInput{
		Type:            billType,
		StoreID:         dto.StoreID,
		CustomerID:      dto.CustomerID,
		CustomerName:    dto.CustomerName,
		CustomerAddress: dto.CustomerAddress,
		CustomerPhone:   dto.CustomerPhone,
		Discount:        dto.Discount,
		ShippingFee:     dto.ShippingFee,
		Total:           total,
		Items:           pricedItems,
	})
	if err != nil {
		return nil, nil, err
	}

	return new(responses.BillDetailResponse).Make(bill), nil, nil
}

func (s Service) DeleteBill(ctx context.Context, id int) error {
	ctx, childSpan := s.tracer.TraceStart(ctx, "DeleteBillService", trace.WithAttributes(attribute.String("service", "DeleteBill")))

	err := s.billRepository().DeleteBill(ctx, id)

	s.tracer.TraceEnd(childSpan)

	return err
}

// UploadSlip validates (magic bytes + size), uploads, and attaches a slip to
// an existing bill. On a DB failure after a successful upload, the newly
// uploaded object is removed. On success, any previous slip object is
// removed best-effort (logged, never fails the request).
func (s Service) UploadSlip(ctx context.Context, id int, file *multipart.FileHeader) (string, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "UploadSlipService", trace.WithAttributes(attribute.String("service", "UploadSlip"), attribute.Int("id", id)))
	defer s.tracer.TraceEnd(childSpan)

	oldKey, err := s.billRepository().GetBillSlipKey(ctx, id)
	if err != nil {
		return "", err
	}

	if file.Size > maxSlipFileSize {
		return "", exception.ErrSlipFileTooLarge
	}

	ext, err := detectSlipExtension(file)
	if err != nil {
		return "", err
	}

	folder := fmt.Sprintf("%s/%s", billSlipPath, time.Now().Format("2006-01-02"))
	fileName := uuid.New().String() + ext
	key := fmt.Sprintf("%s/%s", folder, fileName)

	if err := s.fileSystem.Put(ctx, folder, fileName, file); err != nil {
		return "", err
	}

	if err := s.billRepository().UpdateBillSlip(ctx, id, &key); err != nil {
		// Roll back the upload — best-effort, log-only on failure.
		_ = s.fileSystem.Delete(ctx, key)
		return "", err
	}

	if oldKey != nil && *oldKey != "" && *oldKey != key {
		// Best-effort: a leftover orphaned object is a cleanup nuisance, not
		// a correctness problem, so a failure here must not fail the request.
		if delErr := s.fileSystem.Delete(ctx, *oldKey); delErr != nil {
			s.log.Error(fmt.Sprintf("[BillModule] failed to delete old slip object %q: %v", *oldKey, delErr))
		}
	}

	return responses.SlipURLBuilder(key), nil
}

// DeleteSlip clears a bill's slip, then best-effort removes the underlying
// object. A bill with no slip is left as-is (idempotent).
func (s Service) DeleteSlip(ctx context.Context, id int) error {
	ctx, childSpan := s.tracer.TraceStart(ctx, "DeleteSlipService", trace.WithAttributes(attribute.String("service", "DeleteSlip"), attribute.Int("id", id)))
	defer s.tracer.TraceEnd(childSpan)

	oldKey, err := s.billRepository().GetBillSlipKey(ctx, id)
	if err != nil {
		return err
	}

	if oldKey == nil || *oldKey == "" {
		return nil
	}

	if err := s.billRepository().UpdateBillSlip(ctx, id, nil); err != nil {
		return err
	}

	if delErr := s.fileSystem.Delete(ctx, *oldKey); delErr != nil {
		s.log.Error(fmt.Sprintf("[BillModule] failed to delete slip object %q: %v", *oldKey, delErr))
	}

	return nil
}

// detectSlipExtension reads the file's magic bytes (never trusting the
// client-supplied filename/Content-Type) and returns the extension to store
// it under, or exception.ErrUnsupportedSlipType if it isn't jpeg/png/webp.
func detectSlipExtension(file *multipart.FileHeader) (string, error) {
	f, err := file.Open()
	if err != nil {
		return "", err
	}
	defer f.Close()

	mtype, err := mimetype.DetectReader(f)
	if err != nil {
		return "", err
	}

	// mimetype.Is walks the detected type's parent chain (e.g. some jpeg
	// variants), so match by string on the parents too.
	for m := mtype; m != nil; m = m.Parent() {
		if ext, ok := slipExtByMIME[m.String()]; ok {
			return ext, nil
		}
	}

	return "", exception.ErrUnsupportedSlipType
}
