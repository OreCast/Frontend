package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

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
		c.Next()
	}
}
