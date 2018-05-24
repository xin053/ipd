package ip2region

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/xin053/ipd/config"
	"github.com/xin053/ipd/server"
	"github.com/xin053/ipd/utils"
)

var (
	DB3 *Ip2Region
)

// IPdDb3 struct implement server.Server for ip2region method
type IPdDb3 struct{}

// FromDb3 get ip information from ip2region database
func FromDb3(c *gin.Context) {
	server.FindIPs(c, &IPdDb3{})
}

// FindIP find information about a single IP
func (i *IPdDb3) FindIP(ip string, ch chan config.IPWithGeo, wg *sync.WaitGroup) *config.IPWithGeo {
	if wg != nil {
		defer wg.Done()
	}

	ipWithGeo := &config.IPWithGeo{}
	ipWithGeo.IP = ip

	result, err := DB3.MemorySearch(ip)
	if err != nil {
		return ipWithGeo
	}

	ipInfo := config.IPInfo{
		IP:      ip,
		Country: strings.Replace(result.Country, "0", "", -1),
		Region:  strings.Replace(result.Province, "0", "", -1),
		City:    strings.Replace(result.City, "0", "", -1),
		ISP:     strings.Replace(result.ISP, "0", "", -1),
	}

	ipWithGeo = utils.AddGeo(&ipInfo)
	if ch != nil {
		ch <- *ipWithGeo
	}
	return ipWithGeo
}

// Register register server
func (i *IPdDb3) Register(routine bool) {
	server.Register(i, routine)
}

func init() {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Error(err)
	}
	os.Chdir(dir)
	// download ip2region.db from github
	log.Info("Initializing...please wait...")
	log.Info("downloading ip2region.db from github...")
	res, err := http.Get(config.IP2RegionURL)
	if err != nil {
		log.Warn("download ip2region.db from github failed, you should take a look at this.")
		log.Info("now, we will use ip2region.db at the workplace.")
		log.Error(err)
	} else {
		// remove ip2region.db if exits
		os.Remove("ip2region.db")
		f, err := os.Create("ip2region.db")
		if err != nil {
			log.Warn("creat ip2region.db file failed, you should take a look at this.")
			log.Error(err)
		}
		io.Copy(f, res.Body)
		log.Info("download finished, you can go on")
	}

	// 打开ip数据库文件, 通过监听 ctrl+c, kill -2等信号来关闭文件
	DB3, err = New(filepath.Join(dir, "ip2region.db"))
	if err != nil {
		log.Error(err)
		return
	}
}
