# The unstaticlabs fork of Mattermost

This fork is Mattermost Team Edition with a login through any OpenID Connect
provider. Unstatic Labs uses it with Pocket ID at `chat.unstaticlabs.com`.
Upstream v11 has OpenID Connect only in its licensed Enterprise code.

The fork changes only open-source (AGPLv3) code. It does not use or change
the code under `server/enterprise/`.

## The changes

| File | Change |
|---|---|
| `server/channels/app/oauthproviders/openid/` | A new OpenID Connect provider, with tests |
| `server/cmd/mattermost/main.go` | Registers the provider |
| `server/config/client.go` | Sends the OpenID button settings to the login page and the mobile apps without a licence |
| `server/config/client_test.go` | A test for that change |
| `.github/workflows/usl-image.yml` | Builds the image and pushes it to GHCR |

The provider uses the OAuth flow of the server, and reads the user from the
userinfo endpoint. It refuses a user whose email the provider does not mark
as verified.

## Configuration

Set these values, for example as `MM_OPENIDSETTINGS_*` environment variables:

| Setting | Value |
|---|---|
| `Enable` | `true` |
| `Id`, `Secret` | The client id and secret from the provider |
| `AuthEndpoint` | The authorization endpoint |
| `TokenEndpoint` | The token endpoint |
| `UserAPIEndpoint` | The userinfo endpoint |
| `Scope` | `openid profile email` (the default) |
| `ButtonText` | The text of the login button |
| `UsePreferredUsername` | `true` to use the `preferred_username` claim as the username |

The callback URL at the provider is `<SiteURL>/signup/openid/complete`.

## Branches, tags, and images

- Branch `usl/v<version>` is the upstream tag `v<version>` plus the changes above.
- A push to a `usl/**` branch builds `ghcr.io/unstaticlabs/mattermost-team-edition:sha-<short>`.
- A tag `v<version>-usl.<n>` builds `ghcr.io/unstaticlabs/mattermost-team-edition:<version>-usl.<n>`.

## Upgrade to a new Mattermost version

1. Fetch the new upstream tag, for example `v11.12.0`.
2. Create `usl/v11.12.0` from that tag, and cherry-pick the USL commits of
   the previous `usl/` branch.
3. Run `go test ./channels/app/oauthproviders/... ./config/` in `server/`.
4. Push the branch, then push the tag `v11.12.0-usl.1`.
