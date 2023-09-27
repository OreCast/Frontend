package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/dchest/captcha"
	"github.com/gin-gonic/gin"
)

// Documentation about gib handlers can be found over here:
// https://go.dev/doc/tutorial/web-service-gin

//
// Data structure we use through the code
//

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

// UserRegistationForm represents site registration form on web UI
type UserRegistrationForm struct {
	Name            string `form:"user"`
	Password        string `form:"password"`
	CaptchaID       string `form:"captchaId"`
	CaptchaSolution string `form:"captchaSolution"`
}

// LoginForm represents login form
type LoginForm struct {
	User     string `form:"user" binding:"required"`
	Password string `form:"password" binding:"required"`
}

// CreateBucketForm represents create bucket registration form on web UI
type CreateBucketForm struct {
	Site   string `form:"site"`
	Bucket string `form:"bucket"`
}

// MetaSiteParams represents URI storage params in /meta/:site end-point
type MetaSiteParams struct {
	Site string `uri:"site" binding:"required"`
}

// DocsParams represents URI storage params in /docs/:page end-point
type DocsParams struct {
	Page string `uri:"page" binding:"required"`
}

//
// helper functions
//

// helper function to provides error template message
func errorTmpl(c *gin.Context, msg string, err error) string {
	tmpl := makeTmpl(c, "Status")
	tmpl["Content"] = template.HTML(fmt.Sprintf("<div>%s</div>\n<br/><h3>ERROR</h3>%v", msg, err))
	content := tmplPage("error.tmpl", tmpl)
	return content
}

// helper functiont to provides success template message
func successTmpl(c *gin.Context, msg string) string {
	tmpl := makeTmpl(c, "Status")
	tmpl["Content"] = template.HTML(fmt.Sprintf("<h3>SUCCESS</h3><div>%s</div>", msg))
	content := tmplPage("success.tmpl", tmpl)
	return content
}

//
// GET handlers
//

// CaptchaHandler provides access to captcha server
func CaptchaHandler() gin.HandlerFunc {
	hdlr := captcha.Server(captcha.StdWidth, captcha.StdHeight)
	return func(c *gin.Context) {
		hdlr.ServeHTTP(c.Writer, c.Request)
	}
}

// IndexHandler provides access to GET / end-point
func IndexHandler(c *gin.Context) {
	// check if user cookie is set, this is necessary as we do not
	// use authorization handler for / end-point
	if user, err := c.Cookie("user"); err == nil {
		c.Set("user", user)
	}
	// top and bottom HTTP content from our templates
	tmpl := makeTmpl(c, "OreCast home")
	top := tmplPage("top.tmpl", tmpl)
	bottom := tmplPage("bottom.tmpl", tmpl)
	content := tmplPage("index.tmpl", tmpl)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(top+content+bottom))
}

// DocsHandler provides access to GET /docs end-point
func DocsHandler(c *gin.Context) {
	// check if user cookie is set, this is necessary as we do not
	// use authorization handler for /docs end-point
	if user, err := c.Cookie("user"); err == nil {
		c.Set("user", user)
	}
	tmpl := makeTmpl(c, "Documentation")
	top := tmplPage("top.tmpl", tmpl)
	bottom := tmplPage("bottom.tmpl", tmpl)
	tmpl["Title"] = "OreCast documentation"
	fname := "static/markdown/main.md"
	var params DocsParams
	if err := c.ShouldBindUri(&params); err == nil {
		fname = fmt.Sprintf("static/markdown/%s", params.Page)
	}
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
	tmpl := makeTmpl(c, "MetaData")
	top := tmplPage("top.tmpl", tmpl)
	bottom := tmplPage("bottom.tmpl", tmpl)
	tmpl["Content"] = "OreCast MetaData page"
	content := tmplPage("content.tmpl", tmpl)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(top+content+bottom))
}

// DiscoveryHandler provides access to GET /discovery endpoint
func DiscoveryHandler(c *gin.Context) {
	tmpl := makeTmpl(c, "Discovery")
	top := tmplPage("top.tmpl", tmpl)
	bottom := tmplPage("bottom.tmpl", tmpl)
	tmpl["Content"] = "OreCast discovery"
	content := tmplPage("content.tmpl", tmpl)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(top+content+bottom))
}

// AnalyticsHandler provides access to GET /analytics endpoint
func AnalyticsHandler(c *gin.Context) {
	tmpl := makeTmpl(c, "Analytics")
	top := tmplPage("top.tmpl", tmpl)
	bottom := tmplPage("bottom.tmpl", tmpl)
	tmpl["Content"] = "OreCast analytics page"
	content := tmplPage("content.tmpl", tmpl)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(top+content+bottom))
}

