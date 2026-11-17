package database

import (
	"context"
	"embed"
	"log"
	"os"
	"path"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"maphraohom.app/maphraohom-backoffice/pkg/storage"
	"maphraohom.app/maphraohom-backoffice/src/models"
)

//go:embed platform_logos/*
var platformLogos embed.FS

// platformLogoFolder is the storage folder seeded platform logos go in —
// served like every other file through GET /api/v1/files/<key>.
const platformLogoFolder = "platform-types"

type platformTypeSeed struct {
	Code      string
	Name      string
	LogoFile  string // file name inside platform_logos/
	SortOrder int
}

var platformTypeSeeds = []platformTypeSeed{
	{Code: "facebook", Name: "Facebook", LogoFile: "facebook.png", SortOrder: 1},
	{Code: "tiktok", Name: "TikTok", LogoFile: "tiktok.jpg", SortOrder: 2},
	{Code: "lazada", Name: "Lazada", LogoFile: "lazada.jpg", SortOrder: 3},
	{Code: "shopee", Name: "Shopee", LogoFile: "shopee.png", SortOrder: 4},
	{Code: "thaimart", Name: "Thaimart", LogoFile: "thaimart.png", SortOrder: 5},
}

// SeedPlatformTypes makes sure every built-in platform exists, run once on
// app start after storage is ready.
//
// Insert-only and idempotent: a platform whose code already exists is left
// exactly as it is (so a later rename/logo change/deactivation sticks), and
// a logo is uploaded only when its object is missing from storage. It never
// touches tbl_bills or any other table.
//
// A storage failure is logged, not fatal — the app still serves bills, the
// logo just doesn't load until the next start succeeds.
func SeedPlatformTypes(db *gorm.DB, fs *storage.FileSystem) {
	ctx := context.Background()

	for _, seed := range platformTypeSeeds {
		key := path.Join(platformLogoFolder, seed.LogoFile)

		row := models.PlatformType{
			Code:      seed.Code,
			Name:      seed.Name,
			Logo:      key,
			SortOrder: seed.SortOrder,
			IsActive:  true,
		}
		if err := db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "code"}},
			DoNothing: true,
		}).Create(&row).Error; err != nil {
			log.Fatalf("[App] Seed platform type %q: %v", seed.Code, err)
		}

		if fs == nil {
			continue
		}
		if err := ensurePlatformLogo(ctx, fs, seed.LogoFile); err != nil {
			log.Printf("[App] Seed platform logo %q: %v", key, err)
		}
	}
}

// ensurePlatformLogo uploads the embedded logo to platform-types/<file>
// unless an object is already there.
func ensurePlatformLogo(ctx context.Context, fs *storage.FileSystem, fileName string) error {
	key := path.Join(platformLogoFolder, fileName)

	existing, err := fs.Get(ctx, key)
	if err == nil {
		existing.Close()
		return nil
	}
	if !os.IsNotExist(err) {
		return err
	}

	content, err := platformLogos.ReadFile(path.Join("platform_logos", fileName))
	if err != nil {
		return err
	}

	// storage.Put takes a *multipart.FileHeader or *os.File — stage the
	// embedded bytes in a temp file.
	tmp, err := os.CreateTemp("", "platform-logo-*-"+fileName)
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()

	if _, err = tmp.Write(content); err != nil {
		return err
	}
	if _, err = tmp.Seek(0, 0); err != nil {
		return err
	}

	return fs.Put(ctx, platformLogoFolder, fileName, tmp)
}
