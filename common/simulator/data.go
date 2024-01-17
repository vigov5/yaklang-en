// Package simulator
// @Author bcy2007  2023/8/17 16:19
package simulator

var ElementAttribute = []string{
	"placeholder", "id", "name", "value", "alt",
}

var ElementProperty = []string{
	"innerHTML",
}

var ElementKeyword = append(ElementAttribute, ElementProperty...)

var usernameKeyword = []string{
	"username", "admin",
	"User name", "Account name", "Account",
	"telephone", "email",
	"Mobile phone", "Phone", "Email",
}

var simpleUsernameKeyword = []string{
	"user", "admin", "tele", "email",
	"User", "Account", "Account", "Mobile phone", "Phone", "Email",
}

var passwordKeyword = []string{
	"password", "pwd", "Password",
}

var simplePasswordKeyword = []string{
	"pass", "pwd", "Password",
}

var captchaKeyword = []string{
	"captcha", "register", "check", "validate",
	"Verification code", "Verification code", "Registration code", "verifica", "verify",
}

var simpleCaptchaKeyword = []string{
	"capt", "reg", "Verification", "Verification", "Register", "validate", "verif",
}

var loginKeyword = []string{
	"Login", "login", "submit",
}

var KeywordDict = map[string][]string{
	"Username": usernameKeyword,
	"Password": passwordKeyword,
	"Captcha":  captchaKeyword,
	"Login":    loginKeyword,
}

var SimpleKeywordDict = map[string][]string{
	"Username": simpleUsernameKeyword,
	"Password": simplePasswordKeyword,
	"Captcha":  simpleCaptchaKeyword,
	"Login":    loginKeyword,
}
