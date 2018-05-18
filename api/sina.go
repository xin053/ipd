package api

import (
	"io/ioutil"
	"net/http"

	"github.com/xin053/ipd/config"
)

type Sina struct {
	IP      string `json:"ip"`
	Country string `json:"country"`
	Region  string `json:"province"`
	City    string `json:"city"`
	ISP     string `json:"isp"`
}

func (s *Sina) Name() string {
	return "新浪"
}

func (s *Sina) Url() string {
	return "http://int.dpool.sina.com.cn/iplookup/iplookup.php?format=json&ip="
}

func (s *Sina) Request(ip string) (API, error) {
	url := s.Url() + ip
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

	sina := Sina{}
	json.Unmarshal(body, &sina)
	sina.IP = ip
	return &sina, nil
}

func (s *Sina) JSON() config.IPInfo {
	return config.IPInfo{s.IP, s.Country, s.Region, s.City, s.ISP}
}
