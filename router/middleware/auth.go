package middleware

import (
	"crypto/subtle"
	"net/http"

	"github.com/gabek/owncast/config"
	log "github.com/sirupsen/logrus"
)

// RequireAdminAuth wraps a handler requiring HTTP basic auth for it using the given
// the stream key as the password and and a hardcoded "admin" for username.
func RequireAdminAuth(handler http.HandlerFunc) http.HandlerFunc {
	username := "admin"
	password := config.Config.VideoSettings.StreamingKey
	realm := "Owncast Authenticated Request"

	return func(w http.ResponseWriter, r *http.Request) {
		// The following line is kind of a work around.
		// If you want HTTP Basic Auth + Cors it requires _explicit_ origins to be provided in the
		// Access-Control-Allow-Origin header.  So we just pull out the origin header and specify it.
		// If we want to lock down admin APIs to not be CORS accessible for anywhere, this is where we would do that.
		w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")

		// For request needing CORS, send a 200.
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		user, pass, ok := r.BasicAuth()

		// Failed
		if !ok || subtle.ConstantTimeCompare([]byte(user), []byte(username)) != 1 || subtle.ConstantTimeCompare([]byte(pass), []byte(password)) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			log.Warnln("Failed authentication for", r.URL.Path, "from", r.RemoteAddr, r.UserAgent())
			return
		}

		// Success
		log.Traceln("Authenticated request OK for", r.URL.Path, "from", r.RemoteAddr, r.UserAgent())
		handler(w, r)
	}
}
