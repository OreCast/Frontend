package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// helper function to get metadata
// MetaData represents MetaData object returned from discovery service
type MetaData struct {
	Site        string   `json:"site"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
}

// helper function to fetch sites info from discovery service
func metadata() []MetaData {
	var results []MetaData
	rurl := fmt.Sprintf("%s/meta", Config.MetaDataURL)
	resp, err := http.Get(rurl)
	log.Println("### rurl", rurl, err)
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
