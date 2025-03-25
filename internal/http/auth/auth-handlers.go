package authhttp

import (
	"encoding/json"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log/slog"
	"net/http"
	"net/mail"
	"sso/internal/lib/jwt"
	"sso/internal/storage"
	"sso/internal/storage/postgres"
	"time"
)

type Handler struct {
	storage  *postgres.Storage
	log      *slog.Logger
	tokenTTL time.Duration
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type RegisterRequest struct {
	Name     string
	Email    string
	Password string
}

func NewHandler(storage *postgres.Storage, log *slog.Logger, ttl time.Duration) *Handler {
	return &Handler{storage: storage, log: log, tokenTTL: ttl}
}

func (h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var logreq LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&logreq); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	h.log.Info("request body decoded",
		slog.String("email", logreq.Email),
		slog.String("password", logreq.Password),
	)

	user, err := h.storage.User(ctx, logreq.Email)

	if err != nil {
		h.log.Warn("user not found", err.Error())
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(logreq.Password)); err != nil {
		h.log.Info("invalid credentials", err.Error())
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	h.log.Info("user successfully logged in")

	app, errr := h.storage.App(ctx, 1)
	if errr != nil {
		http.Error(w, errr.Error(), http.StatusInternalServerError)
		return
	}

	token, err := jwt.NewToken(user, app, h.tokenTTL)
	if err != nil {
		h.log.Error("failed to create token", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(token)
	if err != nil {
		h.log.Error("failed to encode token", err.Error())
	}
}

func (h *Handler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var regReq RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&regReq); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.log.Info("registering new user")

	h.log.Info("request body decoded",
		slog.String("email", regReq.Email),
		slog.String("name", regReq.Name),
	)

	passHash, err := bcrypt.GenerateFromPassword([]byte(regReq.Password), bcrypt.DefaultCost)
	if err != nil {
		h.log.Error("failed to hash password", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.log.Info(string(passHash))

	id, err := h.storage.SaveUser(r.Context(), regReq.Email, regReq.Name, passHash)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			h.log.Warn("user already exists", err.Error())
			http.Error(w, "user already exists", http.StatusConflict)
		}
		h.log.Error("failed to save user", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.log.Info("successfully registred user")

	err = json.NewEncoder(w).Encode(id)
	if err != nil {
		h.log.Error("failed to encode id", err.Error())
	}
}

func (h *Handler) IsAdminHandler(w http.ResponseWriter, r *http.Request) {
	// Декодируем тело запроса
	var request struct {
		UserID int `json:"user_id"`
	}

	h.log.Info("checking if user is admin")
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.log.Error("failed to decode request", err.Error())
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	// Проверяем user_id
	if request.UserID == 0 {
		h.log.Error("invalid user ID")
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}
	res, err := h.storage.IsAdmin(r.Context(), int64(request.UserID))
	if err != nil {
		h.log.Warn("user not found", err.Error())
		http.Error(w, "user not found", http.StatusNotFound)
	}

	h.log.Info(
		"user checked",
		slog.Int("user_id", request.UserID),
		slog.Bool("is_admin", res),
	)
	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		h.log.Error("failed to encode user", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	const op = "handler.Health"

	log := h.log.With(
		slog.String("op", op),
		slog.String("method", r.Method),
	)

	log.Info("health check request")

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Barabara"))
}

func validateLogin(email string, password string) error {
	_, err := mail.ParseAddress(email)
	if err != nil {
		return status.Error(codes.InvalidArgument, "invalid email")
	}
	if password == "" {
		return status.Error(codes.InvalidArgument, "invalid password")
	}
	//if request.GetAppId() == emptyValue {
	//	return status.Error(codes.InvalidArgument, "invalid app id")
	//}
	return nil
}
