package api

import (
	"math/rand"
	"sync"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/xin053/ipd/config"
	"github.com/xin053/ipd/server"
	"github.com/xin053/ipd/utils"
)

var apiList = []API{&TaoBao{}, &Sina{}, &BaiDu{}, &Pconline{}}

// IPdAPI struct implement server.Server for api method
type IPdAPI struct{}

// FromAPI get ip information from public IP API
func FromAPI(c *gin.Context) {
	server.FindIPs(c, &IPdAPI{})
}

// FindIP find information about a single IP
func (i *IPdAPI) FindIP(ip string, ch chan config.IPWithGeo, wg *sync.WaitGroup) *config.IPWithGeo {
	if wg != nil {
		defer wg.Done()
	}

	ipWithGeo := &config.IPWithGeo{}
	ipWithGeo.IP = ip

	count := 0
	var randomAPI API
	randomAPI = apiList[rand.Intn(len(apiList))]
	result, err := randomAPI.Request(ip)

	for err != nil {
		log.Error(err)
		count++
		if count >= 3 {
			return ipWithGeo
		}

		randomAPI = apiList[rand.Intn(len(apiList))]
		result, err = randomAPI.Request(ip)
	}

	ipInfo := result.JSON()
	ipWithGeo = utils.AddGeo(&ipInfo)
	if ch != nil {
		ch <- *ipWithGeo
	}
	return ipWithGeo
}

// Register register server
func (i *IPdAPI) Register(routine bool) {
	server.Register(i, routine)
}
