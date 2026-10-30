package ui

import (
	"bytes"
	"crypto/sha1"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"members-platform/internal/auth"
	"members-platform/internal/db"
	"members-platform/internal/mailer"
	"net/http"
	"net/mail"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

// ported from the old php apply app (apply.hacklab.to)

const maxPictureSize = 10 << 20

var usernameRegex = regexp.MustCompile("^[a-z0-9]+$")
var profileHashRegex = regexp.MustCompile("^[A-Fa-z0-9]{40}$")

// GIF, JPEG, PNG, BMP
var pictureExtensions = map[string]string{
	"image/gif":  "gif",
	"image/jpeg": "jpeg",
	"image/png":  "png",
	"image/bmp":  "bmp",
}

type ApplyForm struct {
	Error       string
	Values      map[string]string
	FieldErrors map[string]string
}

func newApplyForm() ApplyForm {
	return ApplyForm{
		Values:      map[string]string{"token_type": "fob"},
		FieldErrors: map[string]string{},
	}
}

func registerApplyRoutes(r chi.Router) {
	r.Get("/apply/", func(rw http.ResponseWriter, r *http.Request) {
		if err := PageWithShell(r.Context(), rw, "apply", newApplyForm()); err != nil {
			log.Println(err)
		}
	})

	r.Post("/apply/", func(rw http.ResponseWriter, r *http.Request) {
		form := newApplyForm()

		if err := r.ParseMultipartForm(maxPictureSize); err != nil && !errors.Is(err, http.ErrNotMultipart) {
			log.Println(err)
			form.Error = "internal server error(1)"
			MaybeHtmxComponent(rw, r, "apply", form)
			return
		}

		level := r.Context().Value(auth.Ctx__AuthLevel).(auth.AuthLevel)

		switch r.Form.Get("action") {
		case "login":
			if level > auth.AuthLevel_LoggedOut {
				form.Error = "auth failed(1)"
				MaybeHtmxComponent(rw, r, "apply", form)
				return
			}

			checkPassword := os.Getenv("APPLICANT_PASSWORD")
			if checkPassword == "" {
				form.Error = "internal server error(2)"
				MaybeHtmxComponent(rw, r, "apply", form)
				return
			}

			password := r.Form.Get("password")

			if subtle.ConstantTimeCompare([]byte(checkPassword), []byte(password)) == 0 {
				form.Error = "auth failed(2)"
				MaybeHtmxComponent(rw, r, "apply", form)
				return
			}

			http.SetCookie(rw, &http.Cookie{
				Name:     "HL-Session",
				Value:    "Basic " + base64.StdEncoding.EncodeToString([]byte(strings.Join([]string{"applicant", password}, ":"))),
				HttpOnly: true,
				Path:     "/",
			})

			if r.Header.Get("HX-Request") == "true" {
				rw.Header().Set("HX-Location", "/apply/")
			} else {
				http.Redirect(rw, r, "/apply/", http.StatusFound)
			}
		case "submit":
			if level != auth.AuthLevel_Applicant {
				http.Redirect(rw, r, "/apply/", http.StatusFound)
				return
			}
			handleApplication(rw, r, form)
		default:
			form.Error = "internal server error(0)"
			MaybeHtmxComponent(rw, r, "apply", form)
		}
	})

	r.Get("/apply/profile/{hash}", func(rw http.ResponseWriter, r *http.Request) {
		hash := chi.URLParam(r, "hash")
		if !profileHashRegex.MatchString(hash) {
			http.NotFound(rw, r)
			return
		}
		picture, err := db.GetApplicantPicture(r.Context(), strings.ToLower(hash))
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				log.Println(err)
			}
			http.NotFound(rw, r)
			return
		}
		dir := os.Getenv("APPLY_PROFILES_DIR")
		if dir == "" || picture == "" {
			http.NotFound(rw, r)
			return
		}
		http.ServeFile(rw, r, filepath.Join(dir, filepath.Base(picture)))
	})
}

