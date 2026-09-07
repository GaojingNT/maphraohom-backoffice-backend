package user_module

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"maphraohom.app/maphraohom-backoffice/pkg/database/paginator"
	"maphraohom.app/maphraohom-backoffice/src/models"
	"maphraohom.app/maphraohom-backoffice/src/modules/user_module/dtos"
	"maphraohom.app/maphraohom-backoffice/src/modules/user_module/responses"
)

func (s Service) GetUsers(ctx context.Context, paginate *paginator.Pagination) (*paginator.Pagination, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "GetUsersService", trace.WithAttributes(attribute.String("service", "GetUsers")))

	result, err := s.userRepository().GetUserPaginate(ctx, paginate)

	// Convert the result to a collection of responses
	if result.Data != nil {
		result.Data = new(responses.GetUsersCollection).Collection(result.Data.([]models.User))
	}

	s.tracer.TraceEnd(childSpan)

	return result, err
}

func (s Service) GetUser(ctx context.Context, id int) (*responses.GetUserByIDResponse, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "GetUserService", trace.WithAttributes(attribute.String("service", "GetUser")))

	user, err := s.userRepository().GetUserByID(ctx, id)

	s.tracer.TraceEnd(childSpan)

	return new(responses.GetUserByIDResponse).Make(user), err
}

func (s Service) CreateUser(ctx context.Context, userDto *dtos.CreateUser) error {
	ctx, childSpan := s.tracer.TraceStart(ctx, "CreateUserService", trace.WithAttributes(attribute.String("service", "CreateUser")))

	user := new(models.User)
	user.FirstName = userDto.FirstName
	user.LastName = userDto.LastName
	user.Email = userDto.Email
	user.RoleID = userDto.RoleID

	err := s.userRepository().CreateUser(ctx, user)

	s.tracer.TraceEnd(childSpan)

	return err
}

func (s Service) UpdateUser(ctx context.Context, id int, userDto *dtos.UpdateUser) error {
	ctx, childSpan := s.tracer.TraceStart(ctx, "UpdateUserService", trace.WithAttributes(attribute.String("service", "UpdateUser")))

	user := new(models.User)
	user.FirstName = userDto.FirstName
	user.LastName = userDto.LastName
	user.Email = userDto.Email
	user.RoleID = userDto.RoleID

	err := s.userRepository().UpdateUser(ctx, id, user)

	s.tracer.TraceEnd(childSpan)

	return err
}

func (s Service) DeleteUser(ctx context.Context, id int) error {
	ctx, childSpan := s.tracer.TraceStart(ctx, "DeleteUserService", trace.WithAttributes(attribute.String("service", "DeleteUser")))

	err := s.userRepository().DeleteUser(ctx, id)

	s.tracer.TraceEnd(childSpan)

	return err
}
