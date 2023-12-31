package main

import (
	"encoding/json"
	"fmt"
	"log"

	oreConfig "github.com/OreCast/common/config"
)

// helper function to get metadata
// MetaData represents MetaData object returned from discovery service
type MetaData struct {
	ID          string   `json:"id"`
	Site        string   `json:"site"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	Bucket      string   `json:"bucket"`
}

// MetaDataRecord represents MetaData record returned by discovery service
type MetaDataRecord struct {
	Status string     `json:"status"`
	Data   []MetaData `json:"data"`
}

// helper function to fetch sites info from discovery service
func metadata(site string) MetaDataRecord {
	var results MetaDataRecord
	rurl := fmt.Sprintf("%s/meta/%s", oreConfig.Config.Services.MetaDataURL, site)
	resp, err := httpGet(rurl)
	if oreConfig.Config.Frontend.WebServer.Verbose > 0 {
		log.Println("query MetaData service rurl", rurl, err)
	}
	if err != nil {
		log.Println("ERROR:", err)
		return results
	}
	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&results); err != nil {
		log.Println("ERROR:", err)
		return results
	}
	return results
}

// helper function to fetch sites info from discovery service
func getMetaRecord(mid string) MetaDataRecord {
	var results MetaDataRecord
	rurl := fmt.Sprintf("%s/meta/record/%s", oreConfig.Config.Services.MetaDataURL, mid)
	resp, err := httpGet(rurl)
	if oreConfig.Config.Frontend.WebServer.Verbose > 0 {
		log.Println("query MetaData service rurl", rurl, err)
	}
	if err != nil {
		log.Println("ERROR:", err)
		return results
	}
	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&results); err != nil {
		log.Println("ERROR:", err)
		return results
	}
	return results
}
