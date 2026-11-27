# Voice notes in Mattermost

Web and desktop users can record a voice note in a channel or thread and listen before sending. When an administrator has configured the workspace speech service, sending also asks that service to transcribe the recording. Mattermost publishes the transcript as the message text and keeps the audio as a regular attachment in the same channel or thread.

If speech recognition is unavailable, the recording still posts. A channel member can use **Get transcript** on the saved note, read the result and choose **Share transcript in thread**. Reading the file and posting a reply both use the member’s native Mattermost session. The speech plugin receives audio supplied by that session. It has no dedicated file-store account.

## Build and install

Run `usl/voice-notes/build ./voice-build` with Git, Go 1.27.4, Node, npm, Python 3 and the ordinary Go build tools. The script checks out the pinned MIT-licensed recorder source, applies the USL changes, runs the Go and browser unit suites, and creates an installable Linux amd64 bundle. Set `VOICE_RECORDER_SOURCE` to a Git checkout containing the pinned revision for an offline source build. Go and npm dependencies must already be cached for a fully offline build.

An authorized workspace operator installs the resulting bundle through the normal Mattermost plugin manager. If the upstream recorder is installed, the new bundle updates the same plugin ID and keeps its settings. The CI artifact contains the USL build; upstream signatures apply only to upstream bundles. Use your workspace’s existing signing and review process.

In **System Console → Plugins → Mattermost Voice Messages**, set **Workspace speech endpoint** to the full transcription URL of the workspace-controlled service. The default is empty. The endpoint receives one multipart `file` with a generic filename and returns JSON with a `text` field. The shared Whisper service detects French and English. Channel names, post text and member credentials are never included in that request. Keep browser transcription off to use the workspace service for new notes.

The current shared service accepts recordings of up to five minutes with a total request size of 16 MiB. Larger audio can still be shared under the normal Mattermost file limits. Mattermost web and desktop use the web plugin interface. Official mobile apps can play the ordinary audio attachment. Recording in their native composers remains tracked upstream. [Mobile feature request](https://github.com/mattermost/mattermost-mobile/issues/9833)

## Upstream and maintenance

The recorder is [Mattermost Voice Messages 0.3.8](https://github.com/icoretech/mattermost-voice-messages/releases/tag/v0.3.8), pinned to `f72cb0508ac2e5179c56051375b21a9a528db001`. The upstream recorder owns capture, preview, file upload and channel or thread placement. The small changes to those files live in `recorder.patch`. The workspace speech client, native retry flow and their tests live in `overlay/`. The package includes the upstream license and the browser speech library’s notices.

When updating the pin, build and test the package, then exercise the normal send, a thread reply, hearing the note as another member and recovering a transcript after the speech service returns. Check that removing the reader from a private channel takes away access to both the post and the file.
