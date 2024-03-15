package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/kjasn/simple-bank/db/sqlc"
)

type createAccountRequest struct {
	Owner    string `json:"owner" binding:"require"`
	Currency string `json:"currency" binding:"require,oneof=USD EUR RMB"`
}


func (server *Server) createAccount(ctx *gin.Context) {
	var req createAccountRequest
	// get request from context
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.CreateAccountParams {
		Owner: req.Owner,
		Balance: 0,
		Currency: req.Currency,
	}

	account, err := server.store.CreateAccount(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, account)
}