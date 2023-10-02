package main

import (
	"log"
	"net/http"

	authz "github.com/OreCast/Authz/auth"
	"github.com/gin-gonic/gin"
)

// _token is used across all authorized APIs
var _token *authz.Token

// gin cookies
// https://gin-gonic.com/docs/examples/cookie/
// more advanced use-case:
// https://stackoverflow.com/questions/66289603/use-existing-session-cookie-in-gin-router
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// check if our user key is set
		if user, err := c.Cookie("user"); err == nil {
			if Config.Verbose > 0 {
				log.Println(c.Request.Method, c.Request.URL.Path, user)
			}
			c.Set("user", user)
			if err := refreshToken(); err != nil {
				content := errorTmpl(c, "unable to get valid token", err)
				c.Data(http.StatusUnauthorized, "text/html; charset=utf-8", []byte(content))
				return
			}
			return
		}

		if user, ok := c.Get("user"); !ok {
			if Config.Verbose > 0 {
				log.Println(c.Request.Method, c.Request.URL.Path)
			}
			c.Redirect(http.StatusFound, "/login")
		} else {
			if Config.Verbose > 0 {
				log.Println(c.Request.Method, c.Request.URL.Path, user)
			}
		}
		if err := refreshToken(); err != nil {
			content := errorTmpl(c, "unable to get valid token", err)
			c.Data(http.StatusUnauthorized, "text/html; charset=utf-8", []byte(content))
			return
		}
		c.Next()
	}
}

func refreshToken() error {
	// check and obtain token
	var err error
	if _token == nil {
		if token, err := getToken(); err == nil {
			_token = &token
		} else {
			return err
		}
	} else {
		err = _token.Validate(Config.AuthzClientId)
	}
	return err
}
