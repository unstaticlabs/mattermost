package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost/server/public/model"
)

// Leave room for multipart headers inside the speech service's 16 MiB limit.
const speechAudioLimit = (16 << 20) - 4096

var speechHTTP = &http.Client{
	Timeout:       20 * time.Second,
	CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
}

func validSpeechEndpoint(raw string) bool {
	if raw == "" {
		return true
	}
	u, err := url.Parse(raw)
	return err == nil && (u.Scheme == "https" || u.Scheme == "http") && u.Hostname() != "" &&
		u.User == nil && u.RawQuery == "" && u.Fragment == ""
}

// The browser supplies the recording. Reading an existing attachment uses the
// caller's ordinary Mattermost file endpoint; the plugin never reads its store.
func (p *Plugin) handleWorkspaceTranscript(w http.ResponseWriter, r *http.Request) {
	cfg := p.getConfiguration()
	endpoint := strings.TrimSpace(cfg.WorkspaceSpeechURL)
	if endpoint == "" || !cfg.voiceMessagesEnabled() {
		http.Error(w, "Transcription is unavailable.", http.StatusServiceUnavailable)
		return
	}
	channelID := mux.Vars(r)["channel_id"]
	userID := r.Header.Get("Mattermost-User-ID")
	if !model.IsValidId(channelID) || userID == "" || !p.API.HasPermissionToChannel(userID, channelID, model.PermissionReadChannel) {
		http.Error(w, "You cannot access this channel.", http.StatusForbidden)
		return
	}
	if _, err := p.API.GetChannelMember(channelID, userID); err != nil {
		http.Error(w, "You cannot access this channel.", http.StatusForbidden)
		return
	}
	mime, ext, ok := detectVoiceAudio(r.Header.Get("Content-Type"))
	if !ok || mime == "" {
		http.Error(w, "Unsupported audio format.", http.StatusUnsupportedMediaType)
		return
	}
	var form bytes.Buffer
	mw := multipart.NewWriter(&form)
	part, err := mw.CreateFormFile("file", "voice-note"+ext)
	if err != nil {
		http.Error(w, "Could not prepare the audio.", http.StatusInternalServerError)
		return
	}
	size, err := io.Copy(part, http.MaxBytesReader(w, r.Body, speechAudioLimit))
	if err != nil || size == 0 {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			http.Error(w, "This recording is too large to transcribe.", http.StatusRequestEntityTooLarge)
		} else {
			http.Error(w, "No readable audio was received.", http.StatusBadRequest)
		}
		return
	}
	if err := mw.Close(); err != nil {
		http.Error(w, "Could not prepare the audio.", http.StatusInternalServerError)
		return
	}
	upstream, err := http.NewRequestWithContext(r.Context(), http.MethodPost, endpoint, &form)
	if err != nil {
		http.Error(w, "Transcription is unavailable.", http.StatusServiceUnavailable)
		return
	}
	upstream.Header.Set("Content-Type", mw.FormDataContentType())
	resp, err := speechHTTP.Do(upstream)
	if err != nil {
		http.Error(w, "Transcription is unavailable.", http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusRequestEntityTooLarge || resp.StatusCode == http.StatusUnprocessableEntity {
		http.Error(w, "The speech service could not process this recording.", http.StatusUnprocessableEntity)
		return
	}
	var transcript struct {
		Text string `json:"text"`
	}
	if resp.StatusCode != http.StatusOK || json.NewDecoder(io.LimitReader(resp.Body, 256<<10)).Decode(&transcript) != nil {
		http.Error(w, "Transcription is unavailable.", http.StatusServiceUnavailable)
		return
	}
	transcript.Text = normalizeVoiceTranscript(transcript.Text)
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(transcript)
}
