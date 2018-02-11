package main

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"

	// session "github.com/ipfans/echo-session"
	session "github.com/JamsMendez/echo-session"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/facebook"
	"github.com/markbates/goth/providers/twitter"
	"github.com/stretchr/objx"

	echoauth "../oauth"
)

var (
	oauth2Client = echoauth.Client{
		DefaultProvider: "facebook",
		Store:           sessionStoreFnk(sessionStoreGet),
	}
)

func init() {
	setupLog()
	db := gormConnect()
	logger.Infof("db", db)
}

func main() {
	// Init oauth provider
	goth.UseProviders(
		facebook.New(
			os.Getenv("FACEBOOK_KEY"),
			os.Getenv("FACEBOOK_SECRET"),
			"http://localhost:3000/auth/facebook/callback",
		),
		twitter.NewAuthenticate(
			os.Getenv("TWITTER_KEY"),
			os.Getenv("TWITTER_SECRET"),
			"http://localhost:3000/auth/twitter/callback"),
	)

	serv := echo.New()
	serv.Use(middleware.Logger())
	serv.Use(middleware.Recover())
	store := session.NewCookieStore([]byte("secret"))
	serv.Use(session.Sessions("sessid", store))

	t := &Template{
		templates: template.Must(template.ParseGlob("template/*.html")),
	}
	serv.Renderer = t
	serv.GET("/", RenderIndex)
	serv.GET("/auth/:provider", oauth2Client.Begin)
	serv.GET("/logout", oauth2Client.End)
	serv.GET("/auth/:provider/callback",
		oauth2Client.Callback(func(user goth.User, err error, ctx echo.Context) error {
			if err != nil {
				logger.Warnf("", err)
			}
			logger.Debug(user)
			authCookieValue := objx.New(map[string]interface{}{
				"name":       user.Name,
				"avatar_url": user.AvatarURL,
			}).MustBase64()
			cookie := new(http.Cookie)
			cookie.Name = "auth"
			cookie.Value = authCookieValue
			cookie.Path = "/"
			ctx.SetCookie(cookie)
			return ctx.Redirect(http.StatusFound, "/")
		}))

	fmt.Println("Run server: http://localhost:3000")
	serv.Start(":3000")
}

// Session wrapper
type Session struct {
	session.Session
}

// Get returns the session value associated to the given key.
func (s Session) Get(key string) (interface{}, error) {
	return s.Session.Get(key), nil
}

// Set sets the session value associated to the given key.
func (s Session) Set(key string, value interface{}) error {
	s.Session.Set(key, value)
	return nil
}

// Delete removes the session value associated to the given key.
func (s Session) Delete(key string) error {
	s.Session.Delete(key)
	return nil
}

// Save saves all sessions used during the current request.
func (s Session) Save(ctx echo.Context) error {
	return s.Session.Save()
}

func sessionStoreGet(ctx echo.Context) (echoauth.Session, error) {
	return &Session{
		Session: session.Default(ctx),
	}, nil
}

type sessionStoreFnk func(ctx echo.Context) (echoauth.Session, error)

func (f sessionStoreFnk) Get(ctx echo.Context) (echoauth.Session, error) {
	return f(ctx)
}

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func RenderIndex(c echo.Context) error {
	data := make(map[string]interface{})
	cookie, err := c.Cookie("auth")
	if err == nil {
		data["UserData"] = objx.MustFromBase64(cookie.Value)
	}
	return c.Render(http.StatusOK, "index", data)
}
