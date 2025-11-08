package auth

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

// AuthRepo defines repository interface for auth operations
type AuthRepo interface {
	// CheckUserExistByEmail checks if user exists by email
	CheckUserExistByEmail(ctx context.Context, email string) (bool, error)
	// CheckUserExistByTelephone checks if user exists by telephone (E.164 format)
	CheckUserExistByTelephone(ctx context.Context, telephoneAreaCode, telephone string) (bool, error)

	// UserLogin logs in user with email and password
	UserLogin(ctx context.Context,  email, password, ip, userAgent string) (*LoginResult, error)
	// TelephoneLogin logs in user with telephone and password or code
	TelephoneLogin(ctx context.Context,  telephoneAreaCode, telephone, password, telephoneCode, ip, userAgent string) (*LoginResult, error)

	// UserRegister registers a new user with email
	UserRegister(ctx context.Context,  email, password, invite, code, ip, userAgent string) (*LoginResult, error)
	// TelephoneRegister registers a new user with telephone
	TelephoneRegister(ctx context.Context,  telephoneAreaCode, telephone, password, invite, code, ip, userAgent string) (*LoginResult, error)

	// ResetPassword resets user password with email
	ResetPassword(ctx context.Context,  email, password, code, ip, userAgent string) (*LoginResult, error)
	// TelephoneResetPassword resets user password with telephone
	TelephoneResetPassword(ctx context.Context,  telephoneAreaCode, telephone, password, code, ip, userAgent string) (*LoginResult, error)
}

// LoginResult contains login result information
type LoginResult struct {
	Token string
}

// AuthUsecase handles auth business logic
type AuthUsecase struct {
	repo AuthRepo
	log  *log.Helper
}

// NewAuthUsecase creates a new auth usecase
func NewAuthUsecase(repo AuthRepo, logger log.Logger) *AuthUsecase {
	return &AuthUsecase{
		repo: repo,
		log:  log.NewHelper(log.With(logger, "module", "biz/auth")),
	}
}

// CheckUser checks if user exists by email
func (uc *AuthUsecase) CheckUser(ctx context.Context,  email string) (bool, error) {
	exist, err := uc.repo.CheckUserExistByEmail(ctx, email)
	if err != nil {
		uc.log.Errorw("CheckUserExistByEmail error", "error", err,  "email", email)
		return false, err
	}
	return exist, nil
}

// CheckUserTelephone checks if user exists by telephone
func (uc *AuthUsecase) CheckUserTelephone(ctx context.Context,  telephoneAreaCode, telephone string) (bool, error) {
	exist, err := uc.repo.CheckUserExistByTelephone(ctx, telephoneAreaCode, telephone)
	if err != nil {
		uc.log.Errorw("CheckUserExistByTelephone error", "error", err,  "telephone", telephone)
		return false, err
	}
	return exist, nil
}

// UserLogin logs in user with email and password
func (uc *AuthUsecase) UserLogin(ctx context.Context,  email, password, ip, userAgent string) (*LoginResult, error) {
	result, err := uc.repo.UserLogin(ctx, email, password, ip, userAgent)
	if err != nil {
		uc.log.Errorw("UserLogin error", "error", err,  "email", email)
		return nil, err
	}
	return result, nil
}

// TelephoneLogin logs in user with telephone and password or code
func (uc *AuthUsecase) TelephoneLogin(ctx context.Context,  telephoneAreaCode, telephone, password, telephoneCode, ip, userAgent string) (*LoginResult, error) {
	result, err := uc.repo.TelephoneLogin(ctx, telephoneAreaCode, telephone, password, telephoneCode, ip, userAgent)
	if err != nil {
		uc.log.Errorw("TelephoneLogin error", "error", err,  "telephone", telephone)
		return nil, err
	}
	return result, nil
}

// UserRegister registers a new user with email
func (uc *AuthUsecase) UserRegister(ctx context.Context,  email, password, invite, code, ip, userAgent string) (*LoginResult, error) {
	result, err := uc.repo.UserRegister(ctx, email, password, invite, code, ip, userAgent)
	if err != nil {
		uc.log.Errorw("UserRegister error", "error", err,  "email", email)
		return nil, err
	}
	return result, nil
}

// TelephoneRegister registers a new user with telephone
func (uc *AuthUsecase) TelephoneRegister(ctx context.Context,  telephoneAreaCode, telephone, password, invite, code, ip, userAgent string) (*LoginResult, error) {
	result, err := uc.repo.TelephoneRegister(ctx, telephoneAreaCode, telephone, password, invite, code, ip, userAgent)
	if err != nil {
		uc.log.Errorw("TelephoneRegister error", "error", err,  "telephone", telephone)
		return nil, err
	}
	return result, nil
}

// ResetPassword resets user password with email
func (uc *AuthUsecase) ResetPassword(ctx context.Context,  email, password, code, ip, userAgent string) (*LoginResult, error) {
	result, err := uc.repo.ResetPassword(ctx, email, password, code, ip, userAgent)
	if err != nil {
		uc.log.Errorw("ResetPassword error", "error", err,  "email", email)
		return nil, err
	}
	return result, nil
}

// TelephoneResetPassword resets user password with telephone
func (uc *AuthUsecase) TelephoneResetPassword(ctx context.Context,  telephoneAreaCode, telephone, password, code, ip, userAgent string) (*LoginResult, error) {
	result, err := uc.repo.TelephoneResetPassword(ctx, telephoneAreaCode, telephone, password, code, ip, userAgent)
	if err != nil {
		uc.log.Errorw("TelephoneResetPassword error", "error", err,  "telephone", telephone)
		return nil, err
	}
	return result, nil
}
