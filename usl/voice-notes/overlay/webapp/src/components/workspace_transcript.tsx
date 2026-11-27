import type { Post } from "@mattermost/types/posts";
import React, { useRef, useState } from "react";
import {
  postTranscriptReply,
  readVoiceFile,
  voiceLabel,
  workspaceTranscript,
} from "../api/workspace_speech";

export function WorkspaceTranscript({
  post,
  fileId,
}: {
  post: Post;
  fileId: string;
}) {
  const [text, setText] = useState("");
  const [notice, setNotice] = useState("");
  const [busy, setBusy] = useState(false);
  const [shared, setShared] = useState(false);
  const pendingId = useRef("");

  async function retry() {
    setBusy(true);
    setNotice("");
    try {
      const audio = await readVoiceFile(fileId);
      const result = await workspaceTranscript(audio, post.channel_id);
      setText(result);
      if (!result)
        setNotice(
          voiceLabel("No speech was recognized.", "Aucune parole reconnue."),
        );
    } catch (error) {
      setNotice(
        error instanceof Error
          ? error.message
          : voiceLabel("Try again later.", "Réessayez plus tard."),
      );
    } finally {
      setBusy(false);
    }
  }

  async function share() {
    setBusy(true);
    setNotice("");
    try {
      if (!pendingId.current) pendingId.current = crypto.randomUUID();
      await postTranscriptReply(post, text, pendingId.current);
      setShared(true);
    } catch (error) {
      setNotice(
        error instanceof Error
          ? error.message
          : voiceLabel("Try again later.", "Réessayez plus tard."),
      );
    } finally {
      setBusy(false);
    }
  }

  if (shared)
    return (
      <p role="status">
        {voiceLabel(
          "Transcript shared in the thread.",
          "Transcription publiée dans le fil.",
        )}
      </p>
    );
  return (
    <div className="VoiceMessageTranscript">
      {text && <p>{text}</p>}
      <button
        className="style--none color--link"
        type="button"
        disabled={busy}
        onClick={text ? share : retry}
      >
        {busy
          ? voiceLabel("Working…", "En cours…")
          : text
            ? voiceLabel(
                "Share transcript in thread",
                "Publier la transcription dans le fil",
              )
            : voiceLabel("Get transcript", "Transcrire le vocal")}
      </button>
      {notice && <p role="status">{notice}</p>}
    </div>
  );
}
