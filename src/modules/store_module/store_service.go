package store_module

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"maphraohom.app/maphraohom-backoffice/internal/exception"
	"maphraohom.app/maphraohom-backoffice/pkg/database/paginator"
	"maphraohom.app/maphraohom-backoffice/pkg/imagevalidate"
	"maphraohom.app/maphraohom-backoffice/src/models"
	"maphraohom.app/maphraohom-backoffice/src/modules/store_module/dtos"
	"maphraohom.app/maphraohom-backoffice/src/modules/store_module/responses"
)

// storeImagePath is the MinIO/local-storage folder store logo/signature
// images are kept under, partitioned by kind ("logo" or "signature") and
// day so no single folder grows unbounded.
const storeImagePath = "store"

func (s Service) GetStores(ctx context.Context, paginate *paginator.Pagination) (*paginator.Pagination, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "GetStoresService", trace.WithAttributes(attribute.String("service", "GetStores")))

	result, err := s.storeRepository().GetStorePaginate(ctx, paginate)

	if result != nil && result.Data != nil {
		result.Data = responses.StoreListItem{}.Collection(result.Data.([]models.Store))
	}

	s.tracer.TraceEnd(childSpan)

	return result, err
}

func (s Service) GetStore(ctx context.Context, id int) (*responses.StoreDetailResponse, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "GetStoreService", trace.WithAttributes(attribute.String("service", "GetStore")))

	store, err := s.storeRepository().GetStoreByID(ctx, id)

	s.tracer.TraceEnd(childSpan)

	if err != nil {
		return nil, err
	}

	return new(responses.StoreDetailResponse).Make(store), nil
}

// UpdateStore replaces a store's name, address, and phone. Logo/Signature
// are untouched here — see UploadLogo/UploadSignature.
func (s Service) UpdateStore(ctx context.Context, id int, dto *dtos.UpdateStore) (*responses.StoreDetailResponse, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "UpdateStoreService", trace.WithAttributes(attribute.String("service", "UpdateStore")))
	defer s.tracer.TraceEnd(childSpan)

	store, err := s.storeRepository().UpdateStore(ctx, id, UpdateStoreInput{
		Name:    dto.Name,
		Address: dto.Address,
		Phone:   dto.Phone,
	})
	if err != nil {
		return nil, err
	}

	return new(responses.StoreDetailResponse).Make(store), nil
}

// uploadStoreImage validates (magic bytes + size), uploads, and attaches a
// logo/signature image to an existing store. kind is "logo" or "signature" —
// it names the storage sub-folder and picks which repository getter/setter
// to call. On a DB failure after a successful upload, the newly uploaded
// object is removed. On success, any previous image object is removed
// best-effort (logged, never fails the request).
func (s Service) uploadStoreImage(ctx context.Context, id int, file *multipart.FileHeader, kind string, getKey func(context.Context, int) (string, error), setKey func(context.Context, int, string) error) (string, error) {
	oldKey, err := getKey(ctx, id)
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

	folder := fmt.Sprintf("%s/%s/%s", storeImagePath, kind, time.Now().Format("2006-01-02"))
	fileName := uuid.New().String() + ext
	key := fmt.Sprintf("%s/%s", folder, fileName)

	if err := s.fileSystem.Put(ctx, folder, fileName, file); err != nil {
		return "", err
	}

	if err := setKey(ctx, id, key); err != nil {
		// Roll back the upload — best-effort, log-only on failure.
		_ = s.fileSystem.Delete(ctx, key)
		return "", err
	}

	if oldKey != "" && oldKey != key {
		// Best-effort: a leftover orphaned object is a cleanup nuisance, not
		// a correctness problem, so a failure here must not fail the request.
		if delErr := s.fileSystem.Delete(ctx, oldKey); delErr != nil {
			s.log.Error(fmt.Sprintf("[StoreModule] failed to delete old %s object %q: %v", kind, oldKey, delErr))
		}
	}

	return responses.FileURLBuilder(key), nil
}

// deleteStoreImage clears a store's logo/signature, then best-effort removes
// the underlying object. A store with no image set is left as-is (idempotent).
func (s Service) deleteStoreImage(ctx context.Context, id int, kind string, getKey func(context.Context, int) (string, error), setKey func(context.Context, int, string) error) error {
	oldKey, err := getKey(ctx, id)
	if err != nil {
		return err
	}

	if oldKey == "" {
		return nil
	}

	if err := setKey(ctx, id, ""); err != nil {
		return err
	}

	if delErr := s.fileSystem.Delete(ctx, oldKey); delErr != nil {
		s.log.Error(fmt.Sprintf("[StoreModule] failed to delete %s object %q: %v", kind, oldKey, delErr))
	}

	return nil
}

func (s Service) UploadLogo(ctx context.Context, id int, file *multipart.FileHeader) (string, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "UploadLogoService", trace.WithAttributes(attribute.String("service", "UploadLogo"), attribute.Int("id", id)))
	defer s.tracer.TraceEnd(childSpan)

	repo := s.storeRepository()
	return s.uploadStoreImage(ctx, id, file, "logo", repo.GetStoreLogoKey, repo.UpdateStoreLogo)
}

func (s Service) DeleteLogo(ctx context.Context, id int) error {
	ctx, childSpan := s.tracer.TraceStart(ctx, "DeleteLogoService", trace.WithAttributes(attribute.String("service", "DeleteLogo"), attribute.Int("id", id)))
	defer s.tracer.TraceEnd(childSpan)

	repo := s.storeRepository()
	return s.deleteStoreImage(ctx, id, "logo", repo.GetStoreLogoKey, repo.UpdateStoreLogo)
}

func (s Service) UploadSignature(ctx context.Context, id int, file *multipart.FileHeader) (string, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "UploadSignatureService", trace.WithAttributes(attribute.String("service", "UploadSignature"), attribute.Int("id", id)))
	defer s.tracer.TraceEnd(childSpan)

	repo := s.storeRepository()
	return s.uploadStoreImage(ctx, id, file, "signature", repo.GetStoreSignatureKey, repo.UpdateStoreSignature)
}

func (s Service) DeleteSignature(ctx context.Context, id int) error {
	ctx, childSpan := s.tracer.TraceStart(ctx, "DeleteSignatureService", trace.WithAttributes(attribute.String("service", "DeleteSignature"), attribute.Int("id", id)))
	defer s.tracer.TraceEnd(childSpan)

	repo := s.storeRepository()
	return s.deleteStoreImage(ctx, id, "signature", repo.GetStoreSignatureKey, repo.UpdateStoreSignature)
}

// GetLastPrices returns, for every product, the price used in this store's
// most recent bill of the given type — used to prefill the create-bill form.
func (s Service) GetLastPrices(ctx context.Context, storeID int, billType string) ([]responses.LastPriceItem, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "GetLastPricesService", trace.WithAttributes(attribute.String("service", "GetLastPrices")))

	rows, err := s.storeRepository().GetLastPrices(ctx, storeID, billType)

	s.tracer.TraceEnd(childSpan)

	if err != nil {
		return nil, err
	}

	items := make([]responses.LastPriceItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, responses.LastPriceItem{
			ProductID: row.ProductID,
			Price:     row.Price,
		})
	}

	return items, nil
}
