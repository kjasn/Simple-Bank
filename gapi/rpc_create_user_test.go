package gapi

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	mockdb "github.com/kjasn/simple-bank/db/mock"
	db "github.com/kjasn/simple-bank/db/sqlc"
	"github.com/kjasn/simple-bank/pb"
	"github.com/kjasn/simple-bank/utils"
	"github.com/kjasn/simple-bank/worker"
	mockwk "github.com/kjasn/simple-bank/worker/mock"
	"github.com/stretchr/testify/require"
)

// custom matcher to check createUserParams
type eqCreateUserTxParamsMatcher struct {
	arg db.CreateUserTxParams
	password string
	user db.User
}

func (expected eqCreateUserTxParamsMatcher) Matches(x interface{}) bool {
	actualArg, ok := x.(db.CreateUserTxParams)

	if !ok {
		return false
	}

	err := utils.CheckPassword(expected.password, actualArg.HashedPassword)
	if err != nil {
		return false
	}

	expected.arg.HashedPassword = actualArg.HashedPassword
	if !reflect.DeepEqual(expected.arg.CreateUserParams, actualArg.CreateUserParams) {
		return false
	}

	// call AfterCreate
	err = actualArg.AfterCreated(expected.user)
	return err == nil
}

func (expected eqCreateUserTxParamsMatcher) String() string {
	return fmt.Sprintf("matches arg %v and password %v", expected.arg, expected.password)
}

func EqCreateUserTxParams(arg db.CreateUserTxParams, password string, user db.User) gomock.Matcher {
	return eqCreateUserTxParamsMatcher{arg, password, user}
}


func TestCreateUserAPI(t *testing.T) {
	user, password := createRandomUser(t)

	testCases := []struct {
		name string
		req *pb.CreateUserRequest 
		buildStubs func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor)
		checkResponse func(t *testing.T, res *pb.CreateUserResponse, err error)
	} {
		{
			name: "OK",
			req: &pb.CreateUserRequest {
				Username: user.Username,
				Password: password,
				FullName: user.FullName,
				Email: user.Email,
			},
			buildStubs: func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor) {
				arg := db.CreateUserTxParams{
					CreateUserParams: db.CreateUserParams {
						Username: user.Username,
						FullName: user.FullName,
						Email: user.Email,
					},

				}
				// use custom matcher to check password
				store.EXPECT().CreateUserTx(gomock.Any(), EqCreateUserTxParams(arg, password, user)).
				Times(1).Return(db.CreateUserTxResult{User: user}, nil)


				taskPayload := &worker.PayloadSendVerifyEmail {
					Username: user.Username,
				}
				taskDistributor.EXPECT().DistributeTaskSendVerifyEmail(gomock.Any(), taskPayload, gomock.Any()).Times(1).Return(nil)
			},

			checkResponse: func(t *testing.T, res *pb.CreateUserResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, res)
				// get created user
				createdUser := res.GetUser()
				require.Equal(t, createdUser.Username, user.Username)
				require.Equal(t, createdUser.FullName, user.FullName)
				require.Equal(t, createdUser.Email, user.Email)
				require.WithinDuration(t, createdUser.PasswordChangedAt.AsTime(), user.PasswordChangedAt, 
				time.Second)
				require.WithinDuration(t, createdUser.CreatedAt.AsTime(), user.CreatedAt, time.Second)
			},
		},
		// {
		// 	name: "InternalError",
		// 	req: &pb.CreateUserRequest{
		// 		Username: user.Username,
		// 		Password: password,
		// 		FullName: user.FullName,
		// 		Email: user.Email,
		// 	},
		// 	buildStubs: func(store *mockdb.MockStore) {
		// 		store.EXPECT().CreateUserTx(gomock.Any(), gomock.Any()).
		// 		Times(1).Return(db.User{}, sql.ErrConnDone)	// except no user found with internal server error
		// 	},
		// 	checkResponse: func(t *testing.T, res *pb.CreateUserResponse, err error) {
		// 		require.Equal(t, http.StatusInternalServerError, recorder.Code)
		// 	},
		// },
		// {
		// 	name: "DuplicatedUsername",
		// 	req: &pb.CreateUserRequest{
		// 		Username: user.Username,
		// 		Password: password,
		// 		FullName: user.FullName,
		// 		Email: user.Email,
		// 	},
		// 	buildStubs: func(store *mockdb.MockStore) {
		// 		store.EXPECT().CreateUserTx(gomock.Any(), gomock.Any()).Times(1).
		// 		Return(db.User{}, &pq.Error{Code: "23505"})	// except duplicated username	
		// 	},
		// 	checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) { 
		// 		require.Equal(t, http.StatusForbidden, recorder.Code)
		// 	},
		// },
		// {
		// 	name: "InvalidUsername",
		// 	req: &pb.CreateUserRequest{
		// 		Username: "invalid-username",	// username can not include '-'
		// 		Password: password,
		// 		FullName: user.FullName,
		// 		Email: user.Email,
		// 	},
		// 	buildStubs: func(store *mockdb.MockStore) {
		// 		store.EXPECT().CreateUserTx(gomock.Any(), gomock.Any()).Times(0)
		// 	},
		// 	checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) { 
		// 		require.Equal(t, http.StatusBadRequest, recorder.Code)
		// 	},
		// },
		// {
		// 	name: "InvalidEmail",
		// 	req: &pb.CreateUserRequest{
		// 		Username: user.Username,
		// 		Password: password,
		// 		FullName: user.FullName,
		// 		Email: "invalid-Email",
		// 	},
		// 	buildStubs: func(store *mockdb.MockStore) {
		// 		store.EXPECT().CreateUserTx(gomock.Any(), gomock.Any()).Times(0)
		// 	},
		// 	checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) { 
		// 		require.Equal(t, http.StatusBadRequest, recorder.Code)
		// 	},
		// },
		// {
		// 	name: "TooShortPassword",
		// 	req: &pb.CreateUserRequest{
		// 		Username: user.Username,
		// 		Password: "123",	// at least 6 character
		// 		FullName: user.FullName,
		// 		Email: user.Email,
		// 	},
		// 	buildStubs: func(store *mockdb.MockStore) {
		// 		store.EXPECT().
		// 			CreateUserTx(gomock.Any(), gomock.Any()).
		// 			Times(0)
		// 	},
		// 	checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
		// 		require.Equal(t, http.StatusBadRequest, recorder.Code)
		// 	},
		// },
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			// create a new server
			storeCtrl := gomock.NewController(t)
			defer storeCtrl.Finish()
			store := mockdb.NewMockStore(storeCtrl)

			// create task distributor
			taskCtrl := gomock.NewController(t)
			defer taskCtrl.Finish()
			taskDistributor := mockwk.NewMockTaskDistributor(taskCtrl)

			// build stubs
			tc.buildStubs(store, taskDistributor)

			// start test and send request
			server := newTestServer(t, store, taskDistributor)
			res, err := server.CreateUser(context.Background(), tc.req)

			// check response
			tc.checkResponse(t, res, err)
		})
	}
	
}


func createRandomUser(t *testing.T) (db.User, string) {
	password := utils.RandomString(6)	// password before hashed
	hashedPassword, err := utils.HashPassword(password)
	require.NoError(t, err)

	user := db.User{
		Username: utils.RandomOwner(),
		HashedPassword: hashedPassword,
		FullName: utils.RandomOwner(),
		Email: utils.RandomEmail(),
	}

	return user, password
}