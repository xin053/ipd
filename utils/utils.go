package utils

import (
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/xin053/ipd/config"
)

// GetDurationInMillseconds takes a start time and returns a duration in milliseconds
func GetDurationInMillseconds(start time.Time) float64 {
	end := time.Now()
	duration := end.Sub(start)
	milliseconds := float64(duration) / float64(time.Millisecond)
	rounded := float64(int(milliseconds*100+0.5)) / 100
	return rounded
}

// GetClientIP gets the correct IP for the end client instead of the proxy
func GetClientIP(c *gin.Context) string {
	// first check the X-Forwarded-For header
	requester := c.Request.Header.Get("X-Forwarded-For")
	// if empty, check the Real-IP header
	if len(requester) == 0 {
		requester = c.Request.Header.Get("X-Real-IP")
	}
	// if the requester is still empty, use the hard-coded address from the socket
	if len(requester) == 0 {
		requester = c.Request.RemoteAddr
	}

	// if requester is a comma delimited list, take the first one
	// (this happens when proxied via elastic load balancer then again through nginx)
	if strings.Contains(requester, ",") {
		requester = strings.Split(requester, ",")[0]
	}

	return requester
}

// EmptyStrings whether all the given strings are empty or not
func EmptyStrings(s ...string) bool {
	for _, i := range s {
		if i != "" {
			return false
		}
	}
	return true
}

// AddGeo add geo information to IPInfo struct
func AddGeo(ipInfo *config.IPInfo) *config.IPWithGeo {
	switch {
	case ipInfo.City != "" && ipInfo.City != "XX" && ipInfo.City != "0":
		ipInfo.City = strings.Replace(ipInfo.City, "市", "", -1)
		if geo, ok := config.GeoMap[ipInfo.City]; !ok {
			log.Warn("can not get geo information of: ", ipInfo.City, ipInfo.Region, ipInfo.Country, ipInfo.IP)
			return &config.IPWithGeo{*ipInfo, config.Geo{0, 0}}
		} else {
			ipInfo.Country = "中国"
			ipInfo.Region = strings.Replace(ipInfo.Region, "省", "", -1)
			ipInfo.Region = strings.Replace(ipInfo.Region, "市", "", -1)
			ipInfo.Region = strings.Replace(ipInfo.Region, "自治区", "", -1)
			return &config.IPWithGeo{*ipInfo, geo}
		}
	case ipInfo.Region != "":
		ipInfo.Region = strings.Replace(ipInfo.Region, "省", "", -1)
		ipInfo.Region = strings.Replace(ipInfo.Region, "市", "", -1)
		ipInfo.Region = strings.Replace(ipInfo.Region, "自治区", "", -1)
		if geo, ok := config.GeoMap[ipInfo.Region]; !ok {
			log.Warn("can not get geo information of: ", ipInfo.Region, ipInfo.Country, ipInfo.IP)
			return &config.IPWithGeo{*ipInfo, config.Geo{0, 0}}
		} else {
			ipInfo.Country = "中国"
			return &config.IPWithGeo{*ipInfo, geo}
		}
	case ipInfo.Country != "":
		if ipInfo.Country == "中国" || strings.ToLower(ipInfo.Country) == "china" {
			log.Warn("this chinese ip has no region or city: ", ipInfo.IP)
		}
		return &config.IPWithGeo{*ipInfo, config.Geo{0, 0}}
	default:
		return &config.IPWithGeo{*ipInfo, config.Geo{0, 0}}
	}
}

// GetCurrentPath get current caller directory
func GetCurrentPath() string {
	_, filename, _, _ := runtime.Caller(1)

	return path.Dir(filename)
}
