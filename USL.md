# The unstaticlabs fork of Mattermost

This fork is Mattermost Team Edition with a login through any OpenID Connect
provider. Unstatic Labs uses it with Pocket ID at `chat.unstaticlabs.com`.
Upstream has OpenID Connect only in its licensed Enterprise code.

The fork changes only open-source (AGPLv3) code. It does not use or change
the code under `server/enterprise/`.

## The documents

| Document | Use it to |
|---|---|
| [`usl/features.md`](usl/features.md) | Read the features that each USL build must keep, and how to check them |
| [`usl/update.md`](usl/update.md) | Update the fork to a new upstream version |
| [`usl/update-log.md`](usl/update-log.md) | Read the earlier checks, updates, and decisions |
| [`usl/check-upstream.sh`](usl/check-upstream.sh) | Check upstream before an update. The script only reads |
| [`usl/watched-files.txt`](usl/watched-files.txt) | The upstream files that the USL changes depend on |
| [`usl/denylist.txt`](usl/denylist.txt) | The upstream features that the USL build must keep disabled |

## The changes

| File | Change |
|---|---|
| `server/channels/app/oauthproviders/openid/` | A new OpenID Connect provider, with tests |
| `server/cmd/mattermost/main.go` | Registers the provider |
| `server/config/client.go` | Sends the OpenID button settings to the login page and the mobile apps without a licence |
| `server/config/client_test.go` | A test for that change |
| `.github/workflows/usl-image.yml` | Builds the image and pushes it to GHCR |
| `USL.md`, `usl/` | The documentation of the fork |

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

Follow [`usl/update.md`](usl/update.md). Start with `usl/check-upstream.sh`.
