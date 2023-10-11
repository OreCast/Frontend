package main

import (
	"encoding/json"
	"fmt"
	"log"

	oreConfig "github.com/OreCast/common/config"
)

// {"create_by":"OreCast-workflow","creation_date":1696853600,"dataset":"/a/b/c","last_modification_date":1696853600,"last_modified_by":"OreCast-workflow","meta_id":"123xyz","parent":null,"processing":"glibc","site":"Cornell"}
type DBSRecord struct {
	Dataset              string `json:"dataset"`
	MetaId               string `json:"meta_id"`
	Parent               string `json:"parent"`
	Processing           string `json:"processing"`
	Site                 string `json:"site"`
	CreateBy             string `json:"create_by"`
	CreationDate         int64  `json:"creation_date"`
	LastModifiedBy       string `json:"last_modified_by"`
	LastModificationdate int64  `json:"last_modification_date"`
}

func getDatasets(ds string) []DBSRecord {
	var datasets []DBSRecord
	rurl := fmt.Sprintf("%s/datasets", oreConfig.Config.Services.DataBookkeepingURL)
	if ds != "" {
		rurl = fmt.Sprintf("%s/dataset/%s", oreConfig.Config.Services.DataBookkeepingURL, ds)
	}
	resp, err := httpGet(rurl)
	if oreConfig.Config.Frontend.WebServer.Verbose > 0 {
		log.Println("query DataBookkeeping service rurl", rurl, err)
	}
	if err != nil {
		log.Println("ERROR:", err)
		return datasets
	}
	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&datasets); err != nil {
		log.Println("ERROR:", err)
		return datasets
	}
	return datasets

}
