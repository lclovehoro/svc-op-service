package types

import (
	"context"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
	"k8s.io/klog/v2"
)

var (
	keycloakIssuer = ""
	clientID = ""
	clientSecret = ""
	redirectURI = ""
	scope = ""
)



type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Roles    `json:"roles"`
}

type Roles struct {
	Roles []string
}

type Token struct {
	Token string `json:"token"`
}

var code string = "123456"

func NewUser(username string, password string, code string) *User {
	return &User{
		Username: username,
		Password: password,
	}
}

func NewToken() *Token {
	return &Token{
		Token: "abcdefghijklmn",
	}
}

func (u *User) GetUsername() string {
	return u.Username
}

func (u *User) GetPassword() string {
	return u.Password
}

func GetCode() string {
	return code
}


func Auth2Handle(){
	provider, err := oidc.NewProvider(context.Background(), keycloakIssuer)
	if err != nil {
		klog.Errorf("Failed to create OIDC provider: %v", err)
		return
	}
	oauth2Config := &oauth2.Config{

		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  redirectURI,
		Scopes:       []string{oidc.ScopeOpenID, scope}
	}

}


