# Update log

Write the newest entry at the top. Each entry gives the target, the
decisions about the STOP and REVIEW items, and the result.

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
