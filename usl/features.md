# The features of the USL build

This file is the specification of the fork. Each new USL build must keep
all the features in this file. Do the checks after each update.

If you add a USL change, add its feature here. Add the upstream files that
it depends on to [`watched-files.txt`](watched-files.txt).

## F1. Login with OpenID Connect in Team Edition

The login page shows a button for the OpenID Connect provider (Pocket ID).
The user logs in without a Mattermost licence.

| Part | Where |
|---|---|
| The provider | `server/channels/app/oauthproviders/openid/` |
| The registration of the provider | `server/cmd/mattermost/main.go` |
| The OAuth flow of the server (upstream) | `server/channels/app/oauth.go`, `server/channels/web/oauth.go` |
| The provider interface (upstream) | `server/einterfaces/oauthproviders.go` |

The rules:

- The callback URL is `<SiteURL>/signup/openid/complete`.
- The provider reads the user from the userinfo endpoint.
- The provider refuses a user whose email is not marked as verified.
- If `UsePreferredUsername` is `true`, the username comes from the
  `preferred_username` claim.

Checks:

1. In `server/`, run `go test ./channels/app/oauthproviders/...`.
2. Open `https://chat.unstaticlabs.com/login`. Make sure that the OpenID
   button is there.
3. Log in with Pocket ID. Make sure that you get your existing account,
   not a new account.

## F2. The OpenID button without a licence, also in the mobile apps

The server sends the OpenID button settings in the limited client
configuration. The login page and the mobile apps read this configuration
before the user logs in. Upstream sends empty values without a licence.

| Part | Where |
|---|---|
| The change | `server/config/client.go`, in `GenerateLimitedClientConfig` |
| The test | `server/config/client_test.go` |

Checks:

1. In `server/`, run `go test ./config/`.
2. Run `curl -s https://chat.unstaticlabs.com/api/v4/config/client?format=old | jq .EnableSignUpWithOpenId`.
   The result must be `"true"`.
3. Log in with Pocket ID in the mobile app.

## F3. The image in GHCR

A workflow builds the Team Edition package and the image with the
upstream Dockerfile. It pushes the image to GHCR.

| Part | Where |
|---|---|
| The workflow | `.github/workflows/usl-image.yml` |
| The build (upstream) | `server/Makefile`, `server/build/release.mk`, `server/build/Dockerfile` |

| Push | Image tag |
|---|---|
| A `usl/**` branch | `ghcr.io/unstaticlabs/mattermost-team-edition:sha-<short>` |
| A tag `v<version>-usl.<n>` | `ghcr.io/unstaticlabs/mattermost-team-edition:<version>-usl.<n>` |

Checks:

1. Make sure that the "USL image" run is green in the GitHub Actions of
   `unstaticlabs/mattermost`.
2. Make sure that the version on the "About" dialog is the new version.

## F4. No Enterprise code

The fork changes only open-source (AGPLv3) code. It does not use or change
the code under `server/enterprise/`.

Check: run `git diff --stat v<version>..usl/v<version> -- server/enterprise`.
The output must be empty.
