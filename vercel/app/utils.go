package app

import (
	"bytes"
	"crypto/rand"
	math_rand "math/rand"
	"time"

	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/mail"
	"net/smtp"
	"os"
	"sync"

	"github.com/NikhilSharmaWe/go-vercel-app/vercel/internal"
	"github.com/NikhilSharmaWe/go-vercel-app/vercel/models"
	"github.com/NikhilSharmaWe/go-vercel-app/vercel/store"
	"github.com/google/go-github/github"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"

	amqp "github.com/rabbitmq/amqp091-go"

	"gorm.io/driver/postgres"

	"gorm.io/gorm"
)

// For rabbitMQ name for queue and routing key are the same in all 4 cases due to direct type exchange

// The application requires 2 exchanges: "upload" and "deploy"
// And 4 queues

// For upload exchange:
// upload-request: request to upload server will be load balanced so same name for all instances of upload servers
// upload-response-{instance_id}: requires instance id to handle the response on the same instance which makes the request to the grpc server

// For deploy exchange:
// deploy-request
// deploy-response-{instance_id}
type Application struct {
	CookieStore *sessions.CookieStore
	store.UserStore

	GithubClientID        string
	GithubClientSecret    string
	GithubAPICallbackPath string
	GithubClients         map[string]*github.Client

	AppEmail    string
	AppPassword string
	SMTPHost    string
	SMTPPort    string

	UploadResponseClient *internal.RabbitClient
	DeployResponseClient *internal.RabbitClient
	ProjectChannels      map[string]chan models.RabbitMQResponse
	PublishingConn       *amqp.Connection
	RabbitMQInstanceID   string

	RequestHandlerServerAddr string

	sync.RWMutex
}

func NewApplication() (*Application, error) {
	db := createDB()

	userStore := store.NewUserStore(db)

	clientID := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")
	callbackPath := os.Getenv("GITHUB_OAUTH_CALLBACK_PATH")
	appPassword := os.Getenv("APP_PASSWORD")
	appEmail := os.Getenv("APP_EMAIL")
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")

	rabbitMQUser := os.Getenv("RABBITMQ_USER")
	rabbitMQPassword := os.Getenv("RABBITMQ_PASSWORD")
	rabbitMQVhost := os.Getenv("RABBITMQ_VHOST")
	rabbitMQAddr := os.Getenv("RABBITMQ_ADDR")

	instanceID := uuid.NewString()

	// each concurrent task should be done with new channel
	// different connections should be used for publishing and consuming

	consumingConnection, err := internal.ConnectRabbitMQ(rabbitMQUser, rabbitMQPassword, rabbitMQAddr, rabbitMQVhost)
	if err != nil {
		return nil, err
	}

	publishingConnection, err := internal.ConnectRabbitMQ(rabbitMQUser, rabbitMQPassword, rabbitMQAddr, rabbitMQVhost)
	if err != nil {
		return nil, err
	}

	_, err = internal.CreateNewQueueReturnClient(publishingConnection, "upload-request", true, true)
	if err != nil {
		return nil, err
	}

	uploadResponseClient, err := internal.CreateNewQueueReturnClient(consumingConnection, "upload-response-"+instanceID, true, true)
	if err != nil {
		return nil, err
	}

	_, err = internal.CreateNewQueueReturnClient(publishingConnection, "deploy-request", true, true)
	if err != nil {
		return nil, err
	}

	deployResponseClient, err := internal.CreateNewQueueReturnClient(consumingConnection, "deploy-response-"+instanceID, true, true)
	if err != nil {
		return nil, err
	}

	return &Application{
		CookieStore:              sessions.NewCookieStore([]byte(os.Getenv("SECRET"))),
		UserStore:                userStore,
		GithubClientID:           clientID,
		GithubClientSecret:       clientSecret,
		GithubAPICallbackPath:    callbackPath,
		GithubClients:            make(map[string]*github.Client),
		AppEmail:                 appEmail,
		AppPassword:              appPassword,
		SMTPHost:                 smtpHost,
		SMTPPort:                 smtpPort,
		UploadResponseClient:     uploadResponseClient,
		DeployResponseClient:     deployResponseClient,
		RabbitMQInstanceID:       instanceID,
		PublishingConn:           publishingConnection,
		ProjectChannels:          make(map[string]chan models.RabbitMQResponse),
		RequestHandlerServerAddr: os.Getenv("REQUEST_HANDLER_ADDR"),
	}, nil
}

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

func getSession(c echo.Context, key string) (string, error) {
	session := c.Get("session").(*sessions.Session)
	v, ok := session.Values[key]
	if !ok {
		return "", models.ErrInvalidRequest
	}

	return v.(string), nil
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

func generateOTP(length int) (string, error) {
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

func generateID(length int) string {
	charset := "abcdefghijklmnopqrstuvwxyz0123456789"
	id := make([]byte, length)

	seedRand := math_rand.New(math_rand.NewSource(time.Now().UnixNano()))
	for i := range id {
		id[i] = charset[seedRand.Intn(len(charset))]
	}

	return string(id)
}

func (app *Application) setupEmailVerification(c echo.Context, username, email string) error {
	code, err := generateOTP(5)
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
