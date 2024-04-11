package app

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/mail"
	"net/smtp"
	"os"

	"github.com/NikhilSharmaWe/go-vercel-app/models"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func createDB() *gorm.DB {
	db, err := gorm.Open(postgres.Open(os.Getenv("DBADDRESS")), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	return db
}

func (app *Application) alreadyLoggedIn(c echo.Context) bool {
	session := c.Get("session").(*sessions.Session)

	email, ok := session.Values["email"].(string)
	if !ok {
		return false
	}

	if exists, err := app.UserStore.IsExists("email = ?", email); err != nil || !exists {
		return false
	}

	authenticated, ok := session.Values["authenticated"].(bool)
	if ok && authenticated {
		return true
	}

	return false
}

func setSession(c echo.Context, keyValues map[string]any) error {
	session := c.Get("session").(*sessions.Session)
	for k, v := range keyValues {
		session.Values[k] = v
	}

	return session.Save(c.Request(), c.Response())
}

func clearSessionHandler(c echo.Context) error {
	session := c.Get("session").(*sessions.Session)
	session.Options.MaxAge = -1
	return session.Save(c.Request(), c.Response())
}

func (app *Application) getGithubAccessToken(code string) (string, error) {

	requestBodyMap := map[string]string{"client_id": app.GithubClientID, "client_secret": app.GithubClientSecret, "code": code}
	requestJSON, err := json.Marshal(requestBodyMap)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "https://github.com/login/oauth/access_token", bytes.NewBuffer(requestJSON))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	respbody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Represents the response received from Github
	type githubAccessTokenResponse struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		Scope       string `json:"scope"`
	}

	var ghresp githubAccessTokenResponse
	if err := json.Unmarshal(respbody, &ghresp); err != nil {
		return "", err
	}

	return ghresp.AccessToken, nil
}

func (app *Application) createIfNotExists(username, email string, githubAccess bool) error {
	exists, err := app.UserStore.IsExists("email = ?", email)
	if err != nil {
		return err
	}

	if exists {
		return models.ErrUserAlreadyExists
	}

	return app.UserStore.Create(models.UserDBModel{
		Username:     username,
		Email:        email,
		GithubAccess: githubAccess,
	})
}

func (app *Application) sendVerificationCode(username, to, code string) error {
	auth := smtp.PlainAuth("", app.AppEmail, app.AppPassword, "smtp.gmail.com")
	headers := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";"
	message := "Subject: Verification" + "\n" + headers + "\n\n" + fmt.Sprintf("Hello <b>%s</b>, your verification code is: <b>%s<b>", username, code)

	return smtp.SendMail(app.SMTPHost+":"+app.SMTPPort, auth, app.AppEmail, []string{to}, []byte(message))
}

func validMailAddress(address string) bool {
	_, err := mail.ParseAddress(address)
	return err == nil
}

func generateSimpleOTP(length int) (string, error) {
	b := make([]byte, length)
	_, err := io.ReadAtLeast(rand.Reader, b, length)
	if err != nil {
		return "", err
	}

	for i := range b {
		b[i] = byte('0' + int(b[i])%10)
	}
	return string(b), nil
}

func (app *Application) setupEmailVerification(c echo.Context, username, email string) error {
	code, err := generateSimpleOTP(5)
	if err != nil {
		return err
	}

	if err := setSession(c, map[string]any{"verification_code": code}); err != nil {
		return err
	}

	if err := app.sendVerificationCode(username, email, code); err != nil {
		return err
	}

	return nil
}
