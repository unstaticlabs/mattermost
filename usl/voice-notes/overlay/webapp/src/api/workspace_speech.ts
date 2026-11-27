import type { Post } from "@mattermost/types/posts";
import { getSiteBasePath } from "./site_base_path";
import { getCsrfToken } from "./voice_messages";

export function voiceLabel(english: string, french: string): string {
  return (document.documentElement.lang || navigator.language).startsWith("fr")
    ? french
    : english;
}

function mutationHeaders(): Record<string, string> {
  const token = getCsrfToken();
  return token
    ? { "X-CSRF-Token": token, "X-Requested-With": "XMLHttpRequest" }
    : { "X-Requested-With": "XMLHttpRequest" };
}

export async function workspaceTranscript(
  blob: Blob,
  channelId: string,
  signal?: AbortSignal,
): Promise<string> {
  const response = await fetch(
    `${getSiteBasePath()}/plugins/ch.icorete.mattermost-voice-messages/api/v1/channels/${encodeURIComponent(channelId)}/transcribe`,
    {
      method: "POST",
      body: blob,
      credentials: "same-origin",
      headers: mutationHeaders(),
      signal,
    },
  );
  if (response.status === 413 || response.status === 422) {
    throw new Error(
      voiceLabel(
        "This audio could not be transcribed. The workspace service accepts recordings up to five minutes and 16 MiB.",
        "Ce vocal n’a pas pu être transcrit. Le service accepte les enregistrements de cinq minutes et 16 Mio au maximum.",
      ),
    );
  }
  if (!response.ok) {
    throw new Error(
      voiceLabel(
        "The transcript is unavailable. Your audio is still here.",
        "La transcription est indisponible. Votre audio reste accessible.",
      ),
    );
  }
  const { text } = await response.json();
  return typeof text === "string" ? text.trim() : "";
}

export async function readVoiceFile(fileId: string): Promise<Blob> {
  const response = await fetch(
    `${getSiteBasePath()}/api/v4/files/${encodeURIComponent(fileId)}`,
    {
      credentials: "same-origin",
      cache: "no-store",
    },
  );
  if (!response.ok)
    throw new Error(
      voiceLabel("This file is unavailable.", "Ce fichier est indisponible."),
    );
  return response.blob();
}

export async function postTranscriptReply(
  post: Post,
  message: string,
  pendingId: string,
): Promise<void> {
  const response = await fetch(`${getSiteBasePath()}/api/v4/posts`, {
    method: "POST",
    credentials: "same-origin",
    headers: { ...mutationHeaders(), "Content-Type": "application/json" },
    body: JSON.stringify({
      channel_id: post.channel_id,
      root_id: post.root_id || post.id,
      message,
      pending_post_id: pendingId,
    }),
  });
  if (!response.ok)
    throw new Error(
      voiceLabel(
        "Mattermost could not post the reply. Your transcript is still here.",
        "Mattermost n’a pas pu publier la réponse. Votre transcription reste disponible.",
      ),
    );
}
