package api

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	db "github.com/kjasn/simple-bank/db/sqlc"
	"github.com/kjasn/simple-bank/utils"
)

type createUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Role 	 string	`json:"role"`
	Password string `json:"password" binding:"required,min=6"`
	FullName string `json:"full_name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
}

// user data to object, without hashed password
type userDTO struct {
	Username          string    `json:"username"`
	Role 			  string	`json:"role"`
	FullName          string    `json:"full_name"`
	Email             string    `json:"email"`
	PasswordChangedAt time.Time `json:"password_changed_at"`
	CreatedAt         time.Time `json:"created_at"`
}

func toUserDTO(u *db.User) userDTO{
	return userDTO {
		Username: u.Username,
		Role: string(u.Role),
		FullName: u.FullName,
		Email: u.Email,
		PasswordChangedAt: u.PasswordChangedAt,
		CreatedAt: u.CreatedAt,
	}	
}

// createUser
func (server *Server) createUser(ctx *gin.Context) {
	var req createUserRequest
	// get request from context
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	arg := db.CreateUserParams {
		Username: req.Username,
		HashedPassword: hashedPassword,
		FullName: req.FullName,
		Email: req.Email,
	}

	user, err := server.store.CreateUser(ctx, arg)
	if err != nil {
		// if pqErr, ok := err.(*pq.Error); ok {
		if db.ErrorCode(err) == db.UniqueViolation {
			ctx.JSON(http.StatusForbidden, errorResponse(err))	
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, toUserDTO(&user))
}


type getUserRequest struct {
	Username string`uri:"username" binding:"required"`
}

// getUser
func (server *Server) getUser(ctx *gin.Context) {
	var req getUserRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	user, err := server.store.GetUser(ctx, req.Username)
	log.Println("==================")
	log.Println(req.Username)
	log.Println("==================")
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, toUserDTO(&user))
}



type loginUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=6"`
}

type loginUserResponse struct {
	SessionID 		uuid.UUID				`json:"session_id"`
	AccessToken 	string					`json:"access_token"`
	AccessTokenExpiresAt time.Time 			`json:"access_token_expires_at"`
	RefreshToken 	string 					`json:"refresh_token"`
	RefreshTokenTokenExpiresAt time.Time 	`json:"refresh_token_expires_at"`
	User 			userDTO					`json:"user"`
}


func (server *Server) loginUser(ctx *gin.Context) {
	var req loginUserRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	
	user, err := server.store.GetUser(ctx, req.Username)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	err = utils.CheckPassword(req.Password, user.HashedPassword)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}
	
	// create access token 
	accessToken, accessTokenPayload, err := server.tokenMaker.CreateToken(
		user.Username,
		user.Role,
		server.config.AccessTokenDuration,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// create refresh token after access token
	refreshToken, refreshTokenPayload, err := server.tokenMaker.CreateToken(
		user.Username,
		user.Role,
		server.config.RefreshTokenDuration,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}


	// store the refresh token into session table
	session, err := server.store.CreateSession(ctx, db.CreateSessionParams{
		ID: refreshTokenPayload.ID,
		Username: user.Username,
		RefreshToken : refreshToken,
		UserAgent : ctx.Request.UserAgent(),
		ClientIp : ctx.ClientIP(),
		IsBlocked : false, 
		ExpiresAt: refreshTokenPayload.ExpiredAt,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}


	ctx.JSON(http.StatusOK, &loginUserResponse {
		SessionID: session.ID,
		AccessToken: accessToken,
		AccessTokenExpiresAt: accessTokenPayload.ExpiredAt,
		RefreshToken: refreshToken,
		RefreshTokenTokenExpiresAt: refreshTokenPayload.ExpiredAt,
		User: toUserDTO(&user),
	})
}