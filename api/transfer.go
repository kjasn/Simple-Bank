package api

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/kjasn/simple-bank/db/sqlc"
	"github.com/kjasn/simple-bank/token"
)


type transferRequest struct {
	FromAccountID int64 `json:"from_account_id" binding:"required,min=1"`
	ToAccountID   int64 `json:"to_account_id" binding:"required,min=1"`
	Amount        int64 `json:"amount" binding:"required,gt=0"`
	Currency string `json:"currency" binding:"required,currency"`
}


func (server *Server) createTransfer(ctx *gin.Context) {
	var req transferRequest
	// get request from context
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	fromAccount, ok := server.validAccount(ctx, req.FromAccountID, req.Currency)
	if !ok {
		// no response here, response in validAccount function
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if authPayload.Username != fromAccount.Owner {
		err := errors.New("From account doesn't belong to the authenticate user")
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	
	if _, ok := server.validAccount(ctx, req.ToAccountID, req.Currency); !ok {
		return
	}

	arg := db.TransferTxParams{
		FromAccountID: req.FromAccountID,
		ToAccountID: req.ToAccountID,
		Amount: req.Amount,
	}


	result, err := server.store.TransferTx(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, result)
}


// check if the specific accounts with the passed ID exists and their currency matches
func (server *Server) validAccount(ctx *gin.Context, id int64, currency string) (db.Account, bool) {
	account, err := server.store.GetAccount(ctx, id)

	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return account, false
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return account, false
	}

	if account.Currency != currency {
		err = fmt.Errorf("account [%d] currency mismatch: %s vs %s", 
		id, account.Currency, currency)

		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return account, false
	}

	return account, true

}