package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	authz "github.com/OreCast/common/authz"
	oreConfig "github.com/OreCast/common/config"
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
	Login           string `form:"login" json:"login"`
	Password        string `form:"password" json:"password"`
	FirstName       string `form:"first_name" json:"first_name"`
	LastName        string `form:"last_name" json:"last_name"`
	Email           string `form:"email" json:"email"`
	CaptchaID       string `form:"captchaId" json:",omitempty"`
	CaptchaSolution string `form:"captchaSolution" json:",omitempty"`
}

// LoginForm represents login form
type LoginForm struct {
	User     string `form:"user" binding:"required"`
	Password string `form:"password" binding:"required"`
}

// User represents structure used by users DB in Authz service to handle incoming requests
type User struct {
	Login    string
	Password string
}

// ProjectRegistationForm represents project registration form on web UI
type ProjectRegistrationForm struct {
	Project     string `form:"project"`
	Site        string `form:"site"`
	Description string `form:"description"`
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

// MetaIdParams represents URI storage params in /docs/:page end-point
type MetaIdParams struct {
	MetaId string `uri:"mid" binding:"required"`
}

// DsParams represents URI storage params in /docs/:page end-point
type DsParams struct {
	Dataset string `uri:"dataset" binding:"required"`
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
	user, err := c.Cookie("user")
	if err == nil {
		c.Set("user", user)
	}
	// top and bottom HTTP content from our templates
	tmpl := makeTmpl(c, "OreCast home")
	top := tmplPage("top.tmpl", tmpl)
	bottom := tmplPage("bottom.tmpl", tmpl)
	tmpl["LogoClass"] = "show"
	tmpl["MapClass"] = "hide"
	if user != "" {
		tmpl["LogoClass"] = "hide"
		tmpl["MapClass"] = "show"
	}
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

// MetaRecordHandler provides access to GET /meta/record/:mid endpoint
func MetaRecordHandler(c *gin.Context) {
	tmpl := makeTmpl(c, "Sites")
	top := tmplPage("top.tmpl", tmpl)
	bottom := tmplPage("bottom.tmpl", tmpl)
	tmpl["Base"] = oreConfig.Config.Frontend.WebServer.Base
	var params MetaIdParams
	if err := c.ShouldBindUri(&params); err != nil {
		msg := fmt.Sprintf("fail to bind meta/record/:mid parameters, error %v", err)
		content := errorTmpl(c, msg, err)
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(top+content+bottom))
		return
	}

	results := getMetaRecord(params.MetaId)
	if results.Status != "ok" {
		msg := fmt.Sprintf("fail to find mid %s", params.MetaId)
		content := errorTmpl(c, msg, errors.New("Not Found"))
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(top+content+bottom))
		return
	}
	data := results.Data
	record := data[0]
	tmpl["ID"] = record.ID
	tmpl["Description"] = record.Description
	tmpl["Tags"] = record.Tags
	tmpl["Bucket"] = record.Bucket
	meta := tmplPage("meta_record.tmpl", tmpl)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(top+meta+bottom))
}

