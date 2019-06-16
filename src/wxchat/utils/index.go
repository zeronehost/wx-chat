package utils

import (
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"time"
)

func GetUnixMsTime() string {
	return strconv.FormatInt(time.Now().UnixNano()/1000000, 10)
}

func GetUnixTime() string {
	return strconv.FormatInt(time.Now().Unix()/1000000, 10)
}

func CookiesToMap(cookies []*http.Cookie) map[string]string {
	res := map[string]string{}
	for _, v := range cookies {
		res[v.Name] = v.Value
	}
	return res
}

func CookiesToString(cookies []*http.Cookie) string {
	res := ""
	for _, v := range cookies {
		res = fmt.Sprintf("%s%s=%s;", res, v.Name, v.Value)
	}
	return res
}

func UserNameToId(userName string) string {
	r := ""
	for _, v := range []byte(userName) {
		r += strconv.Itoa(int(v))
	}
	return substr(r, 0, 9)
}

func substr(str string, start, end int) string {
	rs := []rune(str)
	len := len(rs)
	if start < 0 || start > len {
		panic("截取起始位置错误")
	}
	if end < 0 || end > len {
		panic("截取结束位置错误")
	}
	return string(rs[start:end])
}

// 根据url获取host
func GetHostByUrl(uri string) string {
	u, err := url.Parse(uri)
	if err != nil {
		return "wx2.qq.com"
	}
	return u.Host
}

// 解析登陆返回的xml
func AnalysisLoginXml(xmlStr string) (string, string, string, string, error) {
	type Error struct {
		Ret         string `xml:"ret"`
		Message     string `xml:"message"`
		Skey        string `xml:"skey"`
		Wxsid       string `xml:"wxsid"`
		Wxuin       string `xml:"wxuin"`
		PassTicket  string `xml:"pass_ticket"`
		Isgrayscale string `xml:"isgrayscale"`
	}
	var v Error
	err := xml.Unmarshal([]byte(xmlStr), &v)
	if err != nil {
		return "", "", "", "", err
	}
	return v.Wxsid, v.Wxuin, v.Skey, v.PassTicket, nil
}

// 解析位置图片
func GetLocationImgFromContent(content string) (string, error) {
	locationImgReg, err := regexp.Compile(`/cgi-bin/mmwebwx-bin/webwxgetpubliclinkimg?(.+)`)
	if err != nil {
		return "", err
	}
	locationImgArr := locationImgReg.FindSubmatch([]byte(content))
	if len(locationImgArr) == 2 {
		return string(locationImgArr[1]), nil
	}
	return "", errors.New("位置图片获取失败")
}

// 解析位置信息
func GetLocationInfoFromOriContent(oriContent string) (string, string, string, error) {
	reg, err := regexp.Compile(`<location x="(.*)" y="(.+)" scale="(.+)" label="(.+)" maptype="(.+)" poiname="[位置]" />`)
	if err != nil {
		return "", "", "", err
	}
	locationArr := reg.FindSubmatch([]byte(oriContent))
	if len(locationArr) == 6 {
		return string(locationArr[1]), string(locationArr[2]), string(locationArr[4]), nil
	}
	return "", "", "", errors.New("uuid获取失败")
}
