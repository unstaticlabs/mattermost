# Update log

Write the newest entry at the top. Each entry gives the target, the
decisions about the STOP and REVIEW items, and the result.

## 2026-11-12: Update to 12.1.0

- **Base:** `v12.1.0` (branch `usl/v12.1.0`). Production runs `12.0.2`,
  image `sha-b82e041`.
- **Target:** final tag `v12.1.0`. GitHub still marks the release as a
  pre-release, but `v12.1.0-rc4` and `v12.1.0` are the same commit.
- **12.0.2 fixes:** `v12.1.0` contains all the security and migration fixes
  of `v12.0.2`. Only release-only commits (version bumps, an E2E change) are
  not in `v12.1.0`. The brotli update is in `v12.1.0` as v1.2.7.
- **Branch:** `v12.1.0` plus the three USL commits. The USL diff is the same
  as on `usl/v12.0.0`. One conflict in this file only.

### Decisions

- STOP, accepted: `server/einterfaces/oauthproviders.go` adds the optional
  interface `OAuthClaimsProvider`. `OAuthProvider` does not change, so the
  USL provider compiles without a change.
- STOP, accepted: `server/channels/app/oauth.go` and
  `server/channels/web/oauth.go` pass an `OAuthAuthorization` struct in
  place of three values. The flow and the provider calls stay the same.
  `oauthAuthorizationScope` adds `AdditionalScopes` only when it is set.
- STOP, accepted: the feature flag `OpenIdAttributeSync` is `true`. The sync
  returns before it calls the provider when no attribute mapping exists.
  The server has no mapping, so nothing changes. If you add a mapping, the
  log shows "the identity provider cannot report claims".
- STOP, accepted: `server/config/client.go` adds `EnableFileSearch` to the
  full client config. The USL change is in the limited config (F2).
- REVIEW, accepted: `EnableFileIndexing` is `true` by default. It is an
  Elasticsearch setting, and the server has no Elasticsearch.
- REVIEW: prepackaged Calls goes to v1.12.6, Boards to v10.1.0, Dataminr to
  v3.0.0. ServiceNow is removed. The server does not use ServiceNow, and
  automatic prepackaged installs are off.
- Database: migrations `000230` to `000235` are new. Back up the database
  before the deploy.
- Workflows: the new docs workflows are on pull requests or manual. None
  runs on a push to `usl/**` or on a schedule.

### Tests

- `go test ./channels/app/oauthproviders/... ./config/` passes.
- `go build ./cmd/mattermost` passes.

## 2026-11-12: Prepare the 12.0.2 security patch

- **Base:** `v12.0.0`, with the USL OpenID provider and documentation commits
  retained on `usl/v12.0.0`.
- **Target:** official final tag `v12.0.2` at
  `13ae6c4cf1a8e73c35a4f0c8816a505506c7acdd`.
- **Integration:** merge the exact patch tag into a feature branch targeting the
  existing USL release branch. This retains upstream and downstream ancestry;
  no major or minor version change, live deployment, or release tag is included.
- **Preflight:** no STOP items, conflicts, or denylist matches. The custom
  provider, provider interface, OAuth flow and unlicensed client-button files
  are unchanged by this patch range.
- **Makefile review:** retain upstream's Dataminr prepackaged plugin update from
  2.0.0 to 3.0.0 and generated-test targets. This does not change persisted plugin
  activation settings. Confirm the enabled plugin set during the coordinated
  deployment checks.
- **Defaults:** the new `EnforceLogPathRoot` feature flag remains off by upstream
  default. No newly enabled feature flag or true configuration default was found.
- **Database:** migration `000242` restores the legacy `AttributeView`; the patch
  also repairs the existing `000216`, `000217` and `000223` migration paths. A
  qualified database/configuration backup is required before deployment. Image
  rollback alone is not the recovery plan after a database migration.
- **Workflow review:** the five changed upstream E2E workflows are manual,
  reusable, or limited to upstream CI/merge/release dispatch. They add no push
  trigger for `usl/**` and no schedule. The USL image publisher is unchanged.
- **Qualification:** the native OpenID provider and configuration suites pass
  against disposable PostgreSQL 18.6. Migration `000216`, `000217`, `000223`
  and `000242`, one-time cloud login, login-device token, log-path, remote-cluster,
  client-route and
  token-store regressions pass. An independent route probe fails on the old
  release, passes with this patch and fails when the route fix is removed. The
  native Linux amd64 Team Edition server build passes without Enterprise code.
  Focused upload-policy API, scheduled-post, thread-move and login-cookie checks
  also pass in native isolated parallel mode, using the same disposable database.
- **Fork CI compatibility:** preserve upstream's public regular Server CI and
  docs builds. Restrict private FIPS/Chainctl bootstrap, Docker Hub publication
  and AWS documentation preview/cleanup to `mattermost/mattermost`. A same-repo
  pull request on the USL fork does not have upstream's private identities.
  Use standard `ubuntu-latest` for USL's normal sharded server tests; keep the
  upstream larger-runner label for the upstream repository. The USL GHCR
  publisher and Team Edition product behavior are unchanged.
- **Remaining rollout checks:** qualify the database/configuration backup, build
  and inspect the published Team Edition image, then confirm F1–F4 runtime login,
  enabled plugins and image version during the coordinated deployment. Local
  preparation is not deployment proof.

## 2026-10-21: Check only, no update

- **Base:** `v12.0.0` (branch `usl/v12.0.0`). Production runs `12.0.0`,
  build `usl-39053810178`.
- **Result:** No update. Upstream has no final release after `v12.0.0`.

### `upstream/release-12.0`: the coming 12.0.1

Upstream has 6 commits after `v12.0.0`. It set the patch version to
`12.0.1`, but it did not tag `v12.0.1` yet.

- No STOP item. The USL commit applies without a conflict.
- New feature flag `EnforceLogPathRoot`, off by default.
- The prepackaged Dataminr plugin goes from v2.0.0 to v3.0.0.
- No database migrations.
- Only the end-to-end test workflows changed.

### `v12.1.0-rc1`: a release candidate

- STOP, accepted: `server/config/client.go` gets one new line,
  `EnableFileSearch`, in `GenerateClientConfig`. The USL change is in
  `GenerateLimitedClientConfig`, so it does not change F2.
- STOP, accepted: `server/channels/web/oauth.go` corrects the id of an
  error text (`singup` to `signup`). It does not change F1.
- The USL commit applies without a conflict.
- New feature flag `HealthDashboard`, off by default.
- New setting `EnableFileIndexing`, `true` by default. It is an
  Elasticsearch setting. Decide in the 12.1 update if you want it off.
- Prepackaged plugins: Calls goes from v1.12.5 to v1.12.6, and upstream
  removes ServiceNow.
- 4 new database migrations (`000230` to `000235`). Back up the database
  before the update.
- Some documentation workflows changed or are new in the fork.
