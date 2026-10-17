// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

// Package oauthopenid is a generic OpenID Connect login provider for Team
// Edition. It is part of the unstaticlabs/mattermost fork, not of upstream
// Mattermost.
//
// It uses only the OAuth flow of the open-source server (app/oauth.go): the
// authorization code flow with the state token and the OAuth cookie, the
// token endpoint, and then the userinfo endpoint. The claims come from the
// userinfo response, which the server gets from the identity provider over
// TLS with the access token. So there is no ID token to verify here.
//
// It reads OpenIdSettings (MM_OPENIDSETTINGS_*): Id, Secret, AuthEndpoint,
// TokenEndpoint, UserAPIEndpoint (the userinfo endpoint), Scope,
// ButtonText, ButtonColor, and UsePreferredUsername.
package oauthopenid

import (
	"encoding/json"
	"errors"
	"io"
	"strings"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/shared/mlog"
	"github.com/mattermost/mattermost/server/public/shared/request"
	"github.com/mattermost/mattermost/server/v8/einterfaces"
)

type OpenIDProvider struct {
}

// OpenIDUser holds the standard OpenID Connect claims that the provider uses.
type OpenIDUser struct {
	Sub               string `json:"sub"`
	Email             string `json:"email"`
	EmailVerified     *bool  `json:"email_verified"`
	Name              string `json:"name"`
	GivenName         string `json:"given_name"`
	FamilyName        string `json:"family_name"`
	PreferredUsername string `json:"preferred_username"`
}

func init() {
	provider := &OpenIDProvider{}
	einterfaces.RegisterOAuthProvider(model.ServiceOpenid, provider)
}

func openIDUserFromJSON(data io.Reader) (*OpenIDUser, error) {
	var u OpenIDUser
	if err := json.NewDecoder(data).Decode(&u); err != nil {
		return nil, err
	}
	return &u, nil
}

// IsValid refuses a user without a subject or an email, and a user whose
// email the identity provider does not mark as verified. Mattermost finds an
// existing account by email when a user switches to this login method, so an
// unverified email could take over the account of another user.
func (u *OpenIDUser) IsValid() error {
	if u.Sub == "" {
		return errors.New("user subject (sub) can't be empty")
	}
	if u.Email == "" {
		return errors.New("user e-mail should not be empty")
	}
	if u.EmailVerified == nil || !*u.EmailVerified {
		return errors.New("user e-mail is not verified by the identity provider")
	}
	return nil
}

func userFromOpenIDUser(logger mlog.LoggerIFace, u *OpenIDUser, settings *model.SSOSettings) *model.User {
	user := &model.User{}

	// The username: the preferred username if the settings ask for it, else
	// the part of the email before the @. To match the other providers, a
	// preferred username is also split at the @.
	username := strings.Split(u.Email, "@")[0]
	if settings != nil && model.SafeDereference(settings.UsePreferredUsername) && u.PreferredUsername != "" {
		username = strings.Split(u.PreferredUsername, "@")[0]
	}
	user.Username = model.CleanUsername(logger, username)

	if u.GivenName != "" || u.FamilyName != "" {
		user.FirstName = u.GivenName
		user.LastName = u.FamilyName
	} else {
		first, last, _ := strings.Cut(strings.TrimSpace(u.Name), " ")
		user.FirstName = first
		user.LastName = strings.TrimSpace(last)
	}

	user.Email = strings.ToLower(u.Email)
	authData := u.Sub
	user.AuthData = &authData
	user.AuthService = model.ServiceOpenid

	return user
}

func (p *OpenIDProvider) GetUserFromJSON(rctx request.CTX, data io.Reader, tokenUser *model.User, settings *model.SSOSettings) (*model.User, error) {
	u, err := openIDUserFromJSON(data)
	if err != nil {
		return nil, err
	}
	if err = u.IsValid(); err != nil {
		return nil, err
	}

	return userFromOpenIDUser(rctx.Logger(), u, settings), nil
}

func (p *OpenIDProvider) GetSSOSettings(_ request.CTX, config *model.Config, service string) (*model.SSOSettings, error) {
	return &config.OpenIdSettings, nil
}

// GetUserFromIdToken returns no user: the claims come from the userinfo
// endpoint instead (see the package comment).
func (p *OpenIDProvider) GetUserFromIdToken(_ request.CTX, idToken string) (*model.User, error) {
	return nil, nil
}

func (p *OpenIDProvider) IsSameUser(_ request.CTX, dbUser, oAuthUser *model.User) bool {
	if dbUser == nil || oAuthUser == nil || dbUser.AuthData == nil || oAuthUser.AuthData == nil {
		return false
	}
	return *dbUser.AuthData == *oAuthUser.AuthData
}
