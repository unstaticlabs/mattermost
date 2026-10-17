// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package oauthopenid

import (
	"strings"
	"testing"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/shared/request"
	"github.com/mattermost/mattermost/server/v8/einterfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const pocketIDUser = `{
	"sub": "6f1c2a9e-0b2d-4c3a-9f7e-1a2b3c4d5e6f",
	"email": "Roger@Example.com",
	"email_verified": true,
	"name": "Roger Example",
	"given_name": "Roger",
	"family_name": "Example",
	"preferred_username": "roger",
	"groups": ["chat"]
}`

func TestRegistered(t *testing.T) {
	assert.IsType(t, &OpenIDProvider{}, einterfaces.GetOAuthProvider(model.ServiceOpenid))
}

func TestGetUserFromJSON(t *testing.T) {
	rctx := request.TestContext(t)
	provider := &OpenIDProvider{}

	t.Run("valid user", func(t *testing.T) {
		user, err := provider.GetUserFromJSON(rctx, strings.NewReader(pocketIDUser), nil, nil)
		require.NoError(t, err)
		assert.Equal(t, "6f1c2a9e-0b2d-4c3a-9f7e-1a2b3c4d5e6f", *user.AuthData)
		assert.Equal(t, model.ServiceOpenid, user.AuthService)
		assert.Equal(t, "roger@example.com", user.Email)
		assert.Equal(t, "roger", user.Username)
		assert.Equal(t, "Roger", user.FirstName)
		assert.Equal(t, "Example", user.LastName)
	})

	t.Run("preferred username when the settings ask for it", func(t *testing.T) {
		data := strings.Replace(pocketIDUser, `"preferred_username": "roger"`, `"preferred_username": "rx@corp"`, 1)
		settings := &model.SSOSettings{UsePreferredUsername: model.NewPointer(true)}
		user, err := provider.GetUserFromJSON(rctx, strings.NewReader(data), nil, settings)
		require.NoError(t, err)
		assert.Equal(t, "rx", user.Username)
	})

	t.Run("name split when there are no given and family names", func(t *testing.T) {
		data := `{"sub": "s1", "email": "a@b.c", "email_verified": true, "name": "Ana Maria Lopez"}`
		user, err := provider.GetUserFromJSON(rctx, strings.NewReader(data), nil, nil)
		require.NoError(t, err)
		assert.Equal(t, "Ana", user.FirstName)
		assert.Equal(t, "Maria Lopez", user.LastName)
		assert.Equal(t, "a", user.Username)
	})

	for name, data := range map[string]string{
		"unverified email":     `{"sub": "s1", "email": "a@b.c", "email_verified": false}`,
		"missing verification": `{"sub": "s1", "email": "a@b.c"}`,
		"missing subject":      `{"email": "a@b.c", "email_verified": true}`,
		"missing email":        `{"sub": "s1", "email_verified": true}`,
		"not JSON":             `not json`,
	} {
		t.Run("refuses "+name, func(t *testing.T) {
			user, err := provider.GetUserFromJSON(rctx, strings.NewReader(data), nil, nil)
			require.Error(t, err)
			assert.Nil(t, user)
		})
	}
}

func TestGetSSOSettings(t *testing.T) {
	config := &model.Config{}
	config.SetDefaults()
	settings, err := (&OpenIDProvider{}).GetSSOSettings(request.TestContext(t), config, model.ServiceOpenid)
	require.NoError(t, err)
	assert.Same(t, &config.OpenIdSettings, settings)
}

func TestIsSameUser(t *testing.T) {
	rctx := request.TestContext(t)
	provider := &OpenIDProvider{}
	a, b, c := "sub-1", "sub-1", "sub-2"

	assert.True(t, provider.IsSameUser(rctx, &model.User{AuthData: &a}, &model.User{AuthData: &b}))
	assert.False(t, provider.IsSameUser(rctx, &model.User{AuthData: &a}, &model.User{AuthData: &c}))
	assert.False(t, provider.IsSameUser(rctx, &model.User{}, &model.User{AuthData: &b}))
	assert.False(t, provider.IsSameUser(rctx, nil, &model.User{AuthData: &b}))
}
