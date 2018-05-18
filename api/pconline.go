package api

import (
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/axgle/mahonia"
	"github.com/xin053/ipd/config"
)

type Pconline struct {
	IP      string `json:"ip"`
	Country string `json:"country"`
	Region  string `json:"pro"`
	City    string `json:"city"`
	ISP     string `json:"addr"`
}

func (p *Pconline) Name() string {
	return "太平洋"
}

func (p *Pconline) Url() string {
	return "http://whois.pconline.com.cn/ipJson.jsp?ip="
}

func (p *Pconline) Request(ip string) (API, error) {
	url := p.Url() + ip
	client := http.Client{
		Timeout: config.APITimeOut,
	}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	result := strings.Split(strings.Split(string(body[:]), "(")[2], ")")[0]

	enc := mahonia.NewDecoder("gbk")
	result = enc.ConvertString(result)

	pconline := Pconline{}
	json.Unmarshal([]byte(result), &pconline)

	pconline.IP = ip
	if pconline.ISP != "" && len(strings.Split(pconline.ISP, " ")) == 2 {
		pconline.ISP = strings.Split(pconline.ISP, " ")[1]
	} else {
		pconline.ISP = ""
	}
	return &pconline, nil
}

func (p *Pconline) JSON() config.IPInfo {
	return config.IPInfo{p.IP, p.Country, p.Region, p.City, p.ISP}
}
