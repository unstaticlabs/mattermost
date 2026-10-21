# Update the fork to a new upstream version

This procedure moves the USL changes to a new upstream tag. It keeps the
features in [`features.md`](features.md). It stops the update if upstream
enables a feature that you want to keep disabled.

The steps in order:

1. Check upstream. You change nothing.
2. Decide about each STOP and REVIEW item.
3. Create the new branch.
4. Test.
5. Deploy the branch image.
6. Do the feature checks.
7. Tag the release.
8. Write the update log.

## 1. Check upstream

Run the check from a `usl/v<version>` branch:

```sh
usl/check-upstream.sh                 # newest final tag, else the release branch
usl/check-upstream.sh v12.1.0         # a specific target
usl/check-upstream.sh --no-fetch ...  # do not fetch again
```

The script only reads. It shows:

- The final releases and the release candidates after the base tag.
- The commits on the release branch that upstream did not release yet.
- The changes in the [watched files](watched-files.txt).
- The feature flags, the new defaults that are `true`, and the
  prepackaged plugins.
- The new database migrations.
- The changed GitHub workflows.
- The matches with the [denylist](denylist.txt).
- A conflict test of the USL commits on the target, in memory only.

The script ends with exit status `1` if one or more items are STOP items.

Use only a final tag, for example `v12.1.0`, for the production build. Do
not deploy a release candidate (`-rc<n>`) or a release branch to
production.

## 2. Decide about each item

Do not continue while a STOP item has no decision. Write each decision in
the [update log](update-log.md).

| Item | What to do |
|---|---|
| STOP: a watched file changed | Read the diff with `git diff <base> <target> -- <file>`. Make sure that the USL feature still works. If upstream changed the interface or the flow, change the USL code on the new branch. |
| STOP: a feature flag is `true` by default | Decide if you want the feature. If not, keep it off in the deployment with `MM_FEATUREFLAGS_<NAME>=false`, and add the flag to the denylist. |
| STOP: a denylist match | Upstream enables a feature that you decided to keep disabled. Keep it off in the deployment before you deploy. If you cannot keep it off with the configuration, do not update. |
| STOP: a conflict | The cherry-pick in step 3 will stop at this file. Plan the fix before you start. |
| REVIEW: a new default that is `true` | Decide if you want the setting. If not, set it to `false` in the deployment with `MM_<SECTION>_<SETTING>=false`. |
| REVIEW: prepackaged plugins | Make sure that a removed plugin is not one that you use. After the deploy, make sure that no new plugin is enabled in **System Console > Plugins**. |
| Database migrations | Back up the database before the deploy. You cannot go back to the old version without this backup. |
| Changed GitHub workflows | The fork runs the upstream workflows that are active. Make sure that a new workflow does not run on a push to `usl/**` or on a schedule. If it does, run `gh workflow disable <file> -R unstaticlabs/mattermost`. |

Keep a feature off with the configuration of the deployment, not with a
change in the code. A change in the code is one more change to carry in
each update.

## 3. Create the new branch

The USL commits are the commits from the old base tag to the old USL
branch. The `usl/` folder is part of these commits.

```sh
git switch -c usl/v<new> v<new>
git cherry-pick v<old>..usl/v<old>
```

If a cherry-pick stops with a conflict, fix the files, then run
`git cherry-pick --continue`.

## 4. Test

In `server/`:

```sh
make setup-go-work
go test ./channels/app/oauthproviders/... ./config/
```

Then run `usl/check-upstream.sh --no-fetch` again on the new branch. The
base is now the new tag.

## 5. Deploy the branch image

Back up the database before this step. A new version can run database
migrations that you cannot undo.

1. Push the branch: `git push origin usl/v<new>`.
2. Wait for the "USL image" workflow. It pushes the image `sha-<short>`.
3. Set the keep-off settings from step 2 in the deployment.
4. Deploy the `sha-<short>` image.

## 6. Do the feature checks

Do all the checks in [`features.md`](features.md). If a check fails, go
back to the old image and restore the database backup.

## 7. Tag the release

```sh
git tag v<new>-usl.1
git push origin v<new>-usl.1
```

The workflow pushes the image `<new>-usl.1`. Deploy this tag, so the
deployment does not use a `sha-` image.

## 8. Write the update log

Add an entry to [`update-log.md`](update-log.md). Write the target, the
decisions from step 2, and the result of the checks.

## Go back to the old version

1. Deploy the old image tag, for example `12.0.0-usl.1`.
2. If the new version ran database migrations, restore the database
   backup before you start the old image.
