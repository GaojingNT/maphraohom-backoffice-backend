package customer_module

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"maphraohom.app/maphraohom-backoffice/pkg/database/paginator"
	"maphraohom.app/maphraohom-backoffice/src/models"
	"maphraohom.app/maphraohom-backoffice/src/modules/customer_module/dtos"
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

func (s Service) CreateCustomer(ctx context.Context, dto *dtos.CreateCustomer) (*responses.CustomerDetailResponse, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "CreateCustomerService", trace.WithAttributes(attribute.String("service", "CreateCustomer")))

	customer, err := s.customerRepository().CreateCustomer(ctx, dto.Name)

	s.tracer.TraceEnd(childSpan)

	if err != nil {
		return nil, err
	}

	return new(responses.CustomerDetailResponse).Make(customer), nil
}

func (s Service) CreateCustomerAddress(ctx context.Context, customerID int, dto *dtos.CreateCustomerAddress) (*responses.CustomerAddressItem, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "CreateCustomerAddressService", trace.WithAttributes(attribute.String("service", "CreateCustomerAddress")))

	address, err := s.customerRepository().CreateCustomerAddress(ctx, customerID, dto.Address, dto.Label, dto.IsDefault)

	s.tracer.TraceEnd(childSpan)

	if err != nil {
		return nil, err
	}

	item := responses.CustomerAddressItem{}.Make(address)
	return &item, nil
}

func (s Service) CreateCustomerPhone(ctx context.Context, customerID int, dto *dtos.CreateCustomerPhone) (*responses.CustomerPhoneItem, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "CreateCustomerPhoneService", trace.WithAttributes(attribute.String("service", "CreateCustomerPhone")))

	phone, err := s.customerRepository().CreateCustomerPhone(ctx, customerID, dto.Phone, dto.Label, dto.IsDefault)

	s.tracer.TraceEnd(childSpan)

	if err != nil {
		return nil, err
	}

	item := responses.CustomerPhoneItem{}.Make(phone)
	return &item, nil
}
