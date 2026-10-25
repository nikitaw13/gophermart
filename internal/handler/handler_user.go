package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/nikitaw13/gophermart/internal/model"
	"go.uber.org/zap"
)

// registerUser handles POST /api/user/register.
func (h *APIHandler) registerUser(w http.ResponseWriter, req *http.Request) {
	user, ok := decodeJSON[model.User](w, req, h.logger)
	if !ok {
		return
	}

	err := user.Validate()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(req.Context(), queryTimeout)
	defer cancel()

	token, err := h.svc.RegisterUser(ctx, user)
	if errors.Is(err, model.ErrUserAlreadyExists) {
		h.logger.Debug("user already exists", zap.String("username", user.Username))
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}
	// Defensive: user.Validate() already rejects over-long passwords before bcrypt runs.
	if errors.Is(err, model.ErrPasswordTooLong) {
		h.logger.Debug(err.Error(), zap.String("username", user.Username))
		http.Error(w, "Password is too long", http.StatusBadRequest)
		return
	}

	if err != nil {
		h.logger.Error("failed to register user", zap.Error(err))
		http.Error(w, "The server failed to fulfill an apparently valid request", http.StatusInternalServerError)
		return
	}

	if token == "" {
		http.Error(w, "The server failed to fulfill an apparently valid request", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", bearerPrefix+token)
	w.WriteHeader(http.StatusOK)
}

// loginUser handles POST /api/user/login.
func (h *APIHandler) loginUser(w http.ResponseWriter, req *http.Request) {
	user, ok := decodeJSON[model.User](w, req, h.logger)
	if !ok {
		return
	}

	err := user.Validate()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(req.Context(), queryTimeout)
	defer cancel()

	token, err := h.svc.LoginUser(ctx, user)
	// All credential failures map to the same 401, so the response does not reveal which part failed.
	if errors.Is(err, model.ErrUserNotFound) || errors.Is(err, model.ErrPasswordTooLong) || errors.Is(err, model.ErrMismatchedPassword) {
		h.logger.Debug(err.Error(), zap.String("username", user.Username))
		http.Error(w, "Invalid login or password", http.StatusUnauthorized)
		return
	}

	if err != nil {
		h.logger.Error("failed to login user", zap.Error(err))
		http.Error(w, "The server failed to fulfill an apparently valid request", http.StatusInternalServerError)
		return
	}

	if token == "" {
		http.Error(w, "The server failed to fulfill an apparently valid request", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", bearerPrefix+token)
	w.WriteHeader(http.StatusOK)
}