func handleApplication(rw http.ResponseWriter, r *http.Request, form ApplyForm) {
	for _, k := range []string{
		"name", "username", "nickname", "contact_email", "list_email", "bio_reason",
		"face_url", "token_type", "sponsor", "second_sponsor", "twitter", "facebook",
		"website", "preferred_gender_pronoun", "heard_from",
	} {
		form.Values[k] = strings.TrimSpace(r.Form.Get(k))
	}
	v := form.Values

	notBlank := func(k string) bool {
		if v[k] == "" {
			form.FieldErrors[k] = "This value should not be blank."
			return false
		}
		return true
	}
	validEmail := func(k string) {
		if a, err := mail.ParseAddress(v[k]); err != nil || a.Address != v[k] {
			form.FieldErrors[k] = "This value is not a valid email address."
		}
	}
	validURL := func(k string) {
		if v[k] == "" {
			return
		}
		u, err := url.Parse(v[k])
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
			form.FieldErrors[k] = "This value is not a valid URL."
		}
	}

	notBlank("name")
	if notBlank("username") && !usernameRegex.MatchString(v["username"]) {
		form.FieldErrors["username"] = "Username can only be alphanumeric characters, all lower case."
	}
	notBlank("nickname")
	if notBlank("contact_email") {
		validEmail("contact_email")
	}
	if notBlank("list_email") {
		validEmail("list_email")
	}
	notBlank("bio_reason")
	notBlank("sponsor")
	notBlank("second_sponsor")
	validURL("face_url")
	validURL("facebook")
	validURL("website")
	if v["token_type"] != "fob" && v["token_type"] != "card" {
		form.FieldErrors["token_type"] = "The value you selected is not a valid choice."
	}

	if len(form.FieldErrors) > 0 {
		MaybeHtmxComponent(rw, r, "apply", form)
		return
	}

	username := strings.ToLower(v["username"])

	faceFile, _, err := r.FormFile("face_file")
	if err != nil && !errors.Is(err, http.ErrMissingFile) {
		log.Println(err)
		form.FieldErrors["face_file"] = "The file could not be uploaded."
		MaybeHtmxComponent(rw, r, "apply", form)
		return
	}
	if faceFile != nil {
		defer faceFile.Close()
	}

	if faceFile == nil && v["face_url"] == "" {
		form.Error = "You must specify a photo URL or upload one yourself!"
		MaybeHtmxComponent(rw, r, "apply", form)
		return
	}

	var picture string

	if faceFile != nil {
		picture, err = savePicture(faceFile, username)
		if err != nil {
			log.Println(err)
			form.FieldErrors["face_file"] = "Invalid picture filetype (GIF, JPEG, PNG, BMP)"
			MaybeHtmxComponent(rw, r, "apply", form)
			return
		}
	}

	if v["face_url"] != "" {
		picture, err = fetchPicture(v["face_url"], username)
		if errors.Is(err, errInvalidPicture) {
			form.FieldErrors["face_url"] = "Invalid picture filetype (GIF, JPEG, PNG, BMP)"
			MaybeHtmxComponent(rw, r, "apply", form)
			return
		} else if err != nil {
			log.Println(err)
			form.FieldErrors["face_url"] = "Unable to retrieve remote photo from URL."
			MaybeHtmxComponent(rw, r, "apply", form)
			return
		}
	}

	profileHash := sha1.Sum([]byte(username))
	profileHashHex := hex.EncodeToString(profileHash[:])

	err = db.InsertApplicant(r.Context(), db.Applicant{
		Name:                   v["name"],
		Username:               username,
		Nickname:               v["nickname"],
		ContactEmail:           v["contact_email"],
		ListEmail:              v["list_email"],
		BioReason:              v["bio_reason"],
		Sponsor:                nullString(v["sponsor"]),
		SecondSponsor:          nullString(v["second_sponsor"]),
		Picture:                nullString(picture),
		TokenType:              v["token_type"],
		Twitter:                nullString(v["twitter"]),
		Facebook:               nullString(v["facebook"]),
		HeardFrom:              nullString(v["heard_from"]),
		ProfileHash:            nullString(profileHashHex),
		Website:                nullString(v["website"]),
		PreferredGenderPronoun: nullString(v["preferred_gender_pronoun"]),
	})
	if errors.Is(err, db.ErrDuplicateApplicant) {
		form.Error = "Nickname is already in use"
		MaybeHtmxComponent(rw, r, "apply", form)
		return
	} else if err != nil {
		log.Println(err)
		form.Error = "internal server error(3)"
		MaybeHtmxComponent(rw, r, "apply", form)
		return
	}

	var notify []string
	for _, e := range strings.Split(os.Getenv("APPLY_NOTIFY_EMAILS"), ",") {
		if e = strings.TrimSpace(e); e != "" {
			notify = append(notify, e)
		}
	}
	oneLine := strings.NewReplacer("\r", "", "\n", " ")
	content, err := mailer.ExecuteTemplate("new-member", mailer.NewApplicantData{
		ToAddresses:            strings.Join(notify, ", "),
		Name:                   oneLine.Replace(v["name"]),
		Nickname:               oneLine.Replace(v["nickname"]),
		ContactEmail:           v["contact_email"],
		Website:                v["website"],
		ProfileURL:             "https://members.hacklab.to/apply/profile/" + profileHashHex,
		Sponsor:                v["sponsor"],
		SecondSponsor:          v["second_sponsor"],
		Twitter:                v["twitter"],
		Facebook:               v["facebook"],
		BioReason:              v["bio_reason"],
		HeardFrom:              v["heard_from"],
		PreferredGenderPronoun: v["preferred_gender_pronoun"],
	})
	if err == nil {
		if len(notify) == 0 {
			err = fmt.Errorf("missing APPLY_NOTIFY_EMAILS in environment")
		} else {
			err = mailer.DoSendEmailToMany("applications@hacklab.to", notify, content)
		}
	}
	if err != nil {
		// the application is already saved at this point
		log.Printf("failed to send application email for %s: %v", username, err)
		form.Error = "Your application was saved, but we couldn't notify the membership. Please contact operations@hacklab.to."
		MaybeHtmxComponent(rw, r, "apply", form)
		return
	}

	// reset for the next applicant
	http.SetCookie(rw, &http.Cookie{
		Name:    "HL-Session",
		Value:   "",
		Path:    "/",
		Expires: time.Now().UTC(),
	})
	http.Redirect(rw, r, "/apply/success/", http.StatusSeeOther)
}

