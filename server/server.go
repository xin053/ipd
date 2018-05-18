package server

import (
	"net"
	"sync"

	"github.com/allegro/bigcache"
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	log "github.com/sirupsen/logrus"
	"github.com/xin053/ipd/config"
	"github.com/xin053/ipd/es"
)

var (
	json       = jsoniter.ConfigCompatibleWithStandardLibrary
	ServerList = [2][]Server{}
)

// Server you can register your own ip recognition service by implement this interface
type Server interface {
	// FindIP find information about a single IP
	FindIP(ip string, ch chan config.IPWithGeo, wg *sync.WaitGroup) *config.IPWithGeo
	// Register register server
	Register(routine bool)
}

// Register register server, only one server can set the routine param to true
func Register(server Server, routine bool) {
	if routine {
		if len(ServerList[0]) == 1 {
			log.Warn("you already have a server set to true at ServerList, this server will be set false.")
			ServerList[1] = append(ServerList[1], server)
			return
		}
		ServerList[0] = append(ServerList[0], server)
	} else {
		ServerList[1] = append(ServerList[1], server)
	}
}

func solveIP(ips []string) ([]string, []config.IPWithGeo) {
	var ipWithGeoCache = []config.IPWithGeo{}
	var ipList []string
	for _, ip := range ips {
		if net.ParseIP(ip) == nil {
			log.Warnf("'%s' is not a valid ip, you should pay attention to this ip", ip)
			continue
		}
		if entry, err := config.Cache.Get(ip); err != nil {
			if err, ok := err.(*bigcache.EntryNotFoundError); ok {
			} else {
				log.Error(err)
			}
		} else {
			var data config.IPWithGeo
			if err := json.Unmarshal(entry, &data); err != nil {
				log.Error(err)
			}
			ipWithGeoCache = append(ipWithGeoCache, data)
			continue
		}
		ipList = append(ipList, ip)
	}
	return ipList, ipWithGeoCache
}

func cacheIP(ipWithGeoList []config.IPWithGeo) {
	for _, ipWithGeo := range ipWithGeoList {
		ipWithGeoJSON, err := json.Marshal(ipWithGeo)
		if err != nil {
			log.Error(err)
		}
		config.Cache.Set(ipWithGeo.IP, ipWithGeoJSON)
	}
}

// FindIPs get information about a list of ip by server asynchronously
func FindIPs(c *gin.Context, server Server) {
	var ipJSON config.IP
	err := c.BindJSON(&ipJSON)
	if err != nil {
		log.Error("missing ip params or format incorrect")
		return
	}

	ipList, ipWithGeoCache := solveIP(ipJSON.IP)

	var wg sync.WaitGroup
	var ch = make(chan config.IPWithGeo, config.IPChanelBuf)
	wg.Add(len(ipList))
	for _, ip := range ipList {
		go server.FindIP(ip, ch, &wg)
	}
	wg.Wait()
	close(ch)

	var ipWithGeoList = []config.IPWithGeo{}
	for ipWithGeo := range ch {
		ipWithGeoList = append(ipWithGeoList, ipWithGeo)
	}

	message, err := json.Marshal(append(ipWithGeoCache, ipWithGeoList...))
	if err != nil {
		log.Error(err)
		return
	}

	c.String(200, string(message))

	// cache these ip
	go cacheIP(ipWithGeoList)

	// store these ip
	if config.StoreES {
		go es.Store(ipWithGeoList...)
	}
}

// GetIP get information about a list of ip by ServerList
// first, ip2region asynchronously; then chunzhe; then geoip2; at last, query by api
func GetIP(c *gin.Context) {
	var ipJSON config.IP
	err := c.BindJSON(&ipJSON)
	if err != nil {
		log.Error("missing ip params or format incorrect")
		return
	}

	ipList, ipWithGeoCache := solveIP(ipJSON.IP)

	var wg sync.WaitGroup
	var ch = make(chan config.IPWithGeo, config.IPChanelBuf)
	wg.Add(len(ipList))
	ipMap := map[string]bool{}
	for _, ip := range ipList {
		ipMap[ip] = false
		go ServerList[0][0].FindIP(ip, ch, &wg)
	}
	wg.Wait()
	close(ch)

	var ipWithGeoList = []config.IPWithGeo{}
	for ipWithGeo := range ch {
		ipMap[ipWithGeo.IP] = true
		if ipWithGeo.GeoX == 0 && ipWithGeo.GeoY == 0 {
			ipWithGeo = *fromList1(ipWithGeo.IP)
		}
		ipWithGeoList = append(ipWithGeoList, ipWithGeo)
	}

	// things left
	for ip, exits := range ipMap {
		if !exits {
			ipWithGeo := fromList1(ip)
			ipWithGeoList = append(ipWithGeoList, *ipWithGeo)
		}
	}

	message, err := json.Marshal(append(ipWithGeoCache, ipWithGeoList...))
	if err != nil {
		log.Error(err)
		return
	}

	c.String(200, string(message))

	// cache these ip
	go cacheIP(ipWithGeoList)

	// store these ip
	if config.StoreES {
		go es.Store(ipWithGeoList...)
	}
}

func fromList1(ip string) *config.IPWithGeo {
	ipWithGeo := &config.IPWithGeo{}
	ipWithGeo.IP = ip
	for _, server := range ServerList[1] {
		ipWithGeo = server.FindIP(ip, nil, nil)
		if ipWithGeo.GeoX != 0 || ipWithGeo.GeoY != 0 {
			return ipWithGeo
		}
	}
	return ipWithGeo
}
