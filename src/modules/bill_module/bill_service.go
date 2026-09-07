package bill_module

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"maphraohom.app/maphraohom-backoffice/pkg/database/paginator"
	"maphraohom.app/maphraohom-backoffice/src/models"
	"maphraohom.app/maphraohom-backoffice/src/modules/bill_module/dtos"
	"maphraohom.app/maphraohom-backoffice/src/modules/bill_module/responses"
)

// billSlipPath is the MinIO/local-storage folder bill slips are kept under.
const billSlipPath = "bills/slips"

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

	return new(responses.BillDetailResponse).Make(bill), err
}

// CreateBill uploads the slip (if provided) then creates the bill. slip may
// be nil — the slip column is left empty in that case.
func (s Service) CreateBill(ctx context.Context, dto *dtos.CreateBill, slip *multipart.FileHeader) (*responses.BillDetailResponse, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "CreateBillService", trace.WithAttributes(attribute.String("service", "CreateBill")))

	var slipKey string
	if slip != nil {
		fileName := fmt.Sprintf("%s%s", uuid.New().String(), filepath.Ext(slip.Filename))
		if err := s.fileSystem.Put(ctx, billSlipPath, fileName, slip); err != nil {
			s.tracer.TraceEnd(childSpan)
			return nil, err
		}
		slipKey = fmt.Sprintf("%s/%s", billSlipPath, fileName)
	}

	bill, err := s.billRepository().CreateBill(ctx, CreateBillInput{
		ProductID:       dto.ProductID,
		StoreID:         dto.StoreID,
		CustomerName:    dto.CustomerName,
		CustomerAddress: dto.CustomerAddress,
		Kilogram:        dto.Kilogram,
		Discount:        dto.Discount,
		ShippingFee:     dto.ShippingFee,
		Slip:            slipKey,
	})

	s.tracer.TraceEnd(childSpan)

	if err != nil {
		return nil, err
	}

	return new(responses.BillDetailResponse).Make(bill), nil
}
