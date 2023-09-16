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

// SiteParam defines form parameters for site call
type SiteParams struct {
	Site   string `form:"site"`
	Bucket string `form:"bucket"`
}

func setupRouter() *gin.Engine {
	// Disable Console Color
	// gin.DisableConsoleColor()
	r := gin.Default()

	// various routes
	r.GET("/docs", DocsHandler)
	r.GET("/meta", MetaDataHandler)
	r.GET("/analytics", AnalyticsHandler)
	r.GET("/discovery", DiscoveryHandler)
	r.GET("/provenance", ProvenanceHandler)
	r.GET("/sites", SitesHandler)
	r.GET("/storage", StorageHandler)

	// static files
	for _, dir := range []string{"js", "css", "images"} {
		filesFS, err := fs.Sub(StaticFs, "static/"+dir)
		if err != nil {
			panic(err)
		}
		m := fmt.Sprintf("%s/%s", Config.Base, dir)
		r.StaticFS(m, http.FS(filesFS))
	}

	// top and bottom HTTP content from our templates
	tmpl := makeTmpl("OreCast home")
	top := tmplPage("top.tmpl", tmpl)
	bottom := tmplPage("bottom.tmpl", tmpl)
	content := tmplPage("index.tmpl", tmpl)
	r.GET("/", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(top+content+bottom))
	})
	return r
}

func Server(configFile string) {
	r := setupRouter()
	sport := fmt.Sprintf(":%d", Config.Port)
	log.Printf("Start HTTP server %s", sport)
	r.Run(sport)
}
