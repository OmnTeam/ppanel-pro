package auth

// InitializePlatformConfig 初始化平台配置
func InitializePlatformConfig(platform string) interface{} {
	var result interface{}
	switch platform {
	case "email":
		result = new(EmailAuthConfig).Marshal()
	case "mobile":
		result = new(MobileAuthConfig).Marshal()
	case "apple":
		result = new(AppleAuthConfig).Marshal()
	case "google":
		result = new(GoogleAuthConfig).Marshal()
	case "github":
		result = new(GithubAuthConfig).Marshal()
	case "facebook":
		result = new(FacebookAuthConfig).Marshal()
	case "telegram":
		result = new(TelegramAuthConfig).Marshal()
	case "device":
		result = new(DeviceConfig).Marshal()
	}
	return result
}
