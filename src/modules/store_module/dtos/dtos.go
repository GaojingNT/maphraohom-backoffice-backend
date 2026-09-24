package dtos

// UpdateStore is the JSON body of PUT /api/v1/stores/:id. Logo and Signature
// are never set here — they're uploaded through their own endpoints
// (PUT /api/v1/stores/:id/logo, PUT /api/v1/stores/:id/signature), same
// split as a bill and its slip.
type UpdateStore struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Phone   string `json:"phone"`
}
