package customer_module

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"maphraohom.app/maphraohom-backoffice/pkg/database/paginator"
	"maphraohom.app/maphraohom-backoffice/src/models"
	"maphraohom.app/maphraohom-backoffice/src/modules/customer_module/responses"
)

func (s Service) GetCustomers(ctx context.Context, paginate *paginator.Pagination) (*paginator.Pagination, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "GetCustomersService", trace.WithAttributes(attribute.String("service", "GetCustomers")))

	result, err := s.customerRepository().GetCustomerPaginate(ctx, paginate)

	if result != nil && result.Data != nil {
		result.Data = responses.CustomerListItem{}.Collection(result.Data.([]models.Customer))
	}

	s.tracer.TraceEnd(childSpan)

	return result, err
}

func (s Service) GetCustomer(ctx context.Context, id int) (*responses.CustomerDetailResponse, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "GetCustomerService", trace.WithAttributes(attribute.String("service", "GetCustomer")))

	customer, err := s.customerRepository().GetCustomerByID(ctx, id)

	s.tracer.TraceEnd(childSpan)

	if err != nil {
		return nil, err
	}

	return new(responses.CustomerDetailResponse).Make(customer), nil
}

func (s Service) GetCustomerAddresses(ctx context.Context, customerID int) ([]responses.CustomerAddressItem, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "GetCustomerAddressesService", trace.WithAttributes(attribute.String("service", "GetCustomerAddresses")))

	addresses, err := s.customerRepository().GetCustomerAddresses(ctx, customerID)

	s.tracer.TraceEnd(childSpan)

	if err != nil {
		return nil, err
	}

	return responses.CustomerAddressItem{}.Collection(addresses), nil
}

func (s Service) GetCustomerPhones(ctx context.Context, customerID int) ([]responses.CustomerPhoneItem, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "GetCustomerPhonesService", trace.WithAttributes(attribute.String("service", "GetCustomerPhones")))

	phones, err := s.customerRepository().GetCustomerPhones(ctx, customerID)

	s.tracer.TraceEnd(childSpan)

	if err != nil {
		return nil, err
	}

	return responses.CustomerPhoneItem{}.Collection(phones), nil
}