// DatasetHandler provides access to GET /dataset endpoint
func DatasetHandler(c *gin.Context) {
	tmpl := makeTmpl(c, "Data")
	top := tmplPage("top.tmpl", tmpl)
	bottom := tmplPage("bottom.tmpl", tmpl)
	tmpl["Base"] = oreConfig.Config.Frontend.WebServer.Base
	var content, dsName string
	var params DsParams
	if err := c.ShouldBindUri(&params); err == nil {
		dsName = params.Dataset
	}
	for _, dobj := range getDatasets(dsName) {
		if oreConfig.Config.Frontend.WebServer.Verbose > 0 {
			log.Printf("processing %+v", dobj)
		}
		tmpl["Dataset"] = dobj.Dataset
		tmpl["Site"] = dobj.Site
		tmpl["MetaId"] = dobj.MetaId
		tmpl["Processing"] = dobj.Processing
		datasetContent := tmplPage("dataset_record.tmpl", tmpl)
		content += fmt.Sprintf("%s", template.HTML(datasetContent))
	}
	tmpl["Content"] = template.HTML(content)
	datasets := tmplPage("datasets.tmpl", tmpl)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(top+datasets+bottom))
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
	tmpl["Base"] = oreConfig.Config.Frontend.WebServer.Base
	var params MetaSiteParams
	if err := c.ShouldBindUri(&params); err != nil {
		msg := fmt.Sprintf("fail to bind meta/:site parameters, error %v", err)
		content := errorTmpl(c, msg, err)
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(top+content+bottom))
		return
	}

	site := params.Site
	var records []MetaData
	for _, sobj := range getSites() {
		if site == sobj.Name {
			if oreConfig.Config.Frontend.WebServer.Verbose > 0 {
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
	tmpl["Base"] = oreConfig.Config.Frontend.WebServer.Base
	var content string
	for _, sobj := range getSites() {
		site := sobj.Name
		if oreConfig.Config.Frontend.WebServer.Verbose > 0 {
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
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(top+content+bottom))
		return
	}
	site := params.Site

	// place request to DataManagement service to get either site or bucket info
	rurl := fmt.Sprintf("%s/storage/%s", oreConfig.Config.Services.DataManagementURL, site)
	if oreConfig.Config.Frontend.WebServer.Verbose > 0 {
		log.Println("query DataManagement", rurl)
	}
	resp, err := httpGet(rurl)
	if err != nil {
		log.Println("ERROR:", err)
		msg := fmt.Sprintf("fail to obtain storage info, error %v", err)
		content := errorTmpl(c, msg, err)
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(top+content+bottom))
		return
	}
	defer resp.Body.Close()
	var bdata SiteBucketsData
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&bdata); err != nil {
		log.Println("ERROR:", err)
		msg := fmt.Sprintf("fail to obtain storage info, error %v", err)
		content := errorTmpl(c, msg, err)
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(top+content+bottom))
		return
	}
	if bdata.Status != "ok" {
		msg := fmt.Sprintf("fail to obtain storage info, error %v", bdata.Error)
		content := errorTmpl(c, msg, nil)
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(top+content+bottom))
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
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(top+content+bottom))
		return
	}
	site := params.Site
	bucket := params.Bucket

	// place request to DataManagement service to get either site or bucket info
	rurl := fmt.Sprintf("%s/storage/%s", oreConfig.Config.Services.DataManagementURL, site)
	if bucket != "" {
		rurl = fmt.Sprintf("%s/storage/%s/%s", oreConfig.Config.Services.DataManagementURL, site, bucket)
	}
	if oreConfig.Config.Frontend.WebServer.Verbose > 0 {
		log.Println("query DataManagement", rurl)
	}
	resp, err := httpGet(rurl)
	if err != nil {
		log.Println("ERROR:", err)
		msg := fmt.Sprintf("fail to obtain storage info, error %v", err)
		content := errorTmpl(c, msg, err)
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(top+content+bottom))
		return
	}
	defer resp.Body.Close()
	var bdata BucketData
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&bdata); err != nil {
		log.Println("ERROR:", err)
		msg := fmt.Sprintf("fail to obtain storage info, error %v", err)
		content := errorTmpl(c, msg, err)
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(top+content+bottom))
		return
	}
	if bdata.Status != "ok" {
		msg := fmt.Sprintf("fail to obtain storage info, error %v", bdata.Error)
		content := errorTmpl(c, msg, nil)
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(top+content+bottom))
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
	tmpl["DataManagementURL"] = oreConfig.Config.Services.DataManagementURL
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
	status := http.StatusOK
	if err := c.ShouldBindUri(&params); err == nil {
		tmpl["Site"] = params.Site
		content = tmplPage("upload_data.tmpl", tmpl)
	} else {
		content = errorTmpl(c, "binding error", err)
		status = http.StatusBadRequest
	}
	c.Data(status, "text/html; charset=utf-8", []byte(top+content+bottom))
}

// S3DeleteHandler provides access to GET /storage/delete endpoint
func S3DeleteHandler(c *gin.Context) {
	tmpl := makeTmpl(c, "Delete bucket")
	top := tmplPage("top.tmpl", tmpl)
	bottom := tmplPage("bottom.tmpl", tmpl)
	var params StorageParams
	var content string
	status := http.StatusOK
	if err := c.ShouldBindUri(&params); err == nil {
		tmpl["Site"] = params.Site
		content = tmplPage("delete_bucket.tmpl", tmpl)
	} else {
		content = errorTmpl(c, "binding error", err)
		status = http.StatusBadRequest
	}
	c.Data(status, "text/html; charset=utf-8", []byte(top+content+bottom))
}

