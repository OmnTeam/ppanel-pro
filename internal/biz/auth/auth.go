package auth

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

type VerifyConfig struct {
	CaptchaType                    string
	TurnstileSiteKey               string
	TurnstileSecret                string
	EnableUserLoginCaptcha         bool
	EnableUserRegisterCaptcha      bool
	EnableAdminLoginCaptcha        bool
	EnableUserResetPasswordCaptcha bool
}

type RequestMeta struct {
	Identifier  string
	LoginType   string
	IP          string
	UserAgent   string
	CfToken     string
	CaptchaID   string
	CaptchaCode string
	SliderToken string
}

type UserLoginParams struct {
	Email    string
	Password string
	Meta     RequestMeta
}

type TelephoneLoginParams struct {
	TelephoneAreaCode string
	Telephone         string
	Password          string
	TelephoneCode     string
	Meta              RequestMeta
}

type UserRegisterParams struct {
	Email    string
	Password string
	Invite   string
	Code     string
	Meta     RequestMeta
}

type TelephoneRegisterParams struct {
	TelephoneAreaCode string
	Telephone         string
	Password          string
	Invite            string
	Code              string
	Meta              RequestMeta
}

type ResetPasswordParams struct {
	Email    string
	Password string
	Code     string
	Meta     RequestMeta
}

type TelephoneResetPasswordParams struct {
	TelephoneAreaCode string
	Telephone         string
	Password          string
	Code              string
	Meta              RequestMeta
}

type AuthRepo interface {
	CheckUserExistByEmail(ctx context.Context, email string) (bool, error)
	CheckUserExistByTelephone(ctx context.Context, telephoneAreaCode, telephone string) (bool, error)
	GetVerifyConfig(ctx context.Context) (*VerifyConfig, error)
	VerifyCaptcha(ctx context.Context, config *VerifyConfig, meta RequestMeta) error
	UserLogin(ctx context.Context, params *UserLoginParams) (*LoginResult, error)
	TelephoneLogin(ctx context.Context, params *TelephoneLoginParams) (*LoginResult, error)
	UserRegister(ctx context.Context, params *UserRegisterParams) (*LoginResult, error)
	TelephoneRegister(ctx context.Context, params *TelephoneRegisterParams) (*LoginResult, error)
	ResetPassword(ctx context.Context, params *ResetPasswordParams) (*LoginResult, error)
	TelephoneResetPassword(ctx context.Context, params *TelephoneResetPasswordParams) (*LoginResult, error)
}

type LoginResult struct {
	Token string
}

type verifyScene string

const (
	verifySceneLogin         verifyScene = "login"
	verifySceneRegister      verifyScene = "register"
	verifySceneResetPassword verifyScene = "reset_password"
)

type AuthUsecase struct {
	repo AuthRepo
	log  *log.Helper
}

func NewAuthUsecase(repo AuthRepo, logger log.Logger) *AuthUsecase {
	return &AuthUsecase{
		repo: repo,
		log:  log.NewHelper(log.With(logger, "module", "biz/auth")),
	}
}

func (uc *AuthUsecase) CheckUser(ctx context.Context, email string) (bool, error) {
	exist, err := uc.repo.CheckUserExistByEmail(ctx, email)
	if err != nil {
		uc.log.Errorw("CheckUserExistByEmail error", "error", err, "email", email)
		return false, err
	}
	return exist, nil
}

func (uc *AuthUsecase) CheckUserTelephone(ctx context.Context, telephoneAreaCode, telephone string) (bool, error) {
	exist, err := uc.repo.CheckUserExistByTelephone(ctx, telephoneAreaCode, telephone)
	if err != nil {
		uc.log.Errorw("CheckUserExistByTelephone error", "error", err, "telephone", telephone)
		return false, err
	}
	return exist, nil
}

func (uc *AuthUsecase) UserLogin(ctx context.Context, params *UserLoginParams) (*LoginResult, error) {
	if err := uc.verifyAccess(ctx, verifySceneLogin, params.Meta); err != nil {
		return nil, err
	}
	result, err := uc.repo.UserLogin(ctx, params)
	if err != nil {
		uc.log.Errorw("UserLogin error", "error", err, "email", params.Email)
		return nil, err
	}
	return result, nil
}

func (uc *AuthUsecase) TelephoneLogin(ctx context.Context, params *TelephoneLoginParams) (*LoginResult, error) {
	if err := uc.verifyAccess(ctx, verifySceneLogin, params.Meta); err != nil {
		return nil, err
	}
	result, err := uc.repo.TelephoneLogin(ctx, params)
	if err != nil {
		uc.log.Errorw("TelephoneLogin error", "error", err, "telephone", params.Telephone)
		return nil, err
	}
	return result, nil
}

func (uc *AuthUsecase) UserRegister(ctx context.Context, params *UserRegisterParams) (*LoginResult, error) {
	if err := uc.verifyAccess(ctx, verifySceneRegister, params.Meta); err != nil {
		return nil, err
	}
	result, err := uc.repo.UserRegister(ctx, params)
	if err != nil {
		uc.log.Errorw("UserRegister error", "error", err, "email", params.Email)
		return nil, err
	}
	return result, nil
}

func (uc *AuthUsecase) TelephoneRegister(ctx context.Context, params *TelephoneRegisterParams) (*LoginResult, error) {
	if err := uc.verifyAccess(ctx, verifySceneRegister, params.Meta); err != nil {
		return nil, err
	}
	result, err := uc.repo.TelephoneRegister(ctx, params)
	if err != nil {
		uc.log.Errorw("TelephoneRegister error", "error", err, "telephone", params.Telephone)
		return nil, err
	}
	return result, nil
}

func (uc *AuthUsecase) ResetPassword(ctx context.Context, params *ResetPasswordParams) (*LoginResult, error) {
	if err := uc.verifyAccess(ctx, verifySceneResetPassword, params.Meta); err != nil {
		return nil, err
	}
	result, err := uc.repo.ResetPassword(ctx, params)
	if err != nil {
		uc.log.Errorw("ResetPassword error", "error", err, "email", params.Email)
		return nil, err
	}
	return result, nil
}

func (uc *AuthUsecase) TelephoneResetPassword(ctx context.Context, params *TelephoneResetPasswordParams) (*LoginResult, error) {
	if err := uc.verifyAccess(ctx, verifySceneResetPassword, params.Meta); err != nil {
		return nil, err
	}
	result, err := uc.repo.TelephoneResetPassword(ctx, params)
	if err != nil {
		uc.log.Errorw("TelephoneResetPassword error", "error", err, "telephone", params.Telephone)
		return nil, err
	}
	return result, nil
}

func (uc *AuthUsecase) verifyAccess(ctx context.Context, scene verifyScene, meta RequestMeta) error {
	config, err := uc.repo.GetVerifyConfig(ctx)
	if err != nil {
		uc.log.Errorw("GetVerifyConfig error", "error", err, "scene", scene)
		return err
	}

	if !config.enabled(scene) {
		return nil
	}
	if err := uc.repo.VerifyCaptcha(ctx, config, meta); err != nil {
		uc.log.Errorw("VerifyCaptcha error", "error", err, "scene", scene)
		return err
	}
	return nil
}

func (c *VerifyConfig) enabled(scene verifyScene) bool {
	if c == nil {
		return false
	}
	switch scene {
	case verifySceneLogin:
		return c.EnableUserLoginCaptcha
	case verifySceneRegister:
		return c.EnableUserRegisterCaptcha
	case verifySceneResetPassword:
		return c.EnableUserResetPasswordCaptcha
	default:
		return false
	}
}
