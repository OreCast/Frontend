package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// DocsHandler provides handler for /docs end-point
func DocsHandler(c *gin.Context) {
	tmpl := makeTmpl("Documentation")
	top := tmplPage("top.tmpl", tmpl)
	bottom := tmplPage("bottom.tmpl", tmpl)
	tmpl["Title"] = "OreCast documentation"
	fname := "static/markdown/ProofOfConcept.md"
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

// MetaDataHandler provides content for /meta endpoint
func MetaDataHandler(c *gin.Context) {
	tmpl := makeTmpl("MetaData")
	top := tmplPage("top.tmpl", tmpl)
	bottom := tmplPage("bottom.tmpl", tmpl)
	tmpl["Content"] = "OreCast MetaData page"
	content := tmplPage("content.tmpl", tmpl)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(top+content+bottom))
}

// DiscoveryHandler provides content for /discovery endpoint
func DiscoveryHandler(c *gin.Context) {
	tmpl := makeTmpl("Discovery")
	top := tmplPage("top.tmpl", tmpl)
	bottom := tmplPage("bottom.tmpl", tmpl)
	tmpl["Content"] = "OreCast discovery"
	content := tmplPage("content.tmpl", tmpl)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(top+content+bottom))
}

// AnalyticsHandler provides content for /analytics endpoint
func AnalyticsHandler(c *gin.Context) {
	tmpl := makeTmpl("Analytics")
	top := tmplPage("top.tmpl", tmpl)
	bottom := tmplPage("bottom.tmpl", tmpl)
	tmpl["Content"] = "OreCast analytics page"
	content := tmplPage("content.tmpl", tmpl)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(top+content+bottom))
}

// ProvenanceHandler provides content for /provenance endpoint
func ProvenanceHandler(c *gin.Context) {
	tmpl := makeTmpl("Provenance")
	top := tmplPage("top.tmpl", tmpl)
	bottom := tmplPage("bottom.tmpl", tmpl)
	tmpl["Content"] = "OreCast provenant page"
	content := tmplPage("content.tmpl", tmpl)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(top+content+bottom))
}

// SiteHandler provides access to /sites endpoint
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

// StorageHandler provide access to /storage endpoint
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
