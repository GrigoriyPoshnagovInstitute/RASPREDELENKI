package transport

import (
	"context"

	"github.com/google/uuid"

	"userservice/api/server/userinternal"
	appmodel "userservice/pkg/user/application/model"
	"userservice/pkg/user/application/query"
	"userservice/pkg/user/application/service"
)

func NewUserInternalAPI(
	userQueryService query.UserQueryService,
	userService service.UserService,
) userinternal.UserInternalServiceServer {
	return &userInternalAPI{
		userQueryService: userQueryService,
		userService:      userService,
	}
}

type userInternalAPI struct {
	userQueryService query.UserQueryService
	userService      service.UserService

	userinternal.UnimplementedUserInternalServiceServer
}

func (u userInternalAPI) StoreUser(ctx context.Context, request *userinternal.StoreUserRequest) (*userinternal.StoreUserResponse, error) {
	var (
		userID uuid.UUID
		err    error
	)
	if request.User.UserID != "" {
		userID, err = uuid.Parse(request.User.UserID)
		if err != nil {
			return nil, err
		}
	}

	userID, err = u.userService.StoreUser(ctx, appmodel.User{
		UserID:   userID,
		Status:   int(request.User.Status),
		Login:    request.User.Login,
		Email:    request.User.Email,
		Telegram: request.User.Telegram,
	})
	if err != nil {
		return nil, err
	}

	return &userinternal.StoreUserResponse{
		UserID: userID.String(),
	}, nil
}

func (u userInternalAPI) FindUser(ctx context.Context, request *userinternal.FindUserRequest) (*userinternal.FindUserResponse, error) {
	userID, err := uuid.Parse(request.UserID)
	if err != nil {
		return nil, err
	}
	user, err := u.userQueryService.FindUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return &userinternal.FindUserResponse{}, nil
	}
	return &userinternal.FindUserResponse{
		User: &userinternal.User{
			UserID:   userID.String(),
			Status:   userinternal.UserStatus(user.Status), // nolint:gosec
			Login:    user.Login,
			Email:    user.Email,
			Telegram: user.Telegram,
		},
	}, nil
}
