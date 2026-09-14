package server

import (
	"github.com/gin-gonic/contrib/sessions"
	gsessions "github.com/gorilla/sessions"
	"net/http"
)

const sessionMaxAge = 30 * 24 * 60 * 60

// cookieStore keeps sessions in a signed cookie. Unlike the store from gin-gonic/contrib,
// it sets SameSite, so the browser doesn't send the session with cross-site requests.
type cookieStore struct {
	*gsessions.CookieStore
}

func newCookieStore(keyPairs ...[]byte) sessions.Store {
	store := &cookieStore{CookieStore: gsessions.NewCookieStore(keyPairs...)}
	store.Options(sessions.Options{Path: "/", MaxAge: sessionMaxAge, HttpOnly: true})
	return store
}

func (s *cookieStore) Options(options sessions.Options) {
	s.CookieStore.Options = &gsessions.Options{
		Path:     options.Path,
		Domain:   options.Domain,
		MaxAge:   options.MaxAge,
		Secure:   options.Secure,
		HttpOnly: options.HttpOnly,
		SameSite: http.SameSiteLaxMode,
	}
}
