package auth_module

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"maphraohom.app/maphraohom-backoffice/src/modules/auth_module/dtos"
	"maphraohom.app/maphraohom-backoffice/src/modules/auth_module/responses"
)

func (s Service) Authenticate(ctx context.Context, dto *dtos.SignInDto) (string, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "AuthenticateService", trace.WithAttributes(attribute.String("service", "Authenticate")))

	accessToken, err := s.authRepository().Authenticate(ctx, dto.Username, dto.Password)

	s.tracer.TraceEnd(childSpan)

	return accessToken, err
}

func (s Service) ForgotPassword(ctx context.Context, dto *dtos.ForgotPasswordDto) error {
	ctx, childSpan := s.tracer.TraceStart(ctx, "ForgotPasswordService", trace.WithAttributes(attribute.String("service", "ForgotPassword")))

	err := s.authRepository().ForgotPassword(ctx, dto.Email)

	s.tracer.TraceEnd(childSpan)

	return err
}

func (s Service) ResetPassword(ctx context.Context, dto *dtos.ResetPasswordDto) error {
	ctx, childSpan := s.tracer.TraceStart(ctx, "ResetPasswordService", trace.WithAttributes(attribute.String("service", "ResetPassword")))

	err := s.authRepository().ResetPassword(ctx, dto.Email, dto.Password, dto.ResetPasswordToken)

	s.tracer.TraceEnd(childSpan)

	return err
}

func (s Service) GetProfile(ctx context.Context, id int) (*responses.GetProfileResponse, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "GetProfileService", trace.WithAttributes(attribute.String("service", "GetProfile")))

	user, err := s.authRepository().GetUserByID(ctx, id)
	if err != nil {
		s.tracer.TraceEnd(childSpan)
		return nil, err
	}

	s.tracer.TraceEnd(childSpan)

	return new(responses.GetProfileResponse).Make(user), nil
}
