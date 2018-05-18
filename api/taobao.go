package api

import (
	"io/ioutil"
	"net/http"

	"github.com/json-iterator/go"
	"github.com/xin053/ipd/config"
)

type TaoBao struct {
	IP      string `json:"ip"`
	Country string `json:"country"`
	Region  string `json:"region"`
	City    string `json:"city"`
	ISP     string `json:"isp"`
}

func (t *TaoBao) Name() string {
	return "淘宝"
}

func (t *TaoBao) Url() string {
	return "http://ip.taobao.com/service/getIpInfo.php?ip="
}

func (t *TaoBao) Request(ip string) (API, error) {
	url := t.Url() + ip
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

	taoBao := TaoBao{}
	json.Unmarshal([]byte(jsoniter.Get(body, "data").ToString()), &taoBao)
	taoBao.IP = ip
	return &taoBao, nil
}

func (t *TaoBao) JSON() config.IPInfo {
	return config.IPInfo{t.IP, t.Country, t.Region, t.City, t.ISP}
}
