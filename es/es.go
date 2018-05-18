package es

import (
	"context"
	"time"

	"github.com/olivere/elastic"
	log "github.com/sirupsen/logrus"
	"github.com/xin053/ipd/config"
)

var (
	ctx    context.Context
	Client *elastic.Client
)

type ESIPInfo struct {
	Time     string `json:"time"`
	IP       string `json:"ip"`
	Country  string `json:"country"`
	Region   string `json:"region"`
	City     string `json:"city"`
	ISP      string `json:"isp"`
	Location struct {
		Lon float64 `json:"lon"`
		Lat float64 `json:"lat"`
	} `json:"location"`
}

// Store store these ip information in elasticsearch.
func Store(ipWithGeos ...config.IPWithGeo) {
	termQuerys := []elastic.Query{}
	bulkRequest := Client.Bulk()
	indexReqs := []elastic.BulkableRequest{}
	for _, ipWithGeo := range ipWithGeos {
		termQuery := elastic.NewTermQuery("ip", ipWithGeo.IP)
		termQuerys = append(termQuerys, termQuery)

		esIPInfo := ESIPInfo{
			Time:    time.Now().Format("2006-01-02 15:04:05.999"),
			IP:      ipWithGeo.IP,
			Country: ipWithGeo.Country,
			Region:  ipWithGeo.Region,
			City:    ipWithGeo.City,
			ISP:     ipWithGeo.ISP,
		}
		esIPInfo.Location.Lon = ipWithGeo.GeoX
		esIPInfo.Location.Lat = ipWithGeo.GeoY

		indexReq := elastic.NewBulkIndexRequest().Index(config.ESIndex).Type("_doc").Doc(esIPInfo)
		indexReqs = append(indexReqs, indexReq)
	}

	if len(indexReqs) == 0 {
		return
	}

	bulkRequest = bulkRequest.Add(indexReqs...)
	// first, delete this ip info in elasticsearch
	q := elastic.NewBoolQuery().Should(termQuerys...)
	// Search with a term query
	_, err := Client.DeleteByQuery().
		Index(config.ESIndex). // search in index "ipinfo"
		Query(q).              // return all results, but ...
		// Pretty(true).       // pretty print request and response JSON
		Do(ctx) // execute
	if err != nil {
		log.Error(err)
		return
	}
	// then, we send a bulk request.
	_, err = bulkRequest.Do(ctx)
	if err != nil {
		log.Error(err)
		return
	}
	return
}

func init() {
	if config.StoreES {
		ctx = context.Background()
		var err error
		Client, err = elastic.NewClient(elastic.SetURL(config.ESURL))
		if err != nil && config.StoreES {
			log.Error(err)
		}

		exists, err := Client.IndexExists(config.ESIndex).Do(ctx)
		if err != nil {
			log.Error(err)
		}

		if !exists {
			log.Infof("elasticsearch index %s does not exits, and will be created now.", config.ESIndex)
			_, err := Client.CreateIndex(config.ESIndex).BodyString(config.ESMapping).Do(ctx)
			if err != nil {
				log.Error(err)
			} else {
				log.Infof("elasticsearch index %s has been created.", config.ESIndex)
			}
		}
	}
}
