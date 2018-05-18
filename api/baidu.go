package api

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/json-iterator/go"
	"github.com/xin053/ipd/config"
)

type BaiDu struct {
	IP      string `json:"ip"`
	Country string `json:"country"`
	Region  string `json:"province"`
	City    string `json:"city"`
	ISP     string `json:"isp"`
}

func (b *BaiDu) Name() string {
	return "百度"
}

func (b *BaiDu) Url() string {
	return fmt.Sprintf("http://api.map.baidu.com/location/ip?ak=%s&ip=", config.BaiDuAK)
}

func (b *BaiDu) Request(ip string) (API, error) {
	url := b.Url() + ip
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Referer", config.BaiDuReferer)

	client := http.Client{
		Timeout: config.APITimeOut,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	baidu := BaiDu{}
	json.Unmarshal([]byte(jsoniter.Get(body, "content", "address_detail").ToString()), &baidu)
	baidu.IP = ip
	return &baidu, nil
}

func (b *BaiDu) JSON() config.IPInfo {
	return config.IPInfo{b.IP, b.Country, b.Region, b.City, b.ISP}
}