// ProjectHandler provides access to GET /project or /project/:page endpoints
func ProjectHandler(c *gin.Context) {
	tmpl := makeTmpl(c, "OreCast projects")
	top := tmplPage("top.tmpl", tmpl)
	bottom := tmplPage("bottom.tmpl", tmpl)

	page := "page"
	var params DocsParams
	if err := c.ShouldBindUri(&params); err == nil {
		page = params.Page
	}
	tname := fmt.Sprintf("project_%s.tmpl", page)
	upage := strings.ToUpper(page[:1]) + page[1:]
	tmpl["Title"] = fmt.Sprintf("%s summary page", upage)
	if page == "page" {
		tmpl["Title"] = "" // no need in title
	} else if page == "registration" {
		tmpl["Title"] = fmt.Sprintf("%s page", upage)
	}
	tmpl["Content"] = template.HTML(tmplPage(tname, tmpl))
	content := tmplPage("projects.tmpl", tmpl)
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
	tmpl := makeTmpl(c, "Site registration")
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

// LogoutHandler provides access to GET /logout endpoint
func LogoutHandler(c *gin.Context) {
	c.SetCookie("user", "", -1, "/", "localhost", false, true)
	c.Redirect(http.StatusFound, "/")
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
	if oreConfig.Config.Frontend.WebServer.Verbose > 0 {
		log.Println("new captcha", captchaStr)
	}
	tmpl["CaptchaId"] = captchaStr
	tmpl["CaptchaPublicKey"] = oreConfig.Config.Frontend.CaptchaPublicKey
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
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(top+content+bottom))
		return
	}

	// encrypt provided user password before sending to Authz server
	form, err = encryptLoginObject(form)
	if err != nil {
		content = errorTmpl(c, "unable to encrypt user password", err)
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(top+content+bottom))
		return
	}

	// make a call to Authz service to check for a user
	rurl := fmt.Sprintf("%s/oauth/authorize?client_id=%s&response_type=code", oreConfig.Config.Services.AuthzURL, oreConfig.Config.Authz.ClientId)
	user := User{Login: form.User, Password: form.Password}
	data, err := json.Marshal(user)
	if err != nil {
		content = errorTmpl(c, "unable to marshal user form, error", err)
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(top+content+bottom))
		return
	}
	resp, err := http.Post(rurl, "application/json", bytes.NewBuffer(data))
	if err != nil {
		content = errorTmpl(c, "unable to POST request to Authz service, error", err)
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(top+content+bottom))
		return
	}
	defer resp.Body.Close()
	data, err = io.ReadAll(resp.Body)
	var response authz.Response
	err = json.Unmarshal(data, &response)
	if err != nil {
		content = errorTmpl(c, "unable handle authz response, error", err)
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(top+content+bottom))
		return
	}
	if oreConfig.Config.Frontend.WebServer.Verbose > 0 {
		log.Printf("INFO: Authz response %+v, error %v", response, err)
	}
	if response.Status != "ok" {
		msg := fmt.Sprintf("No user %s found in Authz service", form.User)
		content = errorTmpl(c, msg, errors.New("user not found"))
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(top+content+bottom))
		return
	}

	c.Set("user", form.User)
	if oreConfig.Config.Frontend.WebServer.Verbose > 0 {
		log.Printf("login from user %s, url path %s", form.User, c.Request.URL.Path)
	}

	// set our user cookie
	if _, err := c.Cookie("user"); err != nil {
		c.SetCookie("user", form.User, 3600, "/", "localhost", false, true)
	}

	// redirect
	c.Redirect(http.StatusFound, "/")
}

// ProjectRegistrationPostHandler provides access to Post /project/registration endpoint
func ProjectRegistrationPostHandler(c *gin.Context) {
	tmpl := makeTmpl(c, "Project registration")
	top := tmplPage("top.tmpl", tmpl)
	bottom := tmplPage("bottom.tmpl", tmpl)

	// parse input form request
	var form ProjectRegistrationForm
	var err error
	content := successTmpl(c, "Project registration is successful")
	if err = c.ShouldBind(&form); err != nil {
		content = errorTmpl(c, "Project registration binding error", err)
	} else {
		if oreConfig.Config.Frontend.WebServer.Verbose > 0 {
			log.Printf("register Project %+v", form)
		}
	}

	// return page
	tmpl["Content"] = template.HTML(content)
	content = tmplPage("content.tmpl", tmpl)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(top+content+bottom))
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
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(top+content+bottom))
		return
	}
	site := form.Site
	bucket := form.Bucket
	// curl -X POST http://localhost:8340/storage/cornell/s3-bucket
	rurl := fmt.Sprintf("%s/storage/%s/%s", oreConfig.Config.Services.DataManagementURL, site, bucket)
	if oreConfig.Config.Frontend.WebServer.Verbose > 0 {
		log.Println("query DataManagement", rurl)
	}
	resp, err := httpPostForm(rurl, url.Values{})
	if err != nil {
		log.Println("ERROR:", err)
		msg := fmt.Sprintf("fail to create bucket %s at site %s, error %v", bucket, site, err)
		content := errorTmpl(c, msg, err)
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(top+content+bottom))
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
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(top+content+bottom))
		return
	}
	site := form.Site
	bucket := form.Bucket
	// curl -X DELETE http://localhost:8340/storage/cornell/s3-bucket
	rurl := fmt.Sprintf("%s/storage/%s/%s", oreConfig.Config.Services.DataManagementURL, site, bucket)
	if oreConfig.Config.Frontend.WebServer.Verbose > 0 {
		log.Println("query DataManagement", rurl)
	}
	req, err := http.NewRequest("DELETE", rurl, nil)
	if err != nil {
		log.Println("ERROR:", err)
		msg := fmt.Sprintf("fail to delete bucket %s at site %s, error %v", bucket, site, err)
		content := errorTmpl(c, msg, err)
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(top+content+bottom))
		return
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("ERROR:", err)
		msg := fmt.Sprintf("fail to delete bucket %s at site %s, error %v", bucket, site, err)
		content := errorTmpl(c, msg, err)
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(top+content+bottom))
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

