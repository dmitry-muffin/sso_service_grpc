package auth

import (
	"context"
	"errors"
	"fmt"
	//"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"sso/internal/domain/models"
	"sso/internal/lib/jwt"
	"sso/internal/storage"
	"time"
)

type Auth struct {
	log         *slog.Logger
	usrSaver    UserSaver
	usrProvider UserProvider
	appProvider AppProvider
	tokenTTL    time.Duration
}

type UserSaver interface {
	SaveUser(
		ctx context.Context,
		email string,
		name string,
		passHash []byte,
	) (uid int, err error)
}

type UserProvider interface {
	User(ctx context.Context, email string) (user models.User, err error)
	IsAdmin(ctx context.Context, userId int64) (isAdmin bool, err error)
}

type AppProvider interface {
	App(ctx context.Context, appId int64) (models.App, error)
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidAppId       = errors.New("invalid app id")
	ErrUserExists         = errors.New("user already exists")
)

func New(
	log *slog.Logger,
	userProvider UserProvider,
	appProvider AppProvider,
	userSaver UserSaver,
	tokenTTL time.Duration,
) *Auth {
	return &Auth{
		usrSaver:    userSaver,
		usrProvider: userProvider,
		log:         log,
		appProvider: appProvider,
		tokenTTL:    tokenTTL,
	}
}

func (a *Auth) Login(
	ctx context.Context,
	email string,
	password string,
	appID int32) (string, error) {

	const op = "auth.Login"
	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)
	log.Info("attempting to login user")
	user, err := a.usrProvider.User(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			a.log.Warn("user not found", err.Error())
			return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}
		a.log.Error("failed to login", err.Error())
		return "", fmt.Errorf("%s: %w", op, err)
	}
	log.Info("here 1")
	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		a.log.Info("invalid credentials", err.Error())
		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}
	log.Info("here 2")
	app, errr := a.appProvider.App(ctx, int64(appID))
	if errr != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	log.Info("here 3")
	log.Info("user successfully logged in")

	token, err := jwt.NewToken(user, app, a.tokenTTL)
	if err != nil {
		a.log.Error("failed to create token", err.Error())
		return "", fmt.Errorf("%s: %w", op, err)
	}
	return token, nil
}

func (a *Auth) RegisterNewUser(
	ctx context.Context,
	email string,
	name string,
	password string,
) (int64, error) {
	const op = "auth.RegisterNewUser"
	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)
	log.Info("registering new user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to hash password", err.Error())
		return 0, fmt.Errorf("%w: %s", op, err)
	}
	log.Info(string(passHash))
	id, err := a.usrSaver.SaveUser(ctx, email, name, passHash)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			log.Warn("user already exists", err.Error())
			return 0, fmt.Errorf("%s: %w", op, ErrUserExists)
		}
		log.Error("failed to save user", err.Error())
		return 0, fmt.Errorf("%w: %s", op, err)
	}

	log.Info("successfully registred user")
	return int64(id), nil
}

func (a *Auth) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	const op = "auth.IsAdmin"
	log := a.log.With(
		slog.String("op", op),
		slog.String("userID", fmt.Sprint(userID)))
	log.Info("checking if user is admin")

	isAdmin, err := a.usrProvider.IsAdmin(ctx, userID)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Warn("user not found", ErrInvalidAppId)
		}
		return false, fmt.Errorf("%s: %w", op, err)
	}
	log.Info("checked if user is admin", slog.Bool("isAdmin", isAdmin))

	return isAdmin, nil
}
