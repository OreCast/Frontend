package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	cryptoutils "github.com/vkuznet/cryptoutils"
)

// Site represents Site object returned from discovery service
type Site struct {
	Name        string `json:"name"`
	URL         string `json:"url"`
	Description string `json:"description"`
}

// helper function to fetch sites info from discovery service
func sites() []Site {
	var out []Site
	rurl := fmt.Sprintf("%s/sites", Config.DiscoveryURL)
	resp, err := http.Get(rurl)
	if err != nil {
		log.Println("ERROR:", err)
		return out
	}
	defer resp.Body.Close()
	var results []Site
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&results); err != nil {
		log.Println("ERROR:", err)
		return out
	}
	return results
}

// SiteObject represents site object
type SiteObject struct {
	Name     string
	Datasets []string
}

// DiscoveryRecord represents structure of data discovery record
type DiscoveryRecord struct {
	Name         string `json:"name"`
	URL          string `json:"url"`
	Endpoint     string `json:"endpoint""`
	AccessKey    string `json:"access_key"`
	AccessSecret string `json:"access_secret"`
	UseSSL       bool   `json:"use_ssl"`
}

func site(site, bucket string) SiteObject {
	surl := fmt.Sprintf("%s/sites", Config.DiscoveryURL)
	if Config.Verbose > 0 {
		log.Println("query", surl)
	}
	resp, err := http.Get(surl)
	var siteObj SiteObject
	if err != nil {
		log.Printf("ERROR: unable to contact DataDiscovery service %s, error %v", surl, err)
		return siteObj
	}
	// read data discovery content
	var records []DiscoveryRecord
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("ERROR: unable to read DataDiscovery response, error %v", err)
		return siteObj
	}
	err = json.Unmarshal(body, &records)
	if err != nil {
		log.Printf("ERROR: unable to unmarshal DataDiscovery response, error %v", err)
		return siteObj
	}
	if Config.Verbose > 0 {
		log.Printf("site records %+v", records)
	}

	for _, rec := range records {
		if rec.Name == site {
			log.Printf("INFO: found %s in DataDiscovery records, will access its s3 via %s", rec.Name, rec.URL)
			// bingo: we got desired site, now we can query its s3 storage for datasets
			log.Println("###", rec.AccessKey, Config.DiscoveryPassword, Config.DiscoveryCipher)
			akey, err := decrypt(rec.AccessKey, Config.DiscoveryPassword, Config.DiscoveryCipher)
			if err != nil {
				log.Printf("ERROR: unable to decrypt data discovery access key, error %v", err)
				return siteObj

			}
			apwd, err := decrypt(rec.AccessSecret, Config.DiscoveryPassword, Config.DiscoveryCipher)
			if err != nil {
				log.Printf("ERROR: unable to decrypt data discovery acess secret, error %v", err)
				return siteObj

			}
			s3 := S3{
				Endpoint:     rec.Endpoint,
				AccessKey:    string(akey),
				AccessSecret: string(apwd),
				UseSSL:       rec.UseSSL,
			}
			if Config.Verbose > 0 {
				log.Printf("### will access %+v", s3)
			}
			obj := SiteObject{
				Name:     site,
				Datasets: datasets(s3, bucket),
			}
			return obj
		}
	}
	return siteObj
}

// helper function to decrypt site hex encoded cipher string
func decrypt(entry, salt, cipher string) (string, error) {
	src := []byte(entry)
	dst := make([]byte, hex.DecodedLen(len(src)))
	n, err := hex.Decode(dst, src)
	if err != nil {
		return "", err
	}
	data, err := cryptoutils.Decrypt(dst[:n], salt, cipher)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
