package responses

import "maphraohom.app/maphraohom-backoffice/src/models"

type (
	GetUsersCollection struct {
		UserResponse
	}
)

func (response *GetUsersCollection) Collection(users []models.User) []GetUsersCollection {
	var userCollection []GetUsersCollection

	for _, user := range users {
		userResponse := GetUsersCollection{
			UserResponse: UserResponse{
				ID:        user.ID,
				Email:     user.Email,
				FirstName: user.FirstName,
				LastName:  user.LastName,
				RoleID:    user.RoleID,
				RoleName: func() string {
					if user.Role != nil {
						return user.Role.Name
					}
					return ""
				}(),
				Permissions: func() []string {
					permissions := make([]string, 0)
					if user.Role == nil {
						return permissions
					}
					for _, permission := range user.Role.Permissions {
						permissions = append(permissions, permission.Name)
					}
					return permissions
				}(),
				CreatedAt: user.CreatedAt.Format("2006-01-02 15:04:05"),
				UpdatedAt: user.UpdatedAt.Format("2006-01-02 15:04:05"),
			},
		}
		userCollection = append(userCollection, userResponse)
	}

	return userCollection
}
