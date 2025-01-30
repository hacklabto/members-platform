package ui

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"members-platform/internal/auth"
	"members-platform/internal/db"
	"members-platform/internal/mailer"
	"members-platform/static"
	"net/http"
	"net/mail"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/adtac/go-akismet/akismet"
)

func Router() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(auth.AuthenticateHTTP)

	// contact form
	r.Post("/wp-json/contact-form-7/v1/contact-forms/844/feedback", func(rw http.ResponseWriter, r *http.Request) {
		email := mailer.ContactFormData{
			UserName: r.FormValue("your-name"),
			UserEmail: r.FormValue("your-email"),
			Subject: r.FormValue("your-subject"),
			Message: r.FormValue("your-message"),

			Border: "#ffb900",
		}

		// validate form
		var validateErr error
		if email.UserName == "" {
			validateErr = fmt.Errorf("name is missing")
		}
		if validateErr == nil {
			if email.UserEmail == "" {
				validateErr = fmt.Errorf("email is missing")
			} else {
				_, validateErr = mail.ParseAddress(email.UserEmail)
			}
		}
		if validateErr == nil {
			cityQuiz := r.FormValue("city-quiz")
			// :)
			if cityQuiz == "" || strings.ToLower(cityQuiz) != os.Getenv("CONTACT_FORM_QUIZ") {
				validateErr = fmt.Errorf("Your answer is not correct")
			}
		}

		akismetKey := os.Getenv("AKISMET_KEY")
		if validateErr == nil && akismetKey != "" {
			// antispam
			// todo: make this work
			// is a pain to get the source ip since http is forwarded from the old vps
			// but that's a problem for later
			isSpam, err := akismet.Check(&akismet.Comment{
				Blog: "https://hacklab.to",
				UserIP: "8.8.8.8",
				UserAgent: "...",
				CommentType: "contact-form",
				CommentAuthor: email.UserName,
				CommentAuthorEmail: email.UserEmail,
				CommentContent: email.Message,
			}, akismetKey)

			if err != nil {
				log.Println(err)
				validateErr = fmt.Errorf("server error, please try again or contact operations@ (1)")
			}

			if isSpam {
				log.Printf("spam check failed for contact form message from %s (%s)", email.UserName, email.UserEmail)
				validateErr = fmt.Errorf("Validation failed")
			}
		}

		if validateErr != nil {
			if !strings.Contains(validateErr.Error(), "server error") {
				email.Error = fmt.Sprintf("Validation errors occurred. Please confirm the fields and submit it again. (%s)", validateErr)
			} else {
				email.Error = validateErr.Error()
			}

			if err := Page(rw, "contact", email); err != nil {
				log.Println(err)
			}
			return
		}

		content, err := mailer.ExecuteTemplateWithoutFooter("contact-form", email)
		if err != nil {
			log.Println(err)
			email.Error = "server error, please try again or contact operations@ (2)"
			if err := Page(rw, "contact", email); err != nil {
				log.Println(err)
			}
			return
		}

		err = mailer.DoSendEmail("info@hacklab.to", content)
		if err != nil {
			log.Println(err)
			email.Error = "server error, please try again or contact operations@ (3)"
			if err := Page(rw, "contact", email); err != nil {
				log.Println(err)
			}
			return
		}

		email.Border = "#00a0d2"
		email.Error = "Your message was sent successfully!"
		if err := Page(rw, "contact", email); err != nil {
			log.Println(err)
		}
	})

	registerStaticRoutes(r)
	registerStaticPages(r)
	registerPasswdRoutes(r)

	// todo: this needs to be POST with CSRF
	r.Get("/logout/", func(rw http.ResponseWriter, r *http.Request) {
		http.SetCookie(rw, &http.Cookie{
			Name:    "HL-Session",
			Value:   "",
			Path:    "/",
			Expires: time.Now().UTC(),
		})
		http.Redirect(rw, r, "/", http.StatusTemporaryRedirect)
	})

	r.Post("/login/", func(rw http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			log.Println(err)
		}
		username := r.Form.Get("username")
		password := r.Form.Get("password")
		ok, err := auth.AuthenticateUser(username, password)
		if err != nil {
			MaybeHtmxComponent(rw, r, "login", shit{Error: err.Error()})
			return
		}
		if !ok {
			MaybeHtmxComponent(rw, r, "login", shit{Error: "Invalid username or password"})
			return
		}

		http.SetCookie(rw, &http.Cookie{
			Name:     "HL-Session",
			Value:    "Basic " + base64.StdEncoding.EncodeToString([]byte(strings.Join([]string{username, password}, ":"))),
			HttpOnly: true,
			Path:     "/",
		})

		// htmx this is stupid
		if r.Header.Get("HX-Request") == "true" {
			rw.Header().Set("HX-Location", "/")
		} else {
			http.Redirect(rw, r, "/", http.StatusFound)
		}
	})

	return r
}

func registerStaticRoutes(r chi.Router) {
	r.Get("/favicon.ico", func(rw http.ResponseWriter, r *http.Request) {
		r.URL.Path = "/static/favicon.ico"
		static.Server().ServeHTTP(rw, r)
	})

	r.Get("/robots.txt", func(rw http.ResponseWriter, _ *http.Request) {
		rw.Write([]byte("User-Agent: *\nDisallow: *\nDisallow: /ban-me/admin.php"))
	})

	r.Get("/static/*", static.Server())
}

func registerStaticPages(r chi.Router) {
	pathPages := map[string]string{
		"/":       "index",
		"/login/": "login",
		"/apply/": "apply",
	}

	for k, v := range pathPages {
		// ???
		p := k
		q := v
		r.Get(p, func(rw http.ResponseWriter, r *http.Request) {
			if err := PageWithShell(r.Context(), rw, q, shit{}); err != nil {
				log.Println(err)
			}
		})
	}
}

