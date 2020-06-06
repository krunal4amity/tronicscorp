package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestUsers(t *testing.T) {

	t.Run("test create user invalid data unhappy", func(t *testing.T) {
		body := `
		{
			"username":"krunal.shimpi@gmail.com",
			"password":"abc12"
		}
		`
		req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(body))
		res := httptest.NewRecorder()
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		e := echo.New()
		c := e.NewContext(req, res)
		uh.Col = usersCol
		err := uh.CreateUser(c)
		t.Logf("res: %#+v\n", string(res.Body.Bytes()))
		assert.Nil(t, err)
		assert.Equal(t, http.StatusBadRequest, res.Code)
	})

	t.Run("test create user", func(t *testing.T) {
		var user User
		body := `
		{
			"username":"krunal.shimpi@gmail.com",
			"password":"abc12345"
		}
		`
		req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(body))
		res := httptest.NewRecorder()
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		e := echo.New()
		c := e.NewContext(req, res)
		uh.Col = usersCol
		err := uh.CreateUser(c)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusCreated, res.Code)
		token := res.Header().Get("X-Auth-Token")
		assert.NotEmpty(t, token)
		err = json.Unmarshal(res.Body.Bytes(), &user)
		assert.Nil(t, err)
		assert.Equal(t, "krunal.shimpi@gmail.com", user.Email)
		assert.Empty(t, user.Password)
	})

	t.Run("test create user again unhappy", func(t *testing.T) {
		body := `
		{
			"username":"krunal.shimpi@gmail.com",
			"password":"abc12345"
		}
		`
		req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(body))
		res := httptest.NewRecorder()
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		e := echo.New()
		c := e.NewContext(req, res)
		uh.Col = usersCol
		err := uh.CreateUser(c)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusBadRequest, res.Code)
	})
	t.Run("test authenticate user", func(t *testing.T) {
		var user User
		body := `
		{
			"username":"krunal.shimpi@gmail.com",
			"password":"abc12345"
		}
		`
		req := httptest.NewRequest(http.MethodPost, "/auth", strings.NewReader(body))
		res := httptest.NewRecorder()
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		e := echo.New()
		c := e.NewContext(req, res)
		uh.Col = usersCol
		err := uh.AuthnUser(c)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, res.Code)
		token := res.Header().Get("X-Auth-Token")
		assert.NotEmpty(t, token)
		err = json.Unmarshal(res.Body.Bytes(), &user)
		assert.Nil(t, err)
		assert.Equal(t, "krunal.shimpi@gmail.com", user.Email)
		assert.Empty(t, user.Password)
	})
}
