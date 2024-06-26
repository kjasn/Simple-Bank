package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	mockdb "github.com/kjasn/simple-bank/db/mock"
	db "github.com/kjasn/simple-bank/db/sqlc"
	"github.com/kjasn/simple-bank/token"
	"github.com/stretchr/testify/require"
)

func TestCreateTransfer(t *testing.T) {
	amount := int64(10)

	user1, _ := createRandomUser(t)
	user2, _ := createRandomUser(t)
	user3, _ := createRandomUser(t)

	account1 := createRandomAccount(user1.Username)
	account2 := createRandomAccount(user2.Username)
	account3 := createRandomAccount(user3.Username)

	account1.Currency = "USD"
	account2.Currency = "USD"
	account3.Currency = "EUR"

	testCases := []struct {
		name string
		body gin.H
		buildStubs func(store *mockdb.MockStore)
		setupAuth func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		checkResponse func(recorder *httptest.ResponseRecorder)
	} {
		{
			name: "OK",
			body: gin.H{
				"from_account_id": account1.ID,
				"to_account_id": account2.ID,
				"amount": amount,
				"currency": account1.Currency,
			},
			buildStubs: func(store *mockdb.MockStore) {
				// get 2 accounts
				store.EXPECT().
				GetAccount(gomock.Any(), gomock.Eq(account1.ID)).
				Times(1).Return(account1, nil)

				store.EXPECT().
				GetAccount(gomock.Any(), gomock.Eq(account2.ID)).
				Times(1).Return(account2, nil)

				arg := db.TransferTxParams{
					FromAccountID: account1.ID,
					ToAccountID: account2.ID,
					Amount: amount,
				}

				store.EXPECT().TransferTx(gomock.Any(), gomock.Eq(arg)).
				Times(1)
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, supportedAuthorizationType, user1.Username, user1.Role, time.Minute)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
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
			server := newTestServer(t, store)
			recorder := httptest.NewRecorder() // create a recorder serve as ResponseWriter

			url := "/transfers"
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			// check response
			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)	// send request and record the response in recorder
			tc.checkResponse(recorder)
		})
	}

}
