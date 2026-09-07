package bill_module

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"maphraohom.app/maphraohom-backoffice/pkg/database/paginator"
	"maphraohom.app/maphraohom-backoffice/src/models"
	"maphraohom.app/maphraohom-backoffice/src/modules/bill_module/responses"
)

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
