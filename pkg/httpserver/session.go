package httpserver

import "net/http"

const sessionCookieName = "replace-this-session-with-a-cookie"

// func withSessionBootstrap(sessionToken string) http.HandlerFunc {
// 	return withSessionCookie(sessionToken, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
// 		w.WriteHeader(http.StatusNoContent)
// 	})).ServeHTTP
// }

func WithSessionCookie(sessionName string, sessionToken string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:     sessionName,
			Value:    sessionToken,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})
		next.ServeHTTP(w, r)
	})
}

func requireSession(r *http.Request, sessionToken string) error {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return err
	}
	if cookie.Value != sessionToken {
		return http.ErrNoCookie
	}
	return nil
}
