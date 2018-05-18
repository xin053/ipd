package qqwry

import (
	"encoding/binary"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/axgle/mahonia"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/xin053/ipd/config"
	"github.com/xin053/ipd/server"
	"github.com/xin053/ipd/utils"
)

const (
	INDEX_LEN       = 7
	REDIRECT_MODE_1 = 0x01
	REDIRECT_MODE_2 = 0x02
)

var gqqwry *QQwry

type QQwry struct {
	buff  []byte
	start uint32
	end   uint32
}

// IPdDb struct implement server.Server for chunzhen method
type IPdDb struct{}

func NewQQwry(file string) (qqwry *QQwry) {
	qqwry = &QQwry{}
	f, e := os.Open(file)
	if e != nil {
		log.Error(e)
		return nil
	}
	defer f.Close()
	qqwry.buff, e = ioutil.ReadAll(f)
	if e != nil {
		log.Error(e)
		return nil
	}
	qqwry.start = binary.LittleEndian.Uint32(qqwry.buff[:4])
	qqwry.end = binary.LittleEndian.Uint32(qqwry.buff[4:8])
	return qqwry
}

// FromDb get ip information from chunzhen ip database
func FromDb(c *gin.Context) {
	server.FindIPs(c, &IPdDb{})
}

// FindIP find information about a single IP
func (i *IPdDb) FindIP(ip string, ch chan config.IPWithGeo, wg *sync.WaitGroup) *config.IPWithGeo {
	if wg != nil {
		defer wg.Done()
	}

	ipWithGeo := &config.IPWithGeo{}
	ipWithGeo.IP = ip
	if gqqwry.buff == nil {
		return ipWithGeo
	}

	var country []byte
	var area []byte
	ip1 := net.ParseIP(ip)
	if ip1 == nil {
		return ipWithGeo
	}
	offset := gqqwry.searchRecord(binary.BigEndian.Uint32(ip1.To4()))
	if offset <= 0 {
		return ipWithGeo
	}
	mode := gqqwry.readMode(offset + 4)
	if mode == REDIRECT_MODE_1 {
		countryOffset := gqqwry.readUint32FromByte3(offset + 5)

		mode = gqqwry.readMode(countryOffset)
		if mode == REDIRECT_MODE_2 {
			c := gqqwry.readUint32FromByte3(countryOffset + 1)
			country = gqqwry.readString(c)
			countryOffset += 4
			area = gqqwry.readArea(countryOffset)

		} else {
			country = gqqwry.readString(countryOffset)
			countryOffset += uint32(len(country) + 1)
			area = gqqwry.readArea(countryOffset)
		}

	} else if mode == REDIRECT_MODE_2 {
		countryOffset := gqqwry.readUint32FromByte3(offset + 5)
		country = gqqwry.readString(countryOffset)
		area = gqqwry.readArea(offset + 8)
	}
	enc := mahonia.NewDecoder("gbk")
	region := enc.ConvertString(string(country))

	ipInfo := config.IPInfo{IP: ip}
	ipInfo.ISP = enc.ConvertString(string(area))

	if strings.Contains(region, "市") || strings.Contains(region, "省") || strings.Contains(region, "区") {
		ipInfo.Country = "中国"
		if strings.Contains(region, "省") {
			if s := strings.Split(region, "省"); len(s) == 2 {
				ipInfo.Region = s[0]
				ipInfo.City = s[1]
			} else {
				ipInfo.Region = s[0]
			}
		}
	} else {
		ipInfo.Country = region
	}

	ipWithGeo = utils.AddGeo(&ipInfo)
	if ch != nil {
		ch <- *ipWithGeo
	}
	return ipWithGeo
}

// Register register server
func (i *IPdDb) Register(routine bool) {
	server.Register(i, routine)
}

func (q *QQwry) readUint32FromByte3(offset uint32) uint32 {
	return byte3ToUInt32(q.buff[offset : offset+3])
}

func (q *QQwry) readMode(offset uint32) byte {
	return q.buff[offset : offset+1][0]
}

func (q *QQwry) readString(offset uint32) []byte {

	i := 0
	for {

		if q.buff[int(offset)+i] == 0 {
			break
		} else {
			i++
		}

	}
	return q.buff[offset : int(offset)+i]
}

func (q *QQwry) readArea(offset uint32) []byte {
	mode := q.readMode(offset)
	if mode == REDIRECT_MODE_1 || mode == REDIRECT_MODE_2 {
		areaOffset := q.readUint32FromByte3(offset + 1)
		if areaOffset == 0 {
			return []byte("")
		} else {
			return q.readString(areaOffset)
		}
	} else {
		return q.readString(offset)
	}
}

func (q *QQwry) getRecord(offset uint32) []byte {
	return q.buff[offset : offset+INDEX_LEN]
}

func (q *QQwry) getIPFromRecord(buf []byte) uint32 {
	return binary.LittleEndian.Uint32(buf[:4])
}

func (q *QQwry) getAddrFromRecord(buf []byte) uint32 {
	return byte3ToUInt32(buf[4:7])
}

func (q *QQwry) searchRecord(ip uint32) uint32 {

	start := q.start
	end := q.end

	// log.Printf("len info %v, %v ---- %v, %v", start, end, hex.EncodeToString(header[:4]), hex.EncodeToString(header[4:]))
	for {
		mid := q.getMiddleOffset(start, end)
		buf := q.getRecord(mid)
		_ip := q.getIPFromRecord(buf)

		// log.Printf(">> %v, %v, %v -- %v", start, mid, end, hex.EncodeToString(buf[:4]))

		if end-start == INDEX_LEN {
			//log.Printf(">> %v, %v, %v -- %v", start, mid, end, hex.EncodeToString(buf[:4]))
			offset := q.getAddrFromRecord(buf)
			buf = q.getRecord(mid + INDEX_LEN)
			if ip < q.getIPFromRecord(buf) {
				return offset
			} else {
				return 0
			}
		}

		// 找到的比较大，向前移
		if _ip > ip {
			end = mid
		} else if _ip < ip { // 找到的比较小，向后移
			start = mid
		} else if _ip == ip {
			return byte3ToUInt32(buf[4:7])
		}

	}
}

func (q *QQwry) getMiddleOffset(start uint32, end uint32) uint32 {
	records := ((end - start) / INDEX_LEN) >> 1
	return start + records*INDEX_LEN
}

func byte3ToUInt32(data []byte) uint32 {
	i := uint32(data[0]) & 0xff
	i |= (uint32(data[1]) << 8) & 0xff00
	i |= (uint32(data[2]) << 16) & 0xff0000
	return i
}

func init() {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Error(err)
	}
	gqqwry = NewQQwry(filepath.Join(dir, "qqwry.dat"))
}
