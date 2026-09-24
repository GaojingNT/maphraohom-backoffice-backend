package responses

import "maphraohom.app/maphraohom-backoffice/src/models"

type (
	// CustomerListItem is the shape returned by GET /customers.
	CustomerListItem struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Phone string `json:"phone"`
	}

	// CustomerDetailResponse is the shape returned by GET /customers/:id.
	CustomerDetailResponse struct {
		ID        int    `json:"id"`
		Name      string `json:"name"`
		Phone     string `json:"phone"`
		CreatedAt string `json:"createdAt"`
		UpdatedAt string `json:"updatedAt"`
	}

	// CustomerAddressItem is the shape returned by GET /customers/:id/addresses.
	CustomerAddressItem struct {
		ID        int    `json:"id"`
		Address   string `json:"address"`
		IsDefault bool   `json:"isDefault"`
		CreatedAt string `json:"createdAt"`
		UpdatedAt string `json:"updatedAt"`
	}

	// CustomerPhoneItem is the shape returned by GET /customers/:id/phones.
	CustomerPhoneItem struct {
		ID        int    `json:"id"`
		Phone     string `json:"phone"`
		IsDefault bool   `json:"isDefault"`
		CreatedAt string `json:"createdAt"`
		UpdatedAt string `json:"updatedAt"`
	}
)

func (CustomerListItem) Collection(customers []models.Customer) []CustomerListItem {
	items := make([]CustomerListItem, 0, len(customers))
	for _, customer := range customers {
		items = append(items, CustomerListItem{
			ID:    customer.ID,
			Name:  customer.Name,
			Phone: customer.Phone,
		})
	}
	return items
}

func (response *CustomerDetailResponse) Make(customer models.Customer) *CustomerDetailResponse {
	return &CustomerDetailResponse{
		ID:        customer.ID,
		Name:      customer.Name,
		Phone:     customer.Phone,
		CreatedAt: customer.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: customer.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (CustomerAddressItem) Make(address models.CustomerAddress) CustomerAddressItem {
	return CustomerAddressItem{
		ID:        address.ID,
		Address:   address.Address,
		IsDefault: address.IsDefault,
		CreatedAt: address.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: address.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (item CustomerAddressItem) Collection(addresses []models.CustomerAddress) []CustomerAddressItem {
	items := make([]CustomerAddressItem, 0, len(addresses))
	for _, address := range addresses {
		items = append(items, item.Make(address))
	}
	return items
}

func (CustomerPhoneItem) Make(phone models.CustomerPhone) CustomerPhoneItem {
	return CustomerPhoneItem{
		ID:        phone.ID,
		Phone:     phone.Phone,
		IsDefault: phone.IsDefault,
		CreatedAt: phone.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: phone.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (item CustomerPhoneItem) Collection(phones []models.CustomerPhone) []CustomerPhoneItem {
	items := make([]CustomerPhoneItem, 0, len(phones))
	for _, phone := range phones {
		items = append(items, item.Make(phone))
	}
	return items
}
