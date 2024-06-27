package gapi

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/jackc/pgx/v5/pgtype"
	mockdb "github.com/kjasn/simple-bank/db/mock"
	db "github.com/kjasn/simple-bank/db/sqlc"
	"github.com/kjasn/simple-bank/pb"
	"github.com/kjasn/simple-bank/token"
	"github.com/kjasn/simple-bank/utils"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestUpdateUserAPI(t *testing.T) {
	user, _ := createRandomUser(t)
	newFullName := utils.RandomOwner()
	newEmail := utils.RandomEmail()
	invalidEmailAddress := "invalid-email-address"

	testCases := []struct {
		name string
		req *pb.UpdateUserRequest
		buildStubs func(store *mockdb.MockStore)
		buildContext func(t *testing.T, tokenMaker token.Maker) context.Context
		checkResponse func(t *testing.T, res *pb.UpdateUserResponse, err error)
	} {
		{
			name: "OK",
			req: &pb.UpdateUserRequest {
				Username: user.Username,
				FullName: &newFullName,
				Email: &newEmail,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdateUserParams {
					Username: user.Username,
					FullName: pgtype.Text{
						String: newFullName,
						Valid: true,
					},
					Email: pgtype.Text{
						String: newEmail,
						Valid: true,
					},
				}

				updatedUser := db.User {
					Username: user.Username,
					HashedPassword: user.HashedPassword,
					FullName: newFullName,
					Email: newEmail,
					PasswordChangedAt: user.PasswordChangedAt,
					CreatedAt: user.CreatedAt,
					IsEmailVerified: user.IsEmailVerified,
				}
				store.EXPECT().UpdateUser(gomock.Any(), gomock.Eq(arg)).
				Times(1).Return(updatedUser, nil)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return buildContextWithBearerToken(t, tokenMaker, user.Username, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.UpdateUserResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, res)
				updatedUser := res.GetUser()
				require.Equal(t, updatedUser.Username, user.Username)
				require.Equal(t, newFullName, updatedUser.FullName)
				require.Equal(t, newEmail, updatedUser.Email)
				require.WithinDuration(t, user.PasswordChangedAt, updatedUser.PasswordChangedAt.AsTime(), time.Second)
				require.WithinDuration(t, user.CreatedAt, updatedUser.CreatedAt.AsTime(), time.Second)
			},
		},
		{
			name: "InternalError",
			req: &pb.UpdateUserRequest {
				Username: user.Username,
				FullName: &newFullName,
				Email: &newEmail,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).
				Times(1).Return(db.User{}, sql.ErrConnDone)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return buildContextWithBearerToken(t, tokenMaker, user.Username, time.Minute)
			}, 
			checkResponse: func(t *testing.T, res *pb.UpdateUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Internal, st.Code())
			},
		},
		{
			name: "NotExist",
			req: &pb.UpdateUserRequest {
				Username: user.Username,
				FullName: &newFullName,
				Email: &newEmail,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).
				Times(1).Return(db.User{}, db.ErrRecordNotFound)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return buildContextWithBearerToken(t, tokenMaker, user.Username, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.UpdateUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.NotFound, st.Code())
			},
		},
		{
			name: "ExpiredToken",
			req: &pb.UpdateUserRequest {
				Username: user.Username,
				FullName: &newFullName,
				Email: &newEmail,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).
				Times(0)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return buildContextWithBearerToken(t, tokenMaker, user.Username, -time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.UpdateUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Unauthenticated, st.Code())
			},
		},
		{
			name: "NoAuthorization",
			req: &pb.UpdateUserRequest {
				Username: user.Username,
				FullName: &newFullName,
				Email: &newEmail,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).
				Times(0)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return context.Background()
			},
			checkResponse: func(t *testing.T, res *pb.UpdateUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Unauthenticated, st.Code())
			},
		},
		{
			name: "InvalidEmailAddress",
			req: &pb.UpdateUserRequest {
				Username: user.Username,
				FullName: &newFullName,
				Email: &invalidEmailAddress,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).
				Times(0)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return buildContextWithBearerToken(t, tokenMaker, user.Username, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.UpdateUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.InvalidArgument, st.Code())
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			// create a new server
			storeCtrl := gomock.NewController(t)
			defer storeCtrl.Finish()
			store := mockdb.NewMockStore(storeCtrl)

			// build stubs
			tc.buildStubs(store)

			// start test and send request
			server := newTestServer(t, store, nil)
			// build context with authorization
			ctx := tc.buildContext(t, server.tokenMaker)

			res, err := server.UpdateUser(ctx, tc.req)

			// check response
			tc.checkResponse(t, res, err)
		})
	}

}