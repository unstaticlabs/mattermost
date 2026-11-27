package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin/plugintest"
	"github.com/stretchr/testify/require"
)

func TestWorkspaceSpeechUsesOnlyAudioAndThePrivateChannelMembership(t *testing.T) {
	audio := []byte("invented audio fixture")
	var calls atomic.Int32
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		require.Equal(t, "/v1/audio/transcriptions", r.URL.Path)
		require.NoError(t, r.ParseMultipartForm(1<<20))
		require.Len(t, r.MultipartForm.Value, 0)
		require.Len(t, r.MultipartForm.File, 1)
		f, h, err := r.FormFile("file")
		require.NoError(t, err)
		defer f.Close()
		require.Equal(t, "voice-note.webm", h.Filename)
		got, err := io.ReadAll(f)
		require.NoError(t, err)
		require.Equal(t, audio, got)
		_, _ = io.WriteString(w, `{"text":" Le micro fonctionne. "}`)
	}))
	defer backend.Close()
	api := &plugintest.API{}
	p := &Plugin{}
	p.SetAPI(api)
	p.setConfiguration(&configuration{WorkspaceSpeechURL: backend.URL + "/v1/audio/transcriptions"})
	router := p.initRouter()
	channel, member := model.NewId(), model.NewId()
	api.On("HasPermissionToChannel", member, channel, model.PermissionReadChannel).Return(true).Once()
	api.On("GetChannelMember", channel, member).Return(&model.ChannelMember{}, (*model.AppError)(nil)).Once()
	send := func() *httptest.ResponseRecorder {
		r := httptest.NewRequest(http.MethodPost, "/api/v1/channels/"+channel+"/transcribe", bytes.NewReader(audio))
		r.Header.Set("Mattermost-User-ID", member)
		r.Header.Set("Content-Type", "audio/webm;codecs=opus")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)
		return w
	}
	w := send()
	require.Equal(t, http.StatusOK, w.Code)
	require.JSONEq(t, `{"text":"Le micro fonctionne."}`, w.Body.String())
	require.Equal(t, "no-store", w.Header().Get("Cache-Control"))
	api.On("HasPermissionToChannel", member, channel, model.PermissionReadChannel).Return(false).Once()
	require.Equal(t, http.StatusForbidden, send().Code)
	require.EqualValues(t, 1, calls.Load())
	api.AssertExpectations(t)
}

func TestWorkspaceSpeechOutageDoesNotDiscloseResponseOrForwardRedirect(t *testing.T) {
	var destinationCalls atomic.Int32
	dest := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { destinationCalls.Add(1) }))
	defer dest.Close()
	for _, status := range []int{http.StatusServiceUnavailable, http.StatusTemporaryRedirect} {
		backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Location", dest.URL)
			w.WriteHeader(status)
			_, _ = io.WriteString(w, "private fixture diagnostic")
		}))
		api := &plugintest.API{}
		p := &Plugin{}
		p.SetAPI(api)
		p.setConfiguration(&configuration{WorkspaceSpeechURL: backend.URL})
		channel, member := model.NewId(), model.NewId()
		api.On("HasPermissionToChannel", member, channel, model.PermissionReadChannel).Return(true)
		api.On("GetChannelMember", channel, member).Return(&model.ChannelMember{}, (*model.AppError)(nil))
		r := httptest.NewRequest(http.MethodPost, "/api/v1/channels/"+channel+"/transcribe", bytes.NewBufferString("fixture"))
		r.Header.Set("Mattermost-User-ID", member)
		r.Header.Set("Content-Type", "audio/ogg")
		w := httptest.NewRecorder()
		p.initRouter().ServeHTTP(w, r)
		backend.Close()
		require.Equal(t, http.StatusServiceUnavailable, w.Code)
		require.NotContains(t, w.Body.String(), "fixture")
	}
	require.Zero(t, destinationCalls.Load())
}
