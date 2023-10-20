package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"time"

	oreConfig "github.com/OreCast/common/config"
	"github.com/gin-gonic/gin"
)

// content is our static web server content.
//
//go:embed static
var StaticFs embed.FS

// helper function to parse given template and return HTML page
func tmplPage(tmpl string, tmplData TmplRecord) string {
	if tmplData == nil {
		tmplData = make(TmplRecord)
	}
	var templates Templates
	page := templates.Tmpl(tmpl, tmplData)
	return page
}

// helper function to make initial template struct
func makeTmpl(c *gin.Context, title string) TmplRecord {
	tmpl := make(TmplRecord)
	tmpl["Title"] = title
	tmpl["User"] = ""
	if user, ok := c.Get("user"); ok {
		tmpl["User"] = user
	}
	tmpl["Base"] = oreConfig.Config.Frontend.WebServer.Base
	tmpl["ServerInfo"] = oreConfig.Info()
	tmpl["Top"] = tmplPage("top.tmpl", tmpl)
	tmpl["Bottom"] = tmplPage("bottom.tmpl", tmpl)
	tmpl["StartTime"] = time.Now().Unix()
	return tmpl
}

// helper function which sets gin router and defines all our server end-points
func setupRouter() *gin.Engine {
	// Disable Console Color
	// gin.DisableConsoleColor()
	r := gin.Default()

	// middlewares: https://gin-gonic.com/docs/examples/using-middleware/
	// Recovery middleware recovers from any panics and writes a 500 if there was one.
	r.Use(gin.Recovery())

	authorized := r.Group("/")

	// GET end-points
	r.GET("/docs", DocsHandler)
	r.GET("/docs/:page", DocsHandler)
	r.GET("/login", LoginHandler)
	r.GET("/logout", LogoutHandler)
	r.GET("/user/registration", UserRegistryHandler)

	// captcha access
	r.GET("/captcha/:file", CaptchaHandler())

	// POST end-poinst
	r.POST("/login", LoginPostHandler)
	r.POST("/user/registration", UserRegistryPostHandler)

	// all other methods ahould be authorized
	authorized.Use(AuthMiddleware())
	{
		// GET methods
		authorized.GET("/datasets", DatasetHandler)
		authorized.GET("/dataset/:dataset", DatasetHandler)

		authorized.GET("/meta", MetaDataHandler)
		authorized.GET("/meta/record/:mid/:site", MetaRecordHandler)
		authorized.GET("/meta/:site", MetaSiteHandler)
		authorized.GET("/meta/:site/upload", MetaUploadHandler)
		authorized.GET("/meta/:site/delete", MetaDeleteHandler)

		authorized.GET("/sites", SitesHandler)
		authorized.GET("/site/:site", SitesHandler)
		authorized.GET("/site/registration", SiteRegistrationHandler)

		authorized.GET("/data/registration", DataRegistrationHandler)
		authorized.GET("/data/:site/upload", DataUploadHandler)
		authorized.GET("/data/:site/delete", DataDeleteHandler)

		authorized.GET("/storage/:site", SiteBucketsHandler)
		authorized.GET("/storage/:site/:bucket", BucketObjectsHandler)
		authorized.GET("/storage/:site/create", S3CreateHandler)
		authorized.GET("/storage/:site/upload", S3UploadHandler)
		authorized.GET("/storage/:site/delete", S3DeleteHandler)

		authorized.GET("/analytics", AnalyticsHandler)
		authorized.GET("/discovery", DiscoveryHandler)
		authorized.GET("/provenance", ProvenanceHandler)
		authorized.GET("/project", ProjectHandler)
		authorized.GET("/project/:page", ProjectHandler)

		// POST methods
		authorized.POST("/project/registration", ProjectRegistrationPostHandler)

		authorized.POST("/site/registration", SiteRegistrationPostHandler)

		authorized.POST("/data/registration", DataRegistrationPostHandler)

		authorized.POST("/storage/create", S3CreatePostHandler)
		authorized.POST("/storage/upload", S3UploadPostHandler)
		authorized.POST("/storage/delete", S3DeletePostHandler)

		authorized.POST("/meta/upload", MetaUploadPostHandler)
		authorized.POST("/meta/delete", MetaDeletePostHandler)

		authorized.POST("/data/upload", DataUploadPostHandler)
		authorized.POST("/data/delete", DataDeletePostHandler)
	}

	// static files
	for _, dir := range []string{"js", "css", "images"} {
		filesFS, err := fs.Sub(StaticFs, "static/"+dir)
		if err != nil {
			panic(err)
		}
		m := fmt.Sprintf("%s/%s", oreConfig.Config.Frontend.WebServer.Base, dir)
		r.StaticFS(m, http.FS(filesFS))
	}

	r.GET("/", IndexHandler)
	return r
}

// Server defines our HTTP server
func Server() {
	r := setupRouter()
	sport := fmt.Sprintf(":%d", oreConfig.Config.Frontend.WebServer.Port)
	log.Printf("Start HTTP server %s", sport)
	r.Run(sport)
}
