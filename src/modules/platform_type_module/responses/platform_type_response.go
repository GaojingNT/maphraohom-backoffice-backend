package responses

import (
	"fmt"

	"maphraohom.app/maphraohom-backoffice/src/models"
)

// PlatformTypeResponse is one sales platform — returned by GET
// /platform-types and embedded in bill list/detail responses.
type PlatformTypeResponse struct {
	ID   int    `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
	// LogoURL goes through the GET /api/v1/files/* proxy; "" when the
	// platform has no logo.
	LogoURL string `json:"logoUrl"`
}

// FileURLBuilder turns a stored logo object key into a URL the client can
// load — the existing GET /api/v1/files/* proxy, like store logos.
var FileURLBuilder = func(key string) string {
	return fmt.Sprintf("/api/v1/files/%s", key)
}

// Make returns nil for a nil platform (a bill without one).
func Make(platformType *models.PlatformType) *PlatformTypeResponse {
	if platformType == nil {
		return nil
	}
	logoURL := ""
	if platformType.Logo != "" {
		logoURL = FileURLBuilder(platformType.Logo)
	}
	return &PlatformTypeResponse{
		ID:      platformType.ID,
		Code:    platformType.Code,
		Name:    platformType.Name,
		LogoURL: logoURL,
	}
}

func Collection(platformTypes []models.PlatformType) []PlatformTypeResponse {
	items := make([]PlatformTypeResponse, 0, len(platformTypes))
	for i := range platformTypes {
		items = append(items, *Make(&platformTypes[i]))
	}
	return items
}
