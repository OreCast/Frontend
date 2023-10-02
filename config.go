package main

// config module
//
// Copyright (c) 2023 - Valentin Kuznetsov <vkuznet@gmail.com>
//

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// OAuthRecord defines OAuth provider's credentials
type OAuthRecord struct {
	Provider     string `json:"provider"`      // name of the provider
	ClientID     string `json:"client_id"`     // client id
	ClientSecret string `json:"client_secret"` // client secret
}

// Configuration stores server configuration parameters
type Configuration struct {
	// web server parts
	Base        string `json:"base"`         // base URL
	LogFile     string `json:"log_file"`     // server log file
	Port        int    `json:"port"`         // server port number
	Verbose     int    `json:"verbose"`      // verbose output
	StaticDir   string `json:"static_dir"`   // speficy static dir location
	RedirectURL string `json:"redirect_url"` // redirect URL for OAuth provider

	// OAuth parts
	OAuth []OAuthRecord `json:"oauth"` // oauth configurations

	// proxy parts
	XForwardedHost      string `json:"X-Forwarded-Host"`       // X-Forwarded-Host field of HTTP request
	XContentTypeOptions string `json:"X-Content-Type-Options"` // X-Content-Type-Options option

	// server parts
	RootCAs       string   `json:"rootCAs"`      // server Root CAs path
	ServerCrt     string   `json:"server_cert"`  // server certificate
	ServerKey     string   `json:"server_key"`   // server certificate
	DomainNames   []string `json:"domain_names"` // LetsEncrypt domain names
	LimiterPeriod string   `json:"rate"`         // limiter rate value

	// captcha parts
	CaptchaSecretKey string `json:"captchaSecretKey"` // re-captcha secret key
	CaptchaPublicKey string `json:"captchaPublicKey"` // re-captcha public key
	CaptchaVerifyUrl string `json:"captchaVerifyUrl"` // re-captcha verify url

	// OreCast parts
	DiscoveryPassword string `json:"discovery_secret"`    // data-discovery password
	DiscoveryCipher   string `json:"discovery_cipher"`    // data-discovery cipher
	DiscoveryURL      string `json:"discovery_url"`       // data-discovery URL
	MetaDataURL       string `json:"metadata_url"`        // meta-data service URL
	DataManagementURL string `json:"datamanagement_url"`  // data-management service URL
	AuthzURL          string `json:"authz_url"`           // Authz service URL
	AuthzClientId     string `json:"authz_client_id"`     // client id of OAuth
	AuthzClientSecret string `json:"authz_client_secret"` // client secret of OAuth
}

// Credentials returns provider OAuth credential record
func (c Configuration) Credentials(provider string) (OAuthRecord, error) {
	for _, rec := range c.OAuth {
		if rec.Provider == provider {
			return rec, nil
		}
	}
	msg := fmt.Sprintf("No OAuth provider %s is found", provider)
	return OAuthRecord{}, errors.New(msg)
}

// Config variable represents configuration object
var Config Configuration

// helper function to parse server configuration file
func parseConfig(configFile string) error {
	data, err := os.ReadFile(filepath.Clean(configFile))
	if err != nil {
		log.Println("WARNING: Unable to read", err)
	} else {
		err = json.Unmarshal(data, &Config)
		if err != nil {
			log.Println("ERROR: Unable to parse", err)
			return err
		}
	}

	// default values
	if Config.Port == 0 {
		Config.Port = 8344
	}
	if Config.LimiterPeriod == "" {
		Config.LimiterPeriod = "100-S"
	}
	if Config.StaticDir == "" {
		cdir, err := os.Getwd()
		if err == nil {
			Config.StaticDir = fmt.Sprintf("%s/static", cdir)
		} else {
			Config.StaticDir = "static"
		}
	}
	if Config.DiscoveryCipher == "" {
		Config.DiscoveryCipher = "aes"
	}
	if Config.DiscoveryURL == "" {
		Config.DiscoveryURL = "http://localhost:8320"
	}
	if Config.MetaDataURL == "" {
		Config.MetaDataURL = "http://localhost:8300"
	}
	if Config.RedirectURL == "" {
		if host, err := os.Hostname(); err == nil {
			Config.RedirectURL = fmt.Sprintf("http://%s:%d%s/github/callback", host, Config.Port, Config.Base)
		} else {
			Config.RedirectURL = fmt.Sprintf("http://localhost:%d%s/github/callback", Config.Port, Config.Base)
		}
	}
	log.Println("DiscoveryURL", Config.DiscoveryURL)
	log.Println("MetaDataURL", Config.MetaDataURL)
	log.Println("RedirectURL", Config.RedirectURL)
	return nil
}