// ProvenanceHandler provides access to GET /provenance endpoint
func ProvenanceHandler(c *gin.Context) {
	tmpl := makeTmpl(c, "Provenance")
	top := tmplPage("top.tmpl", tmpl)
	bottom := tmplPage("bottom.tmpl", tmpl)
	tmpl["Content"] = "OreCast provenant page"
	content := tmplPage("content.tmpl", tmpl)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(top+content+bottom))
}

// MetaSiteHandler provides access to GET /meta/:site endpoint
func MetaSiteHandler(c *gin.Context) {
	tmpl := makeTmpl(c, "Sites")
	top := tmplPage("top.tmpl", tmpl)
	bottom := tmplPage("bottom.tmpl", tmpl)
	tmpl["Base"] = Config.Base
	var params MetaSiteParams
	if err := c.ShouldBindUri(&params); err != nil {
		msg := fmt.Sprintf("fail to bind meta/:site parameters, error %v", err)
		content := errorTmpl(c, msg, err)
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(top+content+bottom))
		return
	}

	site := params.Site
	var records []MetaData
	for _, sobj := range sites() {
		if site == sobj.Name {
			if Config.Verbose > 0 {
				log.Printf("processing %+v", sobj)
			}
			tmpl["Description"] = sobj.Description
			tmpl["UseSSL"] = sobj.UseSSL
			rec := metadata(site)
			if rec.Status == "ok" {
				for _, r := range rec.Data {
					records = append(records, r)
				}
			} else {
				log.Printf("WARNING: failed metadata record %+v", rec)
			}
		}
	}
	tmpl["Site"] = site
	tmpl["Records"] = records
	tmpl["NRecords"] = len(records)
	meta := tmplPage("meta_records.tmpl", tmpl)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(top+meta+bottom))
}

// SiteHandler provides access to GET /sites endpoint
func SitesHandler(c *gin.Context) {
	tmpl := makeTmpl(c, "Sites")
	top := tmplPage("top.tmpl", tmpl)
	bottom := tmplPage("bottom.tmpl", tmpl)
	tmpl["Base"] = Config.Base
	var content string
	for _, sobj := range sites() {
		site := sobj.Name
		if Config.Verbose > 0 {
			log.Printf("processing %+v", sobj)
		}
		rec := metadata(site)
		tmpl["Site"] = site
		tmpl["Description"] = sobj.Description
		tmpl["UseSSL"] = sobj.UseSSL
		tmpl["NRecords"] = len(rec.Data)
		siteContent := tmplPage("site_record.tmpl", tmpl)
		content += fmt.Sprintf("%s", template.HTML(siteContent))
	}
	tmpl["Content"] = template.HTML(content)
	sites := tmplPage("sites.tmpl", tmpl)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(top+sites+bottom))
}

// SiteBucketsHandler provides access to GET /storage/:site endpoint
func SiteBucketsHandler(c *gin.Context) {
	tmpl := makeTmpl(c, "Storage")
	top := tmplPage("top.tmpl", tmpl)
	bottom := tmplPage("bottom.tmpl", tmpl)

	// read end-points uri parameters: /storage/:site
	var params StorageParams
	err := c.ShouldBindUri(&params)
	if err != nil {
		msg := fmt.Sprintf("fail to bind storage parameters, error %v", err)
		content := errorTmpl(c, msg, err)
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
		content := errorTmpl(c, msg, err)
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(top+content+bottom))
		return
	}
	defer resp.Body.Close()
	var bdata SiteBucketsData
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&bdata); err != nil {
		log.Println("ERROR:", err)
		msg := fmt.Sprintf("fail to obtain storage info, error %v", err)
		content := errorTmpl(c, msg, err)
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(top+content+bottom))
		return
	}
	if bdata.Status != "ok" {
		msg := fmt.Sprintf("fail to obtain storage info, error %v", bdata.Error)
		content := errorTmpl(c, msg, nil)
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(top+content+bottom))
		return
	}
	tmpl["StoragePath"] = fmt.Sprintf("/storage/%s", site)
	tmpl["Buckets"] = bdata.Data.Buckets
	tmpl["NBuckets"] = len(bdata.Data.Buckets)
	tmpl["Site"] = site
	content := tmplPage("buckets.tmpl", tmpl)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(top+content+bottom))
}

