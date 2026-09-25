package auth_module

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"maphraohom.app/maphraohom-backoffice/internal/exception"
	"maphraohom.app/maphraohom-backoffice/pkg/imagevalidate"
	"maphraohom.app/maphraohom-backoffice/src/modules/auth_module/dtos"
	"maphraohom.app/maphraohom-backoffice/src/modules/auth_module/responses"
)

// userSignaturePath is the MinIO/local-storage folder user signature images
// are kept under, partitioned by day so no single folder grows unbounded.
const userSignaturePath = "user/signature"

// UpdateProfile replaces the signed-in user's first name, last name and
// email, then returns the refreshed profile.
func (s Service) UpdateProfile(ctx context.Context, id int, dto *dtos.UpdateProfileDto) (*responses.GetProfileResponse, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "UpdateProfileService", trace.WithAttributes(attribute.String("service", "UpdateProfile"), attribute.Int("id", id)))
	defer s.tracer.TraceEnd(childSpan)

	if err := s.authRepository().UpdateProfile(ctx, id, UpdateProfileInput{
		FirstName: strings.TrimSpace(dto.FirstName),
		LastName:  strings.TrimSpace(dto.LastName),
		Email:     strings.TrimSpace(dto.Email),
	}); err != nil {
		return nil, err
	}

	return s.GetProfile(ctx, id)
}

// UploadSignature validates (magic bytes + size), uploads, and attaches a
// signature image to the signed-in user, replacing any previous one. Same
// rollback/cleanup rules as a store logo: on a DB failure the new object is
// removed; on success the old object is removed best-effort.
func (s Service) UploadSignature(ctx context.Context, id int, file *multipart.FileHeader) (string, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "UploadSignatureService", trace.WithAttributes(attribute.String("service", "UploadSignature"), attribute.Int("id", id)))
	defer s.tracer.TraceEnd(childSpan)

	repo := s.authRepository()

	oldKey, err := repo.GetUserSignatureKey(ctx, id)
	if err != nil {
		return "", err
	}

	if file.Size > imagevalidate.MaxFileSize {
		return "", exception.ErrImageFileTooLarge
	}

	ext, err := imagevalidate.DetectExtension(file)
	if err != nil {
		if errors.Is(err, imagevalidate.ErrUnsupported) {
			return "", exception.ErrUnsupportedImageType
		}
		return "", err
	}

	folder := fmt.Sprintf("%s/%s", userSignaturePath, time.Now().Format("2006-01-02"))
	fileName := uuid.New().String() + ext
	key := fmt.Sprintf("%s/%s", folder, fileName)

	if err := s.fileSystem.Put(ctx, folder, fileName, file); err != nil {
		return "", err
	}

	if err := repo.UpdateUserSignature(ctx, id, key); err != nil {
		// Roll back the upload — best-effort, log-only on failure.
		_ = s.fileSystem.Delete(ctx, key)
		return "", err
	}

	if oldKey != "" && oldKey != key {
		if delErr := s.fileSystem.Delete(ctx, oldKey); delErr != nil {
			s.logger.Error(fmt.Sprintf("[AuthModule] failed to delete old signature object %q: %v", oldKey, delErr))
		}
	}

	return responses.FileURLBuilder(key), nil
}

// DeleteSignature clears the signed-in user's signature, then best-effort
// removes the underlying object. No signature set is left as-is (idempotent).
func (s Service) DeleteSignature(ctx context.Context, id int) error {
	ctx, childSpan := s.tracer.TraceStart(ctx, "DeleteSignatureService", trace.WithAttributes(attribute.String("service", "DeleteSignature"), attribute.Int("id", id)))
	defer s.tracer.TraceEnd(childSpan)

	repo := s.authRepository()

	oldKey, err := repo.GetUserSignatureKey(ctx, id)
	if err != nil {
		return err
	}

	if oldKey == "" {
		return nil
	}

	if err := repo.UpdateUserSignature(ctx, id, ""); err != nil {
		return err
	}

	if delErr := s.fileSystem.Delete(ctx, oldKey); delErr != nil {
		s.logger.Error(fmt.Sprintf("[AuthModule] failed to delete signature object %q: %v", oldKey, delErr))
	}

	return nil
}
