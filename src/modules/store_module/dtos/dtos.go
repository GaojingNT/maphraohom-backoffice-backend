package dtos

// UpdateStore is the JSON body of PUT /api/v1/stores/:id. The logo is never
// set here — it's uploaded through its own endpoint
// (PUT /api/v1/stores/:id/logo), same split as a bill and its slip.
type UpdateStore struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Phone   string `json:"phone"`
}
