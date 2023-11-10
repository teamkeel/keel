package oauth

import "github.com/golang-jwt/jwt/v4"

type IdTokenClaims struct {
	jwt.RegisteredClaims
	UserClaims
}

// https://openid.net/specs/openid-connect-basic-1_0.html#StandardClaims
type UserClaims struct {
	// default 'email' scope claims
	Email         string `json:"email,omitempty"`
	EmailVerified bool   `json:"email_verified,omitempty"`

	// default 'profile' scope claims
	Name              string `json:"name,omitempty"`
	GivenName         string `json:"given_name,omitempty"`
	FamilyName        string `json:"family_name,omitempty"`
	MiddleName        string `json:"middle_name,omitempty"`
	NickName          string `json:"nick_name,omitempty"`
	PreferredUsername string `json:"preferred_username,omitempty"`
	Profile           string `json:"profile,omitempty"`
	Picture           string `json:"picture,omitempty"`
	Website           string `json:"website,omitempty"`
	Gender            string `json:"gender,omitempty"`
	ZoneInfo          string `json:"zoneinfo,omitempty"`
	Locale            string `json:"locale,omitempty"`

	// default 'phone' scope claims
	PhoneNumber         string `json:"phone_number,omitempty"`
	PhoneNumberVerified bool   `json:"phone_number_verified,omitempty"`
}