// UserRegistryPostHandler provides access to POST /registry endpoint
func UserRegistryPostHandler(c *gin.Context) {
	tmpl := makeTmpl(c, "User registration")
	top := tmplPage("top.tmpl", tmpl)
	bottom := tmplPage("bottom.tmpl", tmpl)

	// parse input form request
	var form UserRegistrationForm
	var err error
	content := successTmpl(c, "User registation is completed")

	if err = c.ShouldBind(&form); err != nil {
		content = errorTmpl(c, "User registration binding error", err)
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(top+content+bottom))
		return
	}
	if oreConfig.Config.Frontend.WebServer.Verbose > 0 {
		log.Printf("new user %+v", form)
	}

	// first check if user provides the captcha
	if !captcha.VerifyString(form.CaptchaID, form.CaptchaSolution) {
		msg := "Wrong captcha match, robots are not allowed"
		content = errorTmpl(c, msg, err)
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(top+content+bottom))
		return
	}

	// encrypt form password
	form, err = encryptUserObject(form)
	if err != nil {
		content = errorTmpl(c, "unable to encrypt user password", err)
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(top+content+bottom))
		return
	}

	// make a call to Authz service to registry new user
	rurl := fmt.Sprintf("%s/user", oreConfig.Config.Services.AuthzURL)
	data, err := json.Marshal(form)
	if err != nil {
		content = errorTmpl(c, "unable to marshal user form, error", err)
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(top+content+bottom))
		return
	}
	resp, err := http.Post(rurl, "application/json", bytes.NewBuffer(data))
	if err != nil {
		content = errorTmpl(c, "unable to POST request to Authz service, error", err)
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(top+content+bottom))
		return
	}
	defer resp.Body.Close()
	data, err = io.ReadAll(resp.Body)
	var response authz.Response
	err = json.Unmarshal(data, &response)
	if err != nil {
		content = errorTmpl(c, "unable handle authz response, error", err)
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(top+content+bottom))
		return
	}
	if oreConfig.Config.Frontend.WebServer.Verbose > 0 {
		log.Printf("INFO: Authz response %+v, error %v", response, err)
	}
	if response.Status != "ok" {
		msg := fmt.Sprintf("No user %s found in Authz service", form.Login)
		content = errorTmpl(c, msg, errors.New("user not found"))
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(top+content+bottom))
		return
	}

	c.Set("user", form.Login)
	if oreConfig.Config.Frontend.WebServer.Verbose > 0 {
		log.Printf("login from user %s, url path %s", form.Login, c.Request.URL.Path)
	}

	// set our user cookie
	if _, err := c.Cookie("user"); err != nil {
		c.SetCookie("user", form.Login, 3600, "/", "localhost", false, true)
		c.Set("user", form.Login)
	}

	// return page
	// we regenerate top template with new user info
	top = tmplPage("top.tmpl", tmpl)
	// create page content
	tmpl["Content"] = template.HTML(content)
	content = tmplPage("success.tmpl", tmpl)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(top+content+bottom))
}

// SiteRegistrationPostHandler provides access to POST /site/registration endpoint
func SiteRegistrationPostHandler(c *gin.Context) {
	tmpl := makeTmpl(c, "Site registration")
	top := tmplPage("top.tmpl", tmpl)
	bottom := tmplPage("bottom.tmpl", tmpl)

	// parse input form request
	var form Site
	var err error
	content := successTmpl(c, "Site registration is successful")
	if err = c.ShouldBind(&form); err != nil {
		content = errorTmpl(c, "Site registration binding error", err)
	} else {
		if oreConfig.Config.Frontend.WebServer.Verbose > 0 {
			log.Printf("register site %+v", form)
		}

		// encrypt sensitive fields
		form, err = encryptSiteObject(form)
		if err != nil {
			content = errorTmpl(c, "Site registration failure to encrypt Site attributes", err)
		} else {
			// make JSON request to Discovery service
			if data, err := json.Marshal(form); err == nil {
				rurl := fmt.Sprintf("%s/sites", oreConfig.Config.Services.DiscoveryURL)
				resp, err := httpPost(rurl, "application/json", bytes.NewBuffer(data))
				if err != nil {
					content = errorTmpl(c, "Site registration posting to discvoeru service failure", err)
					tmpl["Content"] = template.HTML(content)
				} else {
					if oreConfig.Config.Frontend.WebServer.Verbose > 0 {
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
