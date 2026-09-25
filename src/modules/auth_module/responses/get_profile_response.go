package responses

import (
	"fmt"
	"time"

	"maphraohom.app/maphraohom-backoffice/src/models"
)

type (
	GetProfileResponse struct {
		ID        int    `json:"id"`
		Email     string `json:"email"`
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`

		// Ready-to-use URL of the user's signature image — printed on every
		// receipt they export. Empty string when unset.
		Signature string `json:"signature"`

		// Relations
		Role   *GetProfileRoleResponse   `json:"role"`
		Stores []GetProfileStoreResponse `json:"stores"`

		// Timestamp & Audit fields
		CreatedAt time.Time `json:"createdAt"`
		UpdatedAt time.Time `json:"updatedAt"`
	}
	// Role and permissions
	GetProfileRoleResponse struct {
		ID          int    `json:"id"`
		Name        string `json:"name"`
		Label       string `json:"label"`
		Description string `json:"description"`
		// Relations
		RoleGroup   *GetProfileRoleGroupRoleResponse   `json:"roleGroup"`
		Permissions []GetProfileRolePermissionResponse `json:"permissions"`
		Menus       []GetProfileRoleMenuResponse       `json:"menus"`
	}
	GetProfileRoleGroupRoleResponse struct {
		ID          int    `json:"id"`
		Name        string `json:"name"`
		Code        string `json:"code"`
		Description string `json:"description"`
	}
	GetProfileRoleMenuResponse struct {
		ID          int    `json:"id"`
		Name        string `json:"name"`
		KeyName     string `json:"keyName"`
		Label       string `json:"label"`
		Icon        string `json:"icon"`
		UrlPath     string `json:"urlPath"`
		Description string `json:"description"`
		IsFeature   bool   `json:"isFeature"`
		IsActive    bool   `json:"isActive"`
		Mode        string `json:"mode"`
		// Relations
		SubMenus    []GetProfileRoleMenuResponse       `json:"subMenus,omitempty"`
		Permissions []GetProfileRolePermissionResponse `json:"permissions,omitempty"`
	}
	GetProfileRolePermissionResponse struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	// Stores the user owns (tbl_user_stores)
	GetProfileStoreResponse struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
		Logo string `json:"logo"`
	}
)

func (response *GetProfileResponse) Make(user *models.User) *GetProfileResponse {
	// Make response objects
	roleGroup := &GetProfileRoleGroupRoleResponse{}
	roleMenus := make([]GetProfileRoleMenuResponse, 0)
	rolePermissions := make([]GetProfileRolePermissionResponse, 0)
	responseRole := &GetProfileRoleResponse{}

	// Check if user has role
	if user.Role != nil {
		if user.Role.RoleGroups != nil {
			if len(user.Role.RoleGroups) > 0 {
				roleGroup = &GetProfileRoleGroupRoleResponse{
					ID:          user.Role.RoleGroups[0].ID,
					Name:        user.Role.RoleGroups[0].Name,
					Code:        user.Role.RoleGroups[0].Code,
					Description: user.Role.RoleGroups[0].Description,
				}
			}
		}
		if user.Role.Menus != nil {
			for _, menu := range user.Role.Menus {
				subMenus := make([]GetProfileRoleMenuResponse, 0)
				if menu.SubMenus != nil {
					for _, subMenu := range menu.SubMenus {
						roleSubMenu := GetProfileRoleMenuResponse{
							ID:          subMenu.ID,
							Name:        subMenu.Name,
							KeyName:     subMenu.KeyName,
							Label:       subMenu.Label,
							Icon:        subMenu.Icon,
							UrlPath:     subMenu.UrlPath,
							Description: subMenu.Description,
							IsFeature:   subMenu.IsFeature,
							IsActive:    subMenu.IsActive,
							Mode:        subMenu.Mode,
						}

						subMenus = append(subMenus, roleSubMenu)
					}
				}

				roleMenu := GetProfileRoleMenuResponse{
					ID:          menu.ID,
					Name:        menu.Name,
					KeyName:     menu.KeyName,
					Label:       menu.Label,
					Icon:        menu.Icon,
					UrlPath:     menu.UrlPath,
					Description: menu.Description,
					IsFeature:   menu.IsFeature,
					IsActive:    menu.IsActive,
					Mode:        menu.Mode,
					// Relations
					SubMenus: subMenus,
				}

				roleMenus = append(roleMenus, roleMenu)
			}
		}
		if user.Role.Permissions != nil {
			for _, permission := range user.Role.Permissions {
				rolePermission := GetProfileRolePermissionResponse{
					ID:   permission.ID,
					Name: permission.Name,
				}

				rolePermissions = append(rolePermissions, rolePermission)
			}
		}

		// Mapping content role response
		responseRole = &GetProfileRoleResponse{
			ID:          user.Role.ID,
			Name:        user.Role.Name,
			Label:       user.Role.Label,
			Description: user.Role.Description,
			// Relations
			RoleGroup:   roleGroup,
			Permissions: rolePermissions,
			Menus:       roleMenus,
		}
	} else {
		responseRole = nil
	}

	stores := make([]GetProfileStoreResponse, 0, len(user.Stores))
	for _, store := range user.Stores {
		stores = append(stores, GetProfileStoreResponse{
			ID:   store.ID,
			Name: store.Name,
			Logo: resolveKey(store.Logo),
		})
	}

	response = &GetProfileResponse{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		Signature: resolveKey(user.Signature),
		// Relations
		Role:   responseRole,
		Stores: stores,
		// Timestamp & Audit fields
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	return response
}

// FileURLBuilder turns a stored object key (signature, store logo) into a URL
// the client can load directly — the same GET /api/v1/files/* proxy the
// store and bill modules use.
var FileURLBuilder = func(key string) string {
	return fmt.Sprintf("/api/v1/files/%s", key)
}

// resolveKey turns a stored object key into a ready-to-use URL, leaving an
// unset key as "" rather than a broken "/api/v1/files/" link.
func resolveKey(key string) string {
	if key == "" {
		return ""
	}
	return FileURLBuilder(key)
}
