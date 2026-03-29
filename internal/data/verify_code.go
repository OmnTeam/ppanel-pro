package data

import "fmt"

const (
	verifySceneRegister      = "register"
	verifySceneSecurity      = "security"
	verifySceneResetPassword = "reset_password"
)

func verifyCodeCacheKey(method, scene, account string) string {
	switch method {
	case "mobile":
		return verifyCodeTelephoneCacheKey(scene, account)
	default:
		return verifyCodeEmailCacheKey(scene, account)
	}
}

func verifyCodeEmailCacheKey(scene, email string) string {
	return fmt.Sprintf("%s:%s:%s", AuthCodeCacheKey, scene, email)
}

func verifyCodeTelephoneCacheKey(scene, phoneNumber string) string {
	return fmt.Sprintf("%s:%s:%s", AuthCodeTelephoneCacheKey, scene, phoneNumber)
}

func parseVerifyType(verifyType int32) string {
	switch verifyType {
	case 1:
		return verifySceneRegister
	case 2:
		return verifySceneSecurity
	case 3:
		return verifySceneResetPassword
	default:
		return "unknown"
	}
}
