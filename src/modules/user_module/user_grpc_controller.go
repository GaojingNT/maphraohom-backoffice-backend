package user_module

import (
	"io"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"maphraohom.app/maphraohom-backoffice/internal/exception"
	"maphraohom.app/maphraohom-backoffice/pkg/cache"
	pbUser "maphraohom.app/maphraohom-backoffice/src/modules/user_module/proto/user"
)

func (c GrpcController) GetUsers(stream pbUser.UserService_GetUsersServer) error {
	var (
		ctx, span = c.m.tracer.TraceStart(stream.Context(), "GetUsersController", trace.WithAttributes(attribute.String("server", "grpc"), attribute.String("controller", "GetUsers")))
		err       error
	)

	// Create response body
	users := &pbUser.Users{
		Data: make([]*pbUser.User, 0),
	}

	// Looping
	for {
		var requestUser *pbUser.UserRequest
		requestUser, err = stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Make cache indexing
		cacheTags, cacheKey := c.m.cacher.Indexing([]string{"users"}, "GetUser", requestUser.Id)

		// Get data from cache
		result, err := cache.QueryByParamCache(c.m.cacher, ctx, cacheKey, cacheTags, int(requestUser.Id), c.userService().GetUser)
		if err != nil {
			return exception.GrpcErrorResponseMapping(ctx, err)
		}

		user := result.User
		createdAt, _ := time.Parse(time.RFC3339, user.CreatedAt)
		updatedAt, _ := time.Parse(time.RFC3339, user.UpdatedAt)

		// Appending
		users.Data = append(users.Data, &pbUser.User{
			Id:        int64(user.ID),
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Email:     user.Email,
			CreatedAt: createdAt.Unix(),
			UpdatedAt: updatedAt.Unix(),
		})
	}

	// Response
	stream.SendAndClose(users)

	c.m.tracer.TraceEnd(span)

	return nil
}
