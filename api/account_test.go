package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	mockdb "github.com/kjasn/simple-bank/db/mock"
	db "github.com/kjasn/simple-bank/db/sqlc"
	"github.com/kjasn/simple-bank/utils"
	"github.com/stretchr/testify/require"
)


func TestGetAccountAPI(t *testing.T) {
	user, _ := createRandomUser(t)
	account := createRandomAccount(user.Username)

	testCases := []struct {
		name string
		accountID int64
		buildStubs func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	} {
		{
			name: "OK",
			accountID: account.ID,
			buildStubs: func(store *mockdb.MockStore) {
				// run with any context and a ID equals the created accounts' and call the function 1 time 
				store.EXPECT().
				GetAccount(gomock.Any(), gomock.Eq(account.ID)).
				Times(1).
				Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) { 
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body, account)	
			},
		},
		{
			name: "NotFound",
			accountID: account.ID,	// use the same ID for simplicity
			buildStubs: func(store *mockdb.MockStore) {
				// run with any context and a ID equals the created accounts' and call the function 1 time 
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).
				Times(1).Return(db.Account{}, sql.ErrNoRows)	// except no account found with no rows error
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) { 
				require.Equal(t, http.StatusNotFound, recorder.Code)
				// no account found, do not check body
			},
		},
		{
			name: "InternalError",
			accountID: account.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).
				Times(1).Return(db.Account{}, sql.ErrConnDone)	// except no account found with internal server error
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) { 
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "InvalidID",
			accountID: 0,	// set an invalid ID to mock bad request
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) { 
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			// create a new server
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			// build stubs
			tc.buildStubs(store)

			// start test and send request
			server := NewServer(store)
			recorder := httptest.NewRecorder() // create a recorder serve as ResponseWriter

			url := fmt.Sprintf("/accounts/%d", tc.accountID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			// check response
			server.router.ServeHTTP(recorder, request)	// send request and record the response in recorder
			tc.checkResponse(t, recorder)
		})
	}
	
}

func TestCreateAccountAPI(t *testing.T) {
	user, _ := createRandomUser(t)
	account := createRandomAccount(user.Username)

	testCases := []struct {
		name 	string
		body 	gin.H
		buildStubs func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	} {
		{
			name : "OK",
			body: gin.H{
				"owner": account.Owner,
				"currency": account.Currency,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateAccountParams{
					Owner: account.Owner,
					Currency: account.Currency,
					Balance: 0,
				}
				store.EXPECT().CreateAccount(gomock.Any(), gomock.Eq(arg)).
				Times(1).
				Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body, account)
			},
		},
		{
			name : "InternalError",
			body : gin.H{
				"owner": account.Owner,
				"currency": account.Currency,
			},

			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().CreateAccount(gomock.Any(), gomock.Any()).
				Times(1).
				Return(db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name : "InvalidCurrency",
			body : gin.H{
				"owner": account.Owner,
				"currency": "invalid",
			},

			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().CreateAccount(gomock.Any(), gomock.Any()).
				Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name : "InvalidOwner",
			body : gin.H{
				"owner": "",
				"currency": account.Currency,
			},

			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().CreateAccount(gomock.Any(), gomock.Any()).
				Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},

	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			// create a new server
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			// build stubs
			tc.buildStubs(store)

			// start test and send request
			server := NewServer(store)
			recorder := httptest.NewRecorder() // create a recorder serve as ResponseWriter

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/accounts"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			// check response
			server.router.ServeHTTP(recorder, request)	// send request and record the response in recorder
			tc.checkResponse(t, recorder)
		})
	}
} 

func createRandomAccount(owner string) db.Account {
	// create a random account
	return db.Account{
		ID: utils.RandomInt(1, 1000),
		Owner: owner,
		Balance: utils.RandomMoney(),
		Currency: utils.RandomCurrency(),
	}
}

func requireBodyMatchAccount(t *testing.T, body *bytes.Buffer, account db.Account) {
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var gotAccount db.Account
	err = json.Unmarshal(data, &gotAccount)
	require.NoError(t, err)
	require.Equal(t, account, gotAccount)
}