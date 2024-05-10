package api

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)


type renewAccessTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type renewAccessTokenResponse struct {
	AccessToken 	string					`json:"access_token"`
	AccessTokenExpiresAt time.Time 			`json:"access_token_expires_at"`
}


func (server *Server) renewAccessToken(ctx *gin.Context) {
	var req renewAccessTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// verify refresh token
	refreshPayload, err := server.tokenMaker.VerifyToken(req.RefreshToken)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}
	
	session, err := server.store.GetSession(ctx, refreshPayload.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	if session.IsBlocked {
		err = fmt.Errorf(" blocked session")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
	}

	if session.Username != refreshPayload.Username {
		err = fmt.Errorf("incorrect session user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
	}

	if session.RefreshToken != req.RefreshToken {
		err = fmt.Errorf("mismatched refresh token")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
	}

	if time.Now().After(session.ExpiresAt) {
		err = fmt.Errorf("expired session")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
	}

	// create access token 
	accessToken, accessTokenPayload, err := server.tokenMaker.CreateToken(
		session.Username,	// OR refreshPayload.Username
		server.config.AccessTokenDuration,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}


	ctx.JSON(http.StatusOK, &renewAccessTokenResponse {
		AccessToken: accessToken,
		AccessTokenExpiresAt: accessTokenPayload.ExpiredAt,
	})
}