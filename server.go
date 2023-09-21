package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// content is our static web server content.
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
func makeTmpl(title string) TmplRecord {
	tmpl := make(TmplRecord)
	tmpl["Title"] = title
	tmpl["User"] = ""
	tmpl["Base"] = Config.Base
	tmpl["ServerInfo"] = info()
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

	// GET end-points
	r.GET("/docs", DocsHandler)
	r.GET("/data", DataHandler)
	r.GET("/data/access", DataAccessHandler)
	r.GET("/meta", MetaDataHandler)
	r.GET("/sites", SitesHandler)
	r.GET("/site/access", SiteAccessHandler)
	r.GET("/analytics", AnalyticsHandler)
	r.GET("/discovery", DiscoveryHandler)
	r.GET("/provenance", ProvenanceHandler)
	r.GET("/meta/:site", MetaSiteHandler)
	r.GET("/storage/:site", SiteBucketsHandler)
	r.GET("/storage/:site/:bucket", BucketObjectsHandler)
	r.GET("/storage/:site/create", S3CreateHandler)
	r.GET("/storage/:site/upload", S3UploadHandler)
	r.GET("/storage/:site/delete", S3DeleteHandler)
	r.GET("/site/registration", SiteRegistrationHandler)
	r.GET("/data/registration", DataRegistrationHandler)
	r.GET("/login", LoginHandler)
	r.GET("/registry", UserRegistryHandler)

	// captcha access
	r.GET("/captcha/:file", CaptchaHandler())

	// POST end-poinst
	r.POST("/site/registration", SiteRegistrationPostHandler)
	r.POST("/data/registration", DataRegistrationPostHandler)
	r.POST("/login", LoginPostHandler)
	r.POST("/registry", UserRegistryPostHandler)
	r.POST("/storage/create", S3CreatePostHandler)
	r.POST("/storage/upload", S3UploadPostHandler)
	r.POST("/storage/delete", S3DeletePostHandler)

	// static files
	for _, dir := range []string{"js", "css", "images"} {
		filesFS, err := fs.Sub(StaticFs, "static/"+dir)
		if err != nil {
			panic(err)
		}
		m := fmt.Sprintf("%s/%s", Config.Base, dir)
		r.StaticFS(m, http.FS(filesFS))
	}

	r.GET("/", IndexHandler)
	return r
}

// Server defines our HTTP server
func Server(configFile string) {
	r := setupRouter()
	sport := fmt.Sprintf(":%d", Config.Port)
	log.Printf("Start HTTP server %s", sport)
	r.Run(sport)
}
