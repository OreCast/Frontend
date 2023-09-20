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

// CaptchaHandler provides access to captcha server
func CaptchaHandler() gin.HandlerFunc {
	hdlr := captcha.Server(captcha.StdWidth, captcha.StdHeight)
	return func(c *gin.Context) {
		hdlr.ServeHTTP(c.Writer, c.Request)
	}
}

// IndexHandler provides access to GET / end-point
func IndexHandler(c *gin.Context) {
	// top and bottom HTTP content from our templates
	tmpl := makeTmpl("OreCast home")
	top := tmplPage("top.tmpl", tmpl)
	bottom := tmplPage("bottom.tmpl", tmpl)
	content := tmplPage("index.tmpl", tmpl)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(top+content+bottom))
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
	tmpl["Base"] = Config.Base
	var content string
	for _, sobj := range sites() {
		site := sobj.Name
		if Config.Verbose > 0 {
			log.Printf("processing %+v", sobj)
		}
		records := metadata(site)
		tmpl["Site"] = site
		tmpl["Description"] = sobj.Description
		tmpl["UseSSL"] = sobj.UseSSL
		tmpl["Records"] = records
		tmpl["NRecords"] = len(records)
		siteContent := tmplPage("site_record.tmpl", tmpl)
		content += fmt.Sprintf("%s", template.HTML(siteContent))
	}
	tmpl["Content"] = template.HTML(content)
	sites := tmplPage("sites.tmpl", tmpl)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(top+sites+bottom))
}

// StorageData represents storage data structure returned as data from DataManagement service
type StorageData struct {
	Site    string           `json:"site"`
	Bucket  string           `json:"bucket"`
	Objects []map[string]any `json:"objects"`
}

// BucketData represents storage info structure returned by DataManagement service
type BucketData struct {
	Status string      `json:"status"`
	Data   StorageData `json:"data"`
	Error  string      `json:"error"`
}

// BucketObject represents bucket object returned by DataManagement service
type BucketObject struct {
	Name         string `json:"name"`
	CreationDate string `json:"creationDate"`
}

// BucketsData represents buckets data returned as data from DataManagement service
type BucketsData struct {
	Site    string         `json:"site"`
	Buckets []BucketObject `json:"buckets"`
}

// SiteBucketsData represents site buckets data returned by DataManagement service
type SiteBucketsData struct {
	Status string      `json:"status"`
	Data   BucketsData `json:"data"`
	Error  string      `json:"error"`
}

// StorageParams represents URI storage params in /storage/:site/:bucket end-point
type StorageParams struct {
	Site   string `uri:"site" binding:"required"`
	Bucket string `uri:"bucket"`
}

// Dataset represent dataset record on orecast web UI
type Dataset struct {
	Name         string
	Size         string
	ETag         string
	ShortETag    string
	LastModified string
}

// SiteBucketsHandler provides access to GET /storage/:site endpoint
func SiteBucketsHandler(c *gin.Context) {
	tmpl := makeTmpl("Storage")
	top := tmplPage("top.tmpl", tmpl)
	bottom := tmplPage("bottom.tmpl", tmpl)

	// read end-points uri parameters: /storage/:site
	var params StorageParams
	err := c.ShouldBindUri(&params)
	if err != nil {
		msg := fmt.Sprintf("fail to bind storage parameters, error %v", err)
		content := errorTmpl(msg, err)
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(top+content+bottom))
		return
	}
	site := params.Site

	// place request to DataManagement service to get either site or bucket info
	rurl := fmt.Sprintf("%s/storage/%s", Config.DataManagementURL, site)
	if Config.Verbose > 0 {
		log.Println("query DataManagement", rurl)
	}
	resp, err := http.Get(rurl)
	if err != nil {
		log.Println("ERROR:", err)
		msg := fmt.Sprintf("fail to obtain storage info, error %v", err)
		content := errorTmpl(msg, err)
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(top+content+bottom))
		return
	}
	defer resp.Body.Close()
	var bdata SiteBucketsData
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&bdata); err != nil {
		log.Println("ERROR:", err)
		msg := fmt.Sprintf("fail to obtain storage info, error %v", err)
		content := errorTmpl(msg, err)
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(top+content+bottom))
		return
	}
	if bdata.Status != "ok" {
		msg := fmt.Sprintf("fail to obtain storage info, error %v", bdata.Error)
		content := errorTmpl(msg, nil)
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(top+content+bottom))
		return
	}
	tmpl["StoragePath"] = fmt.Sprintf("/storage/%s holds %d buckets", site, len(bdata.Data.Buckets))
	tmpl["Buckets"] = bdata.Data.Buckets
	tmpl["Site"] = site
	content := tmplPage("buckets.tmpl", tmpl)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(top+content+bottom))
}

// BucketObjectsHandler provides access to GET /storage/:site/:bucket endpoint
func BucketObjectsHandler(c *gin.Context) {
	tmpl := makeTmpl("Storage")
	top := tmplPage("top.tmpl", tmpl)
	bottom := tmplPage("bottom.tmpl", tmpl)

	// read end-points uri parameters: /storage/:site/:bucket
	var params StorageParams
	err := c.ShouldBindUri(&params)
	if err != nil {
		msg := fmt.Sprintf("fail to bind storage parameters, error %v", err)
		content := errorTmpl(msg, err)
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(top+content+bottom))
		return
	}
	site := params.Site
	bucket := params.Bucket

	// place request to DataManagement service to get either site or bucket info
	rurl := fmt.Sprintf("%s/storage/%s", Config.DataManagementURL, site)
	if bucket != "" {
		rurl = fmt.Sprintf("%s/storage/%s/%s", Config.DataManagementURL, site, bucket)
	}
	if Config.Verbose > 0 {
		log.Println("query DataManagement", rurl)
	}
	resp, err := http.Get(rurl)
	if err != nil {
		log.Println("ERROR:", err)
		msg := fmt.Sprintf("fail to obtain storage info, error %v", err)
		content := errorTmpl(msg, err)
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(top+content+bottom))
		return
	}
	defer resp.Body.Close()
	var bdata BucketData
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&bdata); err != nil {
		log.Println("ERROR:", err)
		msg := fmt.Sprintf("fail to obtain storage info, error %v", err)
		content := errorTmpl(msg, err)
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(top+content+bottom))
		return
	}
	if bdata.Status != "ok" {
		msg := fmt.Sprintf("fail to obtain storage info, error %v", bdata.Error)
		content := errorTmpl(msg, nil)
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(top+content+bottom))
		return
	}
	// convert storage buckets data into appropriate HTML structure
	var datasets []Dataset
	for _, b := range bdata.Data.Objects {
		val, _ := b["name"]
		name := fmt.Sprintf("%v", val)
		val, _ = b["etag"]
		etag := fmt.Sprintf("%v", val)
		val, _ = b["size"]
		size := fmt.Sprintf("%v", val)
		val, _ = b["lastModified"]
		ltime := fmt.Sprintf("%v", val)
		d := Dataset{
			Name:         name,
			ETag:         etag,
			ShortETag:    etag[:10],
			LastModified: ltime,
			Size:         size}
		datasets = append(datasets, d)
	}
	tmpl["StoragePath"] = fmt.Sprintf("/storage/%s/%s holds %d objects", site, bucket, len(datasets))
	tmpl["Datasets"] = datasets
	tmpl["DataManagementURL"] = Config.DataManagementURL
	tmpl["Site"] = site
	tmpl["Bucket"] = bucket
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
