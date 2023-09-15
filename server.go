package main

import (
	"embed"
	"fmt"
	"html/template"
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

	// top and bottom HTTP content from our templates
	tmpl := makeTmpl("OreCast")
	top := tmplPage("top.tmpl", tmpl)
	bottom := tmplPage("bottom.tmpl", tmpl)

	tmpl = makeTmpl("Sites")
	r.GET("/sites", func(c *gin.Context) {
		var content string
		for _, sobj := range sites() {
			site := sobj.Name
			content += fmt.Sprintf("Site: <a href=\"%s/storage?site=%s\">%s</a>", Config.Base, site, site)
			if sobj.Description != "" {
				content += "<br/>" + sobj.Description + "<hr/>"
			}
			metaRecords := metadata(site)
			if Config.Verbose > 0 {
				log.Printf("for site %s meta-data records %+v", site, metaRecords)
			}
			content += "<h3>MetaData records</h3>"
			for _, rec := range metaRecords {
				content += fmt.Sprintf("ID: %s", rec.ID)
				content += fmt.Sprintf("<br/>Bucket: <a href=\"%s/storage?site=%s&bucket=%s\">%s</a>", Config.Base, site, rec.Bucket, rec.Bucket)
				if rec.Site == site {
					content += fmt.Sprintf("<br/>Description: %s", rec.Description)
				}
				if len(rec.Tags) > 0 {
					content += fmt.Sprintf("<br/>Tags: %v", rec.Tags)
				}
				content += "<hr/>"
			}
		}
		tmpl["Content"] = template.HTML(content)
		sites := tmplPage("content.tmpl", tmpl)
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(top+sites+bottom))
	})

	tmpl = makeTmpl("Datasets")
	//     tmpl["Datasets"] = datasets("")
	tmpl["Datasets"] = []string{}
	r.GET("/storage", func(c *gin.Context) {
		var params SiteParams
		c.Bind(&params)
		siteObj := site(params.Site, params.Bucket)
		tmpl["Datasets"] = siteObj.Datasets
		tmpl["Site"] = params.Site
		datasets := tmplPage("datasets.tmpl", tmpl)
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(top+datasets+bottom))
	})

	// static files
	for _, dir := range []string{"js", "css", "images"} {
		filesFS, err := fs.Sub(StaticFs, "static/"+dir)
		if err != nil {
			panic(err)
		}
		m := fmt.Sprintf("%s/%s", Config.Base, dir)
		r.StaticFS(m, http.FS(filesFS))
	}

	index := tmplPage("index.tmpl", tmpl)
	r.GET("/", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(top+index+bottom))
	})
	return r
}

func Server(configFile string) {
	r := setupRouter()
	sport := fmt.Sprintf(":%d", Config.Port)
	log.Printf("Start HTTP server %s", sport)
	r.Run(sport)
}
