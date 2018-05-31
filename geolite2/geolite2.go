package geolite2

import (
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/oschwald/geoip2-golang"
	log "github.com/sirupsen/logrus"
	"github.com/xin053/ipd/config"
	"github.com/xin053/ipd/server"
)

var (
	DB2 *geoip2.Reader
)

// IPdDb2 struct implement server.Server for geolite2 method
type IPdDb2 struct{}

// FromDb2 get ip information from Geolite2 database
func FromDb2(c *gin.Context) {
	server.FindIPs(c, &IPdDb2{})
}

// FindIP find information about a single IP
func (i *IPdDb2) FindIP(ip string, ch chan config.IPWithGeo, wg *sync.WaitGroup) *config.IPWithGeo {
	if wg != nil {
		defer wg.Done()
	}

	ipWithGeo := &config.IPWithGeo{}
	ipWithGeo.IP = ip

	record, err := DB2.City(net.ParseIP(ip))
	if err != nil {
		log.Error(err)
		return ipWithGeo
	}

	if len(record.Subdivisions) == 0 {
		return ipWithGeo
	}
	region := record.Subdivisions[0].Names["zh-CN"]
	region = strings.Replace(region, "省", "", -1)
	region = strings.Replace(region, "市", "", -1)
	region = strings.Replace(region, "自治区", "", -1)
	ipWithGeo = &config.IPWithGeo{
		config.IPInfo{
			IP:      ip,
			Country: record.Country.Names["zh-CN"],
			Region:  region,
			City:    record.City.Names["zh-CN"],
		},
		config.Geo{
			GeoX: record.Location.Longitude,
			GeoY: record.Location.Latitude,
		},
	}

	if ch != nil {
		ch <- *ipWithGeo
	}
	return ipWithGeo
}

// Register register server
func (i *IPdDb2) Register(routine bool) {
	server.Register(i, routine)
}

func init() {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Error(err)
	}
	// 打开ip数据库文件, 通过监听 ctrl+c, kill -2等信号来关闭文件
	DB2, err = geoip2.Open(filepath.Join(dir, "GeoLite2-City.mmdb"))
	if err != nil {
		log.Error(err)
		return
	}
}
