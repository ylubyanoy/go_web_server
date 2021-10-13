package handlers

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/ylubyanoy/go_web_server/internal/data"
	"github.com/ylubyanoy/go_web_server/internal/services"
	"github.com/ylubyanoy/go_web_server/internal/utils"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

var UserCreationFailed = "Unable to create user.Please try again later"
var ErrUserAlreadyExists = "User already exists with the given email"
var ErrUserNotFound = "No user account exists with given email. Please sign in first"

var PgDuplicateKeyMsg = "duplicate key value violates unique constraint"
var PgNoRowsMsg = "no rows in result set"

// Below data types are used for encoding and decoding b/t go types and json
type TokenResponse struct {
	RefreshToken string `json:"refresh_token"`
	AccessToken  string `json:"access_token"`
}

type AuthResponse struct {
	RefreshToken string `json:"refresh_token"`
	AccessToken  string `json:"access_token"`
	Username     string `json:"username"`
}

// GenericResponse is the format of our response
type GenericResponse struct {
	Status  bool        `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// UserKey is used as a key for storing the User object in context at middleware
type UserKey struct{}

// UserIDKey is used as a key for storing the UserID in context at middleware
type UserIDKey struct{}

// VerificationDataKey is used as the key for storing the VerificationData in context at middleware
type VerificationDataKey struct{}

// UserHandler wraps instances needed to perform operations on user object
type AuthHandler struct {
	logger *zap.SugaredLogger
	// configs     *utils.Configurations
	validator   *data.Validation
	repo        data.Repository
	authService services.Authentication
	// mailService service.MailService
}

// NewUserHandler returns a new UserHandler instance
func NewAuthHandler(l *zap.SugaredLogger, v *data.Validation, r data.Repository, auth services.Authentication) *AuthHandler {
	return &AuthHandler{
		logger:      l,
		validator:   v,
		repo:        r,
		authService: auth,
	}
}

func (ah *AuthHandler) hashPassword(password string) (string, error) {

	hashedPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		ah.logger.Error("unable to hash password", zap.Error(err))
		return "", err
	}

	return string(hashedPass), nil
}

// Signup handles signup request
func (ah *AuthHandler) Signup(w http.ResponseWriter, r *http.Request) {
	ah.logger.Info("Received a call Signup")
	w.Header().Set("Content-Type", "application/json")

	user := r.Context().Value(UserKey{}).(data.User)

	hashedPass, err := ah.hashPassword(user.Password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		data.ToJSON(&GenericResponse{Status: false, Message: UserCreationFailed}, w)
		return
	}
	user.Password = hashedPass
	user.TokenHash = utils.GenerateRandomString(15)

	err = ah.repo.Create(context.Background(), &user)
	if err != nil {
		ah.logger.Error("unable to insert user to database error", zap.Error(err))
		errMsg := err.Error()
		if strings.Contains(errMsg, PgDuplicateKeyMsg) {
			w.WriteHeader(http.StatusBadRequest)
			data.ToJSON(&GenericResponse{Status: false, Message: ErrUserAlreadyExists}, w)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			data.ToJSON(&GenericResponse{Status: false, Message: UserCreationFailed}, w)
		}
		return
	}

	// Create verification
	mailData := &services.MailData{
		Username: user.Username,
		Code:     utils.GenerateRandomString(8),
	}

	verificationData := &data.VerificationData{
		Email:     user.Email,
		Code:      mailData.Code,
		Type:      data.MailConfirmation,
		ExpiresAt: time.Now().Add(time.Minute * time.Duration(30)),
	}

	err = ah.repo.StoreVerificationData(context.Background(), verificationData)
	if err != nil {
		ah.logger.Error("unable to store mail verification data error", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		data.ToJSON(&GenericResponse{Status: false, Message: UserCreationFailed}, w)
		return
	}

	ah.logger.Debug("User created successfully")
	w.WriteHeader(http.StatusCreated)
	msg := "Please verify your account using the confirmation code"
	data.ToJSON(&GenericResponse{Status: true, Message: msg}, w)
}

// Login handles login request
func (ah *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	reqUser := r.Context().Value(UserKey{}).(data.User)

	user, err := ah.repo.GetUserByEmail(context.Background(), reqUser.Email)
	if err != nil {
		ah.logger.Error("error fetching the user", zap.Error(err))
		errMsg := err.Error()
		if strings.Contains(errMsg, PgNoRowsMsg) {
			w.WriteHeader(http.StatusBadRequest)
			data.ToJSON(&GenericResponse{Status: false, Message: ErrUserNotFound}, w)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			data.ToJSON(&GenericResponse{Status: false, Message: "Unable to retrieve user from database. Please try again later"}, w)
		}
		return
	}

	if !user.IsVerified {
		ah.logger.Error("unverified user")
		w.WriteHeader(http.StatusUnauthorized)
		data.ToJSON(&GenericResponse{Status: false, Message: "Please verify your mail address before login"}, w)
		return
	}

	if valid := ah.authService.Authenticate(&reqUser, user); !valid {
		ah.logger.Debug("Authetication of user failed")
		w.WriteHeader(http.StatusBadRequest)
		data.ToJSON(&GenericResponse{Status: false, Message: "Incorrect password"}, w)
		return
	}

	accessToken, err := ah.authService.GenerateAccessToken(user)
	if err != nil {
		ah.logger.Error("unable to generate access token", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		data.ToJSON(&GenericResponse{Status: false, Message: "Unable to login the user. Please try again later"}, w)
		return
	}
	refreshToken, err := ah.authService.GenerateRefreshToken(user)
	if err != nil {
		ah.logger.Error("unable to generate refresh token", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		data.ToJSON(&GenericResponse{Status: false, Message: "Unable to login the user. Please try again later"}, w)
		return
	}

	ah.logger.Debug("successfully generated token", "accesstoken", accessToken, "refreshtoken", refreshToken)
	w.WriteHeader(http.StatusOK)
	data.ToJSON(&GenericResponse{
		Status:  true,
		Message: "Successfully logged in",
		Data:    &AuthResponse{AccessToken: accessToken, RefreshToken: refreshToken, Username: user.Username},
	}, w)
}

// VerifyMail verifies the provided confirmation code and set the User state to verified
func (ah *AuthHandler) VerifyMail(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	ah.logger.Debug("verifying the confimation code")
	verificationData := r.Context().Value(VerificationDataKey{}).(data.VerificationData)
	verificationData.Type = data.MailConfirmation

	actualVerificationData, err := ah.repo.GetVerificationData(context.Background(), verificationData.Email, verificationData.Type)
	if err != nil {
		ah.logger.Error("unable to fetch verification data", zap.Error(err))

		if strings.Contains(err.Error(), PgNoRowsMsg) {
			w.WriteHeader(http.StatusNotAcceptable)
			data.ToJSON(&GenericResponse{Status: false, Message: ErrUserNotFound}, w)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		data.ToJSON(&GenericResponse{Status: false, Message: "Unable to verify mail. Please try again later"}, w)
		return
	}

	valid, err := ah.verify(actualVerificationData, &verificationData)
	if !valid {
		w.WriteHeader(http.StatusNotAcceptable)
		data.ToJSON(&GenericResponse{Status: false, Message: err.Error()}, w)
		return
	}

	// correct code, update user status to verified.
	err = ah.repo.UpdateUserVerificationStatus(context.Background(), verificationData.Email, true)
	if err != nil {
		ah.logger.Error("unable to set user verification status to true")
		w.WriteHeader(http.StatusInternalServerError)
		data.ToJSON(&GenericResponse{Status: false, Message: "Unable to verify mail. Please try again later"}, w)
		return
	}

	// delete the VerificationData from db
	err = ah.repo.DeleteVerificationData(context.Background(), verificationData.Email, verificationData.Type)
	if err != nil {
		ah.logger.Error("unable to delete the verification data", zap.Error(err))
	}

	ah.logger.Debug("user mail verification succeeded")

	w.WriteHeader(http.StatusAccepted)
	data.ToJSON(&GenericResponse{Status: true, Message: "Mail Verification succeeded"}, w)
}

func (ah *AuthHandler) verify(actualVerificationData *data.VerificationData, verificationData *data.VerificationData) (bool, error) {

	// check for expiration
	if actualVerificationData.ExpiresAt.Before(time.Now()) {
		ah.logger.Error("verification data provided is expired")
		err := ah.repo.DeleteVerificationData(context.Background(), actualVerificationData.Email, actualVerificationData.Type)
		ah.logger.Error("unable to delete verification data from db", zap.Error(err))
		return false, errors.New("confirmation code has expired. Please try generating a new code")
	}

	if actualVerificationData.Code != verificationData.Code {
		ah.logger.Error("verification of mail failed. Invalid verification code provided")
		return false, errors.New("verification code provided is Invalid. Please look in your mail for the code")
	}

	return true, nil
}

// RefreshToken handles refresh token request
func (ah *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	user := r.Context().Value(UserKey{}).(data.User)
	accessToken, err := ah.authService.GenerateAccessToken(&user)
	if err != nil {
		ah.logger.Error("unable to generate access token", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		data.ToJSON(&GenericResponse{Status: false, Message: "Unable to generate access token.Please try again later"}, w)
		return
	}

	w.WriteHeader(http.StatusOK)
	data.ToJSON(&GenericResponse{
		Status:  true,
		Message: "Successfully generated new access token",
		Data:    &TokenResponse{AccessToken: accessToken},
	}, w)
}