// BucketObjectsHandler provides access to GET /storage/:site/:bucket endpoint
func BucketObjectsHandler(c *gin.Context) {
	tmpl := makeTmpl(c, "Storage")
	top := tmplPage("top.tmpl", tmpl)
	bottom := tmplPage("bottom.tmpl", tmpl)

	// read end-points uri parameters: /storage/:site/:bucket
	var params StorageParams
	err := c.ShouldBindUri(&params)
	if err != nil {
		msg := fmt.Sprintf("fail to bind storage parameters, error %v", err)
		content := errorTmpl(c, msg, err)
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
		content := errorTmpl(c, msg, err)
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(top+content+bottom))
		return
	}
	defer resp.Body.Close()
	var bdata BucketData
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&bdata); err != nil {
		log.Println("ERROR:", err)
		msg := fmt.Sprintf("fail to obtain storage info, error %v", err)
		content := errorTmpl(c, msg, err)
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(top+content+bottom))
		return
	}
	if bdata.Status != "ok" {
		msg := fmt.Sprintf("fail to obtain storage info, error %v", bdata.Error)
		content := errorTmpl(c, msg, nil)
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
	tmpl["StoragePath"] = fmt.Sprintf("/storage/%s/%s", site, bucket)
	tmpl["Datasets"] = datasets
	tmpl["DataManagementURL"] = Config.DataManagementURL
	tmpl["NObjects"] = len(datasets)
	tmpl["Site"] = site
	tmpl["Bucket"] = bucket
	content := tmplPage("datasets.tmpl", tmpl)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(top+content+bottom))
}

// S3CreateHandler provides access to GET /storage/create endpoint
func S3CreateHandler(c *gin.Context) {
	tmpl := makeTmpl(c, "Create bucket")
	top := tmplPage("top.tmpl", tmpl)
	bottom := tmplPage("bottom.tmpl", tmpl)
	var params StorageParams
	var content string
	if err := c.ShouldBindUri(&params); err == nil {
		tmpl["Site"] = params.Site
		content = tmplPage("create_bucket.tmpl", tmpl)
	} else {
		content = errorTmpl(c, "binding error", err)
	}
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(top+content+bottom))
}

// S3UploadHandler provides access to GET /storage/upload endpoint
func S3UploadHandler(c *gin.Context) {
	tmpl := makeTmpl(c, "Upload data")
	top := tmplPage("top.tmpl", tmpl)
	bottom := tmplPage("bottom.tmpl", tmpl)
	var params StorageParams
	var content string
	if err := c.ShouldBindUri(&params); err == nil {
		tmpl["Site"] = params.Site
		content = tmplPage("upload_data.tmpl", tmpl)
	} else {
		content = errorTmpl(c, "binding error", err)
	}
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(top+content+bottom))
}

// S3DeleteHandler provides access to GET /storage/delete endpoint
func S3DeleteHandler(c *gin.Context) {
	tmpl := makeTmpl(c, "Delete bucket")
	top := tmplPage("top.tmpl", tmpl)
	bottom := tmplPage("bottom.tmpl", tmpl)
	var params StorageParams
	var content string
	if err := c.ShouldBindUri(&params); err == nil {
		tmpl["Site"] = params.Site
		content = tmplPage("delete_bucket.tmpl", tmpl)
	} else {
		content = errorTmpl(c, "binding error", err)
	}
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(top+content+bottom))
}

// DataHandler provides access to GET /data endpoint
func DataHandler(c *gin.Context) {
	c.String(400, "Not implemented yet")
}

// DataAccessHandler provides access to GET /data/access endpoint
func DataAccessHandler(c *gin.Context) {
	c.String(400, "Not implemented yet")
}

// SiteAccessHandler provides access to GET /site/access endpoint
func SiteAccessHandler(c *gin.Context) {
	c.String(400, "Not implemented yet")
}

// SiteRegistrationHandler provides access to GET /site/registration endpoint
func SiteRegistrationHandler(c *gin.Context) {
	tmpl := makeTmpl(c, "Storage")
	top := tmplPage("top.tmpl", tmpl)
	bottom := tmplPage("bottom.tmpl", tmpl)
	content := tmplPage("site_registration.tmpl", tmpl)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(top+content+bottom))
}

// LoginHandler provides access to GET /login endpoint
func LoginHandler(c *gin.Context) {
	tmpl := makeTmpl(c, "Login")
	top := tmplPage("top.tmpl", tmpl)
	bottom := tmplPage("bottom.tmpl", tmpl)
	content := tmplPage("login.tmpl", tmpl)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(top+content+bottom))
}

