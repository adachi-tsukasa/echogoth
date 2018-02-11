package echoauth

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/labstack/echo"
	"github.com/markbates/goth"
)

// SessionName is the key used to access the session store.
const SessionName = "_gothic_session"

// Errors list
var (
	ErrUndefinedProvider = errors.New("Undefined provider name")
)

// Session interface
type Session interface {
	Get(key string) (interface{}, error)
	Set(key string, value interface{}) error
	Delete(key string) error
	Save(ctx echo.Context) error
}

// SessionStore session accessor
type SessionStore interface {
	Get(ctx echo.Context) (Session, error)
}

// Client OAuth authorize
type Client struct {
	DefaultProvider string
	Store           SessionStore
}

func (cli *Client) Begin(ctx echo.Context) error {
	url, err := cli.GetAuthURL(ctx)
	if err != nil {
		return ctx.String(http.StatusBadRequest, err.Error())
	}
	return ctx.Redirect(http.StatusFound, url)
}

func (cli *Client) End(ctx echo.Context) error {
	cookie := new(http.Cookie)
	fmt.Println(cookie)
	cookie.Name = "auth"
	cookie.Value = ""
	cookie.Path = "/"
	cookie.MaxAge = -1
	ctx.SetCookie(cookie)
	return ctx.Redirect(http.StatusFound, "/")
}

func (cli *Client) Callback(fnk func(user goth.User, err error, ctx echo.Context) error) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		user, err := cli.GetUser(ctx)
		return fnk(user, err, ctx)
	}
}

func (cli *Client) GetUser(ctx echo.Context) (goth.User, error) {
	providerName := cli.getProviderName(ctx)
	if len(providerName) < 1 {
		return goth.User{}, ErrUndefinedProvider
	}

	provider, err := goth.GetProvider(providerName)
	if err != nil {
		return goth.User{}, err
	}

	session, err := cli.Store.Get(ctx)
	if nil != err {
		return goth.User{}, err
	}

	sessionData, err := session.Get(SessionName)
	if nil != err || nil == sessionData {
		return goth.User{}, err
	}

	session.Delete(SessionName)
	session.Save(ctx)

	sess, err := provider.UnmarshalSession(sessionData.(string))
	if err != nil {
		return goth.User{}, err
	}

	_, err = sess.Authorize(provider, url.Values(ctx.Request().URL.Query()))
	if err != nil {
		return goth.User{}, err
	}

	return provider.FetchUser(sess)
}

func (cli *Client) GetAuthURL(ctx echo.Context) (string, error) {
	providerName := cli.getProviderName(ctx)
	if len(providerName) < 1 {
		return "", ErrUndefinedProvider
	}

	provider, err := goth.GetProvider(providerName)
	if err != nil {
		return "", err
	}

	sess, err := provider.BeginAuth(setState(ctx))
	if err != nil {
		return "", err
	}

	url, err := sess.GetAuthURL()
	if err != nil {
		return "", err
	}

	var session Session
	if session, err = cli.Store.Get(ctx); nil == err {
		if err = session.Set(SessionName, sess.Marshal()); nil == err {
			err = session.Save(ctx)
		}
	}
	return url, err
}

func (cli *Client) getProviderName(ctx echo.Context) string {
	provider := ctx.Param("provider")
	if provider == "" {
		return cli.DefaultProvider
	}
	return provider
}

func setState(ctx echo.Context) string {
	if state := ctx.QueryParam("state"); len(state) > 0 {
		return state
	}
	return "state"
}

func getState(ctx echo.Context) string {
	return ctx.QueryParam("state")
}
