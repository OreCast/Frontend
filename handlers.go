package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Documentation about gib handlers can be found over here:
// https://go.dev/doc/tutorial/web-service-gin

// DocsHandler provides access to GET /docs end-point
func DocsHandler(c *gin.Context) {
	tmpl := makeTmpl("Documentation")
	top := tmplPage("top.tmpl", tmpl)
	bottom := tmplPage("bottom.tmpl", tmpl)
	tmpl["Title"] = "OreCast documentation"
	fname := "static/markdown/main.md"
	content, err := mdToHTML(fname)
	if err != nil {
		content = fmt.Sprintf("unable to convert %s to HTML, error %v", fname, err)
		log.Println("ERROR: ", content)
		tmpl["Content"] = content
	}
	tmpl["Content"] = template.HTML(content)
	content = tmplPage("content.tmpl", tmpl)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(top+content+bottom))
}

// MetaDataHandler provides access to GET /meta endpoint
func MetaDataHandler(c *gin.Context) {
	tmpl := makeTmpl("MetaData")
	top := tmplPage("top.tmpl", tmpl)
	bottom := tmplPage("bottom.tmpl", tmpl)
	tmpl["Content"] = "OreCast MetaData page"
	content := tmplPage("content.tmpl", tmpl)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(top+content+bottom))
}

// DiscoveryHandler provides access to GET /discovery endpoint
func DiscoveryHandler(c *gin.Context) {
	tmpl := makeTmpl("Discovery")
	top := tmplPage("top.tmpl", tmpl)
	bottom := tmplPage("bottom.tmpl", tmpl)
	tmpl["Content"] = "OreCast discovery"
	content := tmplPage("content.tmpl", tmpl)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(top+content+bottom))
}

// AnalyticsHandler provides access to GET /analytics endpoint
func AnalyticsHandler(c *gin.Context) {
	tmpl := makeTmpl("Analytics")
	top := tmplPage("top.tmpl", tmpl)
	bottom := tmplPage("bottom.tmpl", tmpl)
	tmpl["Content"] = "OreCast analytics page"
	content := tmplPage("content.tmpl", tmpl)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(top+content+bottom))
}

// ProvenanceHandler provides access to GET /provenance endpoint
func ProvenanceHandler(c *gin.Context) {
	tmpl := makeTmpl("Provenance")
	top := tmplPage("top.tmpl", tmpl)
	bottom := tmplPage("bottom.tmpl", tmpl)
	tmpl["Content"] = "OreCast provenant page"
	content := tmplPage("content.tmpl", tmpl)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(top+content+bottom))
}

// SiteHandler provides access to GET /sites endpoint
func SitesHandler(c *gin.Context) {
	tmpl := makeTmpl("Sites")
	top := tmplPage("top.tmpl", tmpl)
	bottom := tmplPage("bottom.tmpl", tmpl)
	var content string
	for _, sobj := range sites() {
		site := sobj.Name
		content += fmt.Sprintf("Site: <a href=\"%s/storage?site=%s\">%s</a>", Config.Base, site, site)
		content += fmt.Sprintf("<br/>Storage: <a href=\"%s\">S3</a>", sobj.URL)
		if sobj.Description != "" {
			content += "<br/>Description: " + sobj.Description + "<hr/>"
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
}

// StorageHandler provide access to GET /storage endpoint
func StorageHandler(c *gin.Context) {
	tmpl := makeTmpl("Storage")
	top := tmplPage("top.tmpl", tmpl)
	bottom := tmplPage("bottom.tmpl", tmpl)
	var params SiteParams
	c.Bind(&params)
	siteObj := site(params.Site, params.Bucket)
	tmpl["Datasets"] = siteObj.Datasets
	tmpl["Site"] = params.Site
	content := tmplPage("datasets.tmpl", tmpl)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(top+content+bottom))
}

// DataHandler provide access to GET /data endpoint
func DataHandler(c *gin.Context) {
}

// DataAccessHandler provide access to GET /data/access endpoint
func DataAccessHandler(c *gin.Context) {
}

// SiteAccessHandler provide access to GET /site/access endpoint
func SiteAccessHandler(c *gin.Context) {
}

// POST handlers

// SiteRegistrationHandler provide access to GET /site/registration endpoint
func SiteRegistrationHandler(c *gin.Context) {
	tmpl := makeTmpl("Storage")
	top := tmplPage("top.tmpl", tmpl)
	bottom := tmplPage("bottom.tmpl", tmpl)
	content := tmplPage("site_registration.tmpl", tmpl)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(top+content+bottom))
}

// DataRegistrationHandler provide access to GET /data/registration endpoint
func DataRegistrationHandler(c *gin.Context) {
}