// UserRegistryHandler provides access to GET /registry endpoint
func UserRegistryHandler(c *gin.Context) {
	// check if user cookie is set, this is necessary as we do not
	// use authorization handler for /registry end-point
	if user, err := c.Cookie("user"); err == nil {
		c.Set("user", user)
	}
	tmpl := makeTmpl(c, "User registration")
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
	c.String(400, "Not implemented yet")
}

// MetaUploadHandler provides access to GET /meta/upload endpoint
func MetaUploadHandler(c *gin.Context) {
	c.String(400, "Not implemented yet")
}

// MetaDeleteHandler provides access to GET /meta/delete endpoint
func MetaDeleteHandler(c *gin.Context) {
	c.String(400, "Not implemented yet")
}

// DataUploadHandler provides access to GET /data/upload endpoint
func DataUploadHandler(c *gin.Context) {
	c.String(400, "Not implemented yet")
}

// DataDeleteHandler provides access to GET /data/delete endpoint
func DataDeleteHandler(c *gin.Context) {
	c.String(400, "Not implemented yet")
}

// POST handlers

// LoginPostHandler provides access to POST /login endpoint
func LoginPostHandler(c *gin.Context) {
	tmpl := makeTmpl(c, "OreCast login")
	top := tmplPage("top.tmpl", tmpl)
	bottom := tmplPage("bottom.tmpl", tmpl)
	var form LoginForm
	var content string
	var err error

	if err = c.ShouldBind(&form); err != nil {
		content = errorTmpl(c, "login form binding error", err)
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(top+content+bottom))
		return
	}
	// TODO: check user credentials in some backend
	c.Set("user", form.User)
	if Config.Verbose > 0 {
		log.Printf("login from user %s, context %+v, url path %s", form.User, c.Request.URL.Path)
	}

	// set our user cookie
	if _, err := c.Cookie("user"); err != nil {
		c.SetCookie("user", form.User, 3600, "/", "localhost", false, true)
	}

	// redirect to main page
	//     c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), "user", form.User))
	// redirect
	c.Redirect(http.StatusFound, "/")
	//     tmpl["Content"] = fmt.Sprintf("Successfully logged as %s", form.User)
	//     content = tmplPage("content.tmpl", tmpl)
	//     c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(top+content+bottom))
}

// MetaUploadPostHandler provides access to POST /meta/upload endpoint
func MetaUploadPostHandler(c *gin.Context) {
	c.String(400, "Not implemented yet")
}

// MetaDeletePostHandler provides access to POST /meta/delete endpoint
func MetaDeletePostHandler(c *gin.Context) {
	c.String(400, "Not implemented yet")
}

// DataUploadPostHandler provides access to POST /data/upload endpoint
func DataUploadPostHandler(c *gin.Context) {
	c.String(400, "Not implemented yet")
}

// DataDeletePostHandler provides access to POST /data/delete endpoint
func DataDeletePostHandler(c *gin.Context) {
	c.String(400, "Not implemented yet")
}

// S3CreatePostHandler provides access to POST /storage/create endpoint
func S3CreatePostHandler(c *gin.Context) {
	tmpl := makeTmpl(c, "Storage create bucket")
	top := tmplPage("top.tmpl", tmpl)
	bottom := tmplPage("bottom.tmpl", tmpl)
	var form CreateBucketForm
	var content string
	var err error

	if err = c.ShouldBind(&form); err != nil {
		content = errorTmpl(c, "site bucket create binding error", err)
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(top+content+bottom))
		return
	}
	site := form.Site
	bucket := form.Bucket
	// curl -X POST http://localhost:8340/storage/cornell/s3-bucket
	rurl := fmt.Sprintf("%s/storage/%s/%s", Config.DataManagementURL, site, bucket)
	if Config.Verbose > 0 {
		log.Println("query DataManagement", rurl)
	}
	resp, err := http.PostForm(rurl, url.Values{})
	if err != nil {
		log.Println("ERROR:", err)
		msg := fmt.Sprintf("fail to create bucket %s at site %s, error %v", bucket, site, err)
		content := errorTmpl(c, msg, err)
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(top+content+bottom))
		return
	}
	defer resp.Body.Close()

	msg := fmt.Sprintf("New bucket %s at site %s successfully created, response status %s", bucket, site, resp.Status)
	if resp.Status == "200 OK" {
		content = successTmpl(c, msg)
	} else {
		respBody, err := io.ReadAll(resp.Body)
		msg = fmt.Sprintf("failed response %+v", respBody)
		content = errorTmpl(c, msg, err)
	}
	tmpl["Content"] = template.HTML(content)
	content = tmplPage("content.tmpl", tmpl)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(top+content+bottom))
}