var errInvalidPicture = errors.New("invalid picture filetype")

// checks the picture type and writes it to APPLY_PROFILES_DIR, returning the filename
func savePicture(src io.Reader, username string) (string, error) {
	dir := os.Getenv("APPLY_PROFILES_DIR")
	if dir == "" {
		return "", fmt.Errorf("missing APPLY_PROFILES_DIR in environment")
	}

	data, err := io.ReadAll(io.LimitReader(src, maxPictureSize+1))
	if err != nil {
		return "", fmt.Errorf("read picture: %w", err)
	}
	if len(data) > maxPictureSize {
		return "", errInvalidPicture
	}

	ext, ok := pictureExtensions[http.DetectContentType(data)]
	if !ok {
		return "", errInvalidPicture
	}

	picture := username + "." + ext
	if err := os.WriteFile(filepath.Join(dir, picture), data, 0o644); err != nil {
		return "", fmt.Errorf("write picture: %w", err)
	}
	return picture, nil
}

func fetchPicture(u, username string) (string, error) {
	client := http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(u)
	if err != nil {
		return "", fmt.Errorf("fetch picture: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fetch picture: status %d", resp.StatusCode)
	}

	// buffer so a read error isn't reported as a bad filetype
	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, io.LimitReader(resp.Body, maxPictureSize+1)); err != nil {
		return "", fmt.Errorf("fetch picture: %w", err)
	}
	return savePicture(buf, username)
}

func nullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}
