package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// Site represents Site object returned from discovery service
type Site struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// helper function to fetch sites info from discovery service
func sites() []string {
	var out []string
	rurl := fmt.Sprintf("%s/sites", Config.DiscoveryURL)
	resp, err := http.Get(rurl)
	log.Println("### rurl", rurl, err)
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
	for _, r := range results {
		out = append(out, r.Name)
	}
	return out
}

type SiteObject struct {
	Name     string
	Datasets []string
}

func site(site string) SiteObject {
	// TODO: place call to discovery service to get details of the specific site
	// so far to fake it we'll call storage()
	var uri string
	obj := SiteObject{
		Name:     site,
		Datasets: datasets(uri),
	}
	return obj
}
