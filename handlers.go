package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/dchest/captcha"
	"github.com/gin-gonic/gin"
)

// Documentation about gib handlers can be found over here:
// https://go.dev/doc/tutorial/web-service-gin

// helper function to provides error template message
func errorTmpl(msg string, err error) string {
	tmpl := makeTmpl("Status")
	tmpl["Content"] = template.HTML(fmt.Sprintf("<div>%s</div>\n<br/><h3>ERROR</h3>%v", msg, err))
	content := tmplPage("error.tmpl", tmpl)
	return content
}

// helper functiont to provides success template message
func successTmpl(msg string) string {
	tmpl := makeTmpl("Status")
	tmpl["Content"] = template.HTML(fmt.Sprintf("<h3>SUCCESS</h3><div>%s</div>", msg))
	content := tmplPage("success.tmpl", tmpl)
	return content
}

// helper captcha handler
func captchaHandler() gin.HandlerFunc {
	hdlr := captcha.Server(captcha.StdWidth, captcha.StdHeight)

	return func(c *gin.Context) {
		hdlr.ServeHTTP(c.Writer, c.Request)
	}
}

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

// StorageHandler provides access to GET /storage endpoint
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

// DataHandler provides access to GET /data endpoint
func DataHandler(c *gin.Context) {
}

// DataAccessHandler provides access to GET /data/access endpoint
func DataAccessHandler(c *gin.Context) {
}

// SiteAccessHandler provides access to GET /site/access endpoint
func SiteAccessHandler(c *gin.Context) {
}

// SiteRegistrationHandler provides access to GET /site/registration endpoint
func SiteRegistrationHandler(c *gin.Context) {
	tmpl := makeTmpl("Storage")
	top := tmplPage("top.tmpl", tmpl)
	bottom := tmplPage("bottom.tmpl", tmpl)
	content := tmplPage("site_registration.tmpl", tmpl)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(top+content+bottom))
}

// LoginHandler provides access to GET /login endpoint
func LoginHandler(c *gin.Context) {
	tmpl := makeTmpl("Login")
	top := tmplPage("top.tmpl", tmpl)
	bottom := tmplPage("bottom.tmpl", tmpl)
	content := tmplPage("login.tmpl", tmpl)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(top+content+bottom))
}

// UserRegistryHandler provides access to GET /registration endpoint
func UserRegistryHandler(c *gin.Context) {
	tmpl := makeTmpl("User registration")
	top := tmplPage("top.tmpl", tmpl)
	bottom := tmplPage("bottom.tmpl", tmpl)
	captchaStr := captcha.New()
	if Config.Verbose > 0 {
		log.Println("new captcha", captchaStr)
	}
	tmpl["CaptchaId"] = captchaStr
	tmpl["CaptchaPublicKey"] = Config.CaptchaPublicKey
	content := tmplPage("user_registration.tmpl", tmpl)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(top+content+bottom))
}

// DataRegistrationHandler provides access to GET /data/registration endpoint
func DataRegistrationHandler(c *gin.Context) {
}

// POST handlers

// UserRegistationForm represents site registration form on web UI
type UserRegistrationForm struct {
	Name            string `form:"user"`
	Password        string `form:"password"`
	CaptchaID       string `form:"captchaId"`
	CaptchaSolution string `form:"captchaSolution"`
}

// LoginPostHandler provides access to POST /login endpoint
func LoginPostHandler(c *gin.Context) {
}

// UserRegistryHandler provides access to POST /registry endpoint
func UserRegistryPostHandler(c *gin.Context) {
	tmpl := makeTmpl("Storage")
	top := tmplPage("top.tmpl", tmpl)
	bottom := tmplPage("bottom.tmpl", tmpl)

	// parse input form request
	var form UserRegistrationForm
	var err error
	content := successTmpl("User registation is completed")

	// first check if user provides the captcha
	if !captcha.VerifyString(form.CaptchaID, form.CaptchaSolution) {
		msg := "Wrong captcha match, robots are not allowed"
		content = errorTmpl(msg, err)
	}

	if err = c.ShouldBind(&form); err != nil {
		content = errorTmpl("User registration binding error", err)
	}

	// return page
	tmpl["Content"] = template.HTML(content)
	content = tmplPage("user_registration.tmpl", tmpl)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(top+content+bottom))
}

// SiteRegistrationHandler provides access to POST /site/registration endpoint
func SiteRegistrationPostHandler(c *gin.Context) {
	tmpl := makeTmpl("Storage")
	top := tmplPage("top.tmpl", tmpl)
	bottom := tmplPage("bottom.tmpl", tmpl)

	// parse input form request
	var form Site
	var err error
	content := successTmpl("Site registration is successful")
	if err = c.ShouldBind(&form); err != nil {
		content = errorTmpl("Site registration binding error", err)
	} else {
		if Config.Verbose > 0 {
			log.Printf("register site %+v", form)
		}

		// encrypt sensitive fields
		form, err = encryptSiteObject(form)
		if err != nil {
			content = errorTmpl("Site registration failure to encrypt Site attributes", err)
		} else {
			// make JSON request to Discovery service
			if data, err := json.Marshal(form); err == nil {
				rurl := fmt.Sprintf("%s/sites", Config.DiscoveryURL)
				resp, err := http.Post(rurl, "application/json", bytes.NewBuffer(data))
				if err != nil {
					content = errorTmpl("Site registration posting to discvoeru service failure", err)
					tmpl["Content"] = template.HTML(content)
				} else {
					if Config.Verbose > 0 {
						log.Printf("discovery service response: %s", resp.Status)
					}
				}
			} else {
				content = errorTmpl("Site registration json marshalling error", err)
			}
		}
	}

	// return page
	tmpl["Content"] = template.HTML(content)
	content = tmplPage("content.tmpl", tmpl)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(top+content+bottom))
}

// DataRegistrationPostHandler provides access to POST /data/registration endpoint
func DataRegistrationPostHandler(c *gin.Context) {
}
