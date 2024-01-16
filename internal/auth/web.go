package auth

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strings"
)

type _HLContextKey int

const (
	Ctx__AuthenticatedUser _HLContextKey = iota
	Ctx__AuthLevel
)

func AuthenticateHTTP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		// http basic auth, but in a custom header so we can logout via UI
		// sorry for the func() nonsense, I'm a rustacean
		creds := func() string {
			if creds := r.Header.Get("HL-Session"); creds != "" {
				return creds
			} else if cookie, _ := r.Cookie("HL-Session"); cookie != nil {
				return cookie.Value
			}
			return ""
		}()

		username, password := func() (string, string) {
			if creds != "" {
				b64 := strings.TrimPrefix(creds, "Basic ")
				if creds == b64 {
					// unknown auth scheme
					return "", ""
				}
				b, err := base64.StdEncoding.DecodeString(b64)
				if err != nil {
					log.Println(fmt.Errorf("error decoding credentials: %w", err))
					return "", ""
				}
				split := strings.Split(string(b), ":")
				if len(split) != 2 {
					log.Println("invalid count of auth parts")
				}
				return split[0], split[1]
			}
			// no auth provided
			return "", ""
		}()

		if username == "" {
			// unauthenticated request or malformed credentials
			r = r.WithContext(context.WithValue(r.Context(), Ctx__AuthenticatedUser, ""))
			r = r.WithContext(context.WithValue(r.Context(), Ctx__AuthLevel, AuthLevel_LoggedOut))
			next.ServeHTTP(rw, r)
			return
		}

		ok, err := AuthenticateUser(username, password)
		if err != nil {
			panic(err)
		}

		if ok {
			r = r.WithContext(context.WithValue(r.Context(), Ctx__AuthenticatedUser, username))
			// todo: get >member auth level from db
			r = r.WithContext(context.WithValue(r.Context(), Ctx__AuthLevel, AuthLevel_Member))
		} else {
			r = r.WithContext(context.WithValue(r.Context(), Ctx__AuthenticatedUser, ""))
			r = r.WithContext(context.WithValue(r.Context(), Ctx__AuthLevel, AuthLevel_LoggedOut))
		}

		// todo: return 401 on api
		next.ServeHTTP(rw, r)
	})
}