// S3UploadPostHandler provides access to POST /storage/upload endpoint
func S3UploadPostHandler(c *gin.Context) {
	c.String(400, "Not implemented yet")
}

// S3DeletePostHandler provides access to POST /storage/delete endpoint
func S3DeletePostHandler(c *gin.Context) {
	tmpl := makeTmpl(c, "Storage create bucket")
	top := tmplPage("top.tmpl", tmpl)
	bottom := tmplPage("bottom.tmpl", tmpl)
	var form CreateBucketForm
	var content string
	var err error

	if err = c.ShouldBind(&form); err != nil {
		content = errorTmpl(c, "site bucket delete binding error", err)
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(top+content+bottom))
		return
	}
	site := form.Site
	bucket := form.Bucket
	// curl -X DELETE http://localhost:8340/storage/cornell/s3-bucket
	rurl := fmt.Sprintf("%s/storage/%s/%s", Config.DataManagementURL, site, bucket)
	if Config.Verbose > 0 {
		log.Println("query DataManagement", rurl)
	}
	req, err := http.NewRequest("DELETE", rurl, nil)
	if err != nil {
		log.Println("ERROR:", err)
		msg := fmt.Sprintf("fail to delete bucket %s at site %s, error %v", bucket, site, err)
		content := errorTmpl(c, msg, err)
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(top+content+bottom))
		return
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("ERROR:", err)
		msg := fmt.Sprintf("fail to delete bucket %s at site %s, error %v", bucket, site, err)
		content := errorTmpl(c, msg, err)
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(top+content+bottom))
		return
	}
	defer resp.Body.Close()
	msg := fmt.Sprintf("Bucket %s at site %s successfully deleted, response status %s", bucket, site, resp.Status)
	if resp.Status == "200 OK" {
		content = successTmpl(c, msg)
	} else {
		respBody, err := io.ReadAll(resp.Body)
		msg = fmt.Sprintf("failed response %+v", respBody)
		content = errorTmpl(c, msg, err)
	}
	tmpl["Content"] = template.HTML(content)
	content = tmplPage("content.tmpl", tmpl)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(top+content+bottom))
}

// UserRegistryHandler provides access to POST /registry endpoint
func UserRegistryPostHandler(c *gin.Context) {
	tmpl := makeTmpl(c, "Storage")
	top := tmplPage("top.tmpl", tmpl)
	bottom := tmplPage("bottom.tmpl", tmpl)

	// parse input form request
	var form UserRegistrationForm
	var err error
	content := successTmpl(c, "User registation is completed")

	// first check if user provides the captcha
	if !captcha.VerifyString(form.CaptchaID, form.CaptchaSolution) {
		msg := "Wrong captcha match, robots are not allowed"
		content = errorTmpl(c, msg, err)
	}

	if err = c.ShouldBind(&form); err != nil {
		content = errorTmpl(c, "User registration binding error", err)
	}

	// return page
	tmpl["Content"] = template.HTML(content)
	content = tmplPage("user_registration.tmpl", tmpl)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(top+content+bottom))
}

// SiteRegistrationHandler provides access to POST /site/registration endpoint
func SiteRegistrationPostHandler(c *gin.Context) {
	tmpl := makeTmpl(c, "Storage")
	top := tmplPage("top.tmpl", tmpl)
	bottom := tmplPage("bottom.tmpl", tmpl)

	// parse input form request
	var form Site
	var err error
	content := successTmpl(c, "Site registration is successful")
	if err = c.ShouldBind(&form); err != nil {
		content = errorTmpl(c, "Site registration binding error", err)
	} else {
		if Config.Verbose > 0 {
			log.Printf("register site %+v", form)
		}

		// encrypt sensitive fields
		form, err = encryptSiteObject(form)
		if err != nil {
			content = errorTmpl(c, "Site registration failure to encrypt Site attributes", err)
		} else {
			// make JSON request to Discovery service
			if data, err := json.Marshal(form); err == nil {
				rurl := fmt.Sprintf("%s/sites", Config.DiscoveryURL)
				resp, err := http.Post(rurl, "application/json", bytes.NewBuffer(data))
				if err != nil {
					content = errorTmpl(c, "Site registration posting to discvoeru service failure", err)
					tmpl["Content"] = template.HTML(content)
				} else {
					if Config.Verbose > 0 {
						log.Printf("discovery service response: %s", resp.Status)
					}
				}
			} else {
				content = errorTmpl(c, "Site registration json marshalling error", err)
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
	c.String(400, "Not implemented yet")
}