func registerPasswdRoutes(r chi.Router) {
	r.Get("/passwd/", func(rw http.ResponseWriter, r *http.Request) {
		if err := PageWithShell(r.Context(), rw, "passwd", Passwd{Token: r.URL.Query().Get("token")}); err != nil {
			log.Println(err)
		}
	})

	r.Post("/passwd/", func(rw http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			log.Println(err)
		}

		errReply := Passwd{Token: r.Form.Get("token")}

		typ := r.Form.Get("type")
		switch typ {
		case "change":
			newPassword := r.Form.Get("password")
			if len(newPassword) < 12 { // arbitrary
				errReply.Error = "password is too short (must be 12 characters)"
				MaybeHtmxComponent(rw, r, "passwd", errReply)
				return
			}
			if newPassword != r.Form.Get("confirm") {
				errReply.Error = "passwords do not match"
				MaybeHtmxComponent(rw, r, "passwd", errReply)
				return
			}
			udn := fmt.Sprintf("uid=%s,ou=people,dc=hacklab,dc=to", r.Context().Value(auth.Ctx__AuthenticatedUser).(string))
			err := auth.DoChangePassword(udn, r.Form.Get("current"), udn, newPassword)
			if err != nil {
				errReply.Error = err.Error()
				MaybeHtmxComponent(rw, r, "passwd", errReply)
				return
			}
			MaybeHtmxComponent(rw, r, "confirmation", Confirmation{
				Title:   "Change your password",
				Message: "Your password has been successfully changed. Please log in again.",
			})
		case "reset":
			username := r.Form.Get("username")
			if username == "" {
				errReply.Error = "username cannot be null"
				MaybeHtmxComponent(rw, r, "passwd", errReply)
				return
			}

			token, err := auth.CreateResetToken(username)
			if err != nil {
				errReply.Error = err.Error()
				MaybeHtmxComponent(rw, r, "passwd", errReply)
				return
			}

			email, err := auth.GetEmailFromUsername(
				"cn=password_self_service,ou=services,dc=hacklab,dc=to",
				os.Getenv("LDAP_SELFSERVICE_PASSWORD"),
				username,
			)

			if err == auth.ErrInvalidGroup {
				err = auth.SendResetRestrictedEmail(email, username)
				if err != nil {
					errReply.Error = err.Error()
					MaybeHtmxComponent(rw, r, "passwd", errReply)
					return
				}
				MaybeHtmxComponent(rw, r, "confirmation", Confirmation{
					Title:   "Reset your password",
					Message: "A confirmation email has been sent to the address associated with your account.",
				})
				return
			}

			if err != nil {
				errReply.Error = err.Error()
				MaybeHtmxComponent(rw, r, "passwd", errReply)
				return
			}
			err = auth.SendResetEmail(email, username, token)
			if err != nil {
				errReply.Error = err.Error()
				MaybeHtmxComponent(rw, r, "passwd", errReply)
				return
			}
			MaybeHtmxComponent(rw, r, "confirmation", Confirmation{
				Title:   "Reset your password",
				Message: "A confirmation email has been sent to the address associated with your account.",
			})
		case "do-reset":
			token := r.Form.Get("token")
			username, ok := auth.ValidateResetToken(token)
			if !ok {
				errReply.Error = "invalid token"
				MaybeHtmxComponent(rw, r, "passwd", errReply)
				return
			}
			newPassword := r.Form.Get("password")
			if len(newPassword) < 12 { // arbitrary
				errReply.Error = "password is too short (must be 12 characters)"
				MaybeHtmxComponent(rw, r, "passwd", errReply)
				return
			}
			if newPassword != r.Form.Get("confirm") {
				errReply.Error = "passwords do not match"
				MaybeHtmxComponent(rw, r, "passwd", errReply)
				return
			}
			email, err := auth.GetEmailFromUsername(
				"cn=password_self_service,ou=services,dc=hacklab,dc=to",
				os.Getenv("LDAP_SELFSERVICE_PASSWORD"),
				username,
			)
			if err != nil {
				errReply.Error = err.Error()
				MaybeHtmxComponent(rw, r, "passwd", errReply)
				return
			}
			err = auth.SendConfirmationEmail(email, username)
			if err != nil {
				errReply.Error = err.Error()
				MaybeHtmxComponent(rw, r, "passwd", errReply)
				return
			}
			err = auth.DoChangePassword(
				"cn=password_self_service,ou=services,dc=hacklab,dc=to",
				os.Getenv("LDAP_SELFSERVICE_PASSWORD"),
				"uid="+username+",ou=people,dc=hacklab,dc=to",
				newPassword,
			)
			if err != nil {
				errReply.Error = err.Error()
				MaybeHtmxComponent(rw, r, "passwd", errReply)
				return
			}
			if err := db.Redis.Del(context.Background(), "reset-token:"+token).Err(); err != nil {
				errReply.Error = err.Error()
				MaybeHtmxComponent(rw, r, "passwd", errReply)
				return
			}
			MaybeHtmxComponent(rw, r, "confirmation", Confirmation{
				Title:   "Reset your password",
				Message: "Your password has been successfully reset. <a class=\"text-blue-600 hover:text-blue-800\" href=\"/login/\">Log back in?</a>",
			})
		default:
			errReply.Error = "Unknown POST form data"
			MaybeHtmxComponent(rw, r, "passwd", errReply)
		}
	})
}

type shit struct {
	Error string
}
