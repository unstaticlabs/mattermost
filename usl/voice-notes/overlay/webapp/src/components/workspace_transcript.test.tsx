import type { Post } from "@mattermost/types/posts";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import React from "react";
import { afterEach, expect, it, vi } from "vitest";
import { WorkspaceTranscript } from "./workspace_transcript";

const post = {
  id: "voice-post",
  channel_id: "private-channel",
  root_id: "thread-root",
} as Post;
afterEach(() => vi.unstubAllGlobals());

it("reads the file as the member and shares the transcript in the original thread", async () => {
  const fetch = vi
    .fn()
    .mockResolvedValueOnce(
      new Response("qa", { headers: { "Content-Type": "audio/webm" } }),
    )
    .mockResolvedValueOnce(
      Response.json({ text: "Please test the microphone." }),
    )
    .mockResolvedValueOnce(Response.json({}));
  vi.stubGlobal("fetch", fetch);
  render(<WorkspaceTranscript post={post} fileId="voice-file" />);
  fireEvent.click(screen.getByText("Get transcript"));
  await screen.findByText("Please test the microphone.");
  expect(fetch.mock.calls[0][0]).toBe("/api/v4/files/voice-file");
  expect(fetch.mock.calls[0][1].credentials).toBe("same-origin");
  expect(fetch.mock.calls[1][0]).toContain(
    "/channels/private-channel/transcribe",
  );
  fireEvent.click(screen.getByText("Share transcript in thread"));
  await screen.findByText("Transcript shared in the thread.");
  const [path, options] = fetch.mock.calls[2];
  expect(path).toBe("/api/v4/posts");
  expect(JSON.parse(options.body)).toMatchObject({
    channel_id: "private-channel",
    root_id: "thread-root",
    message: "Please test the microphone.",
  });
  expect(options.credentials).toBe("same-origin");
});

it("keeps the note available on a speech outage and retries without uploading a second file", async () => {
  const audio = () =>
    new Response("qa", { headers: { "Content-Type": "audio/ogg" } });
  const fetch = vi
    .fn()
    .mockResolvedValueOnce(audio())
    .mockResolvedValueOnce(new Response("", { status: 503 }))
    .mockResolvedValueOnce(audio())
    .mockResolvedValueOnce(Response.json({ text: "Le micro fonctionne." }));
  vi.stubGlobal("fetch", fetch);
  render(<WorkspaceTranscript post={post} fileId="voice-file" />);
  fireEvent.click(screen.getByText("Get transcript"));
  await screen.findByText(
    "The transcript is unavailable. Your audio is still here.",
  );
  fireEvent.click(screen.getByText("Get transcript"));
  await screen.findByText("Le micro fonctionne.");
  expect(
    fetch.mock.calls.filter(([path]) => path === "/api/v4/files/voice-file"),
  ).toHaveLength(2);
  expect(fetch.mock.calls).toHaveLength(4);
});

it("honours a native file read refusal without requesting speech", async () => {
  const fetch = vi.fn().mockResolvedValue(new Response("", { status: 403 }));
  vi.stubGlobal("fetch", fetch);
  render(<WorkspaceTranscript post={post} fileId="voice-file" />);
  fireEvent.click(screen.getByText("Get transcript"));
  await waitFor(() =>
    expect(screen.getByText("This file is unavailable.")).toBeTruthy(),
  );
  expect(fetch).toHaveBeenCalledTimes(1);
});
