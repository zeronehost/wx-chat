package wxchat

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"
	"wxchat/utils"
)

type pushLoginResponse struct {
	Msg  string
	Ret  string
	Uuid string
}

func (wx *WxChat) beginLogin() error {
	var err error
	wx.Uuid, wx.baseRequest, wx.passTicket, wx.httpClient.Cookies, wx.host, err = wx.storage.getData()
	if err != nil {
		wx.Uuid, err = wx.getUuid()
		if err != nil {
			return err
		}

		wx.triggerGenUuidEvent(wx.Uuid)
		wx.logger.Info("Uuid=" + wx.Uuid)
		err = wx.GetLoginQrcode(wx.Uuid)
		if err != nil {
			return err
		}
		tip := 1
		redirectUrl := ""
		for {
			status, result, err := wx.isAuth(tip)
			if err != nil {
				wx.logger.Error("GetRedirectUrl Error :" + err.Error())
				time.Sleep(time.Second * time.Duration(1))
				continue
			}

			if 200 == status {
				redirectUrl = result
				wx.logger.Info("Redirect=" + redirectUrl)
				wx.triggerConfirmAuthEvent(redirectUrl)
				break
			}

			if 201 == status {
				tip = 0
				wx.logger.Info("Scan Code")
				wx.triggerScanCodeEvent(result)
			}
		}

		wx.host = utils.GetHostByUrl(redirectUrl)
		err = wx.doLogin(redirectUrl)
		if err != nil {
			return err
		}

		wx.storage.setData(wx.Uuid, wx.baseRequest, wx.passTicket, wx.httpClient.Cookies, wx.host)
	}

	wx.logger.Info("Login.")
	wx.triggerLoginEvent(wx.baseRequest.DeviceID)

	return nil
}

// 获取Uuid
func (wx *WxChat) getUuid() (string, error) {

	getUuidApiUrl := wxChatApi["getUuidApi"] + utils.GetUnixMsTime()
	content, err := wx.httpClient.get(getUuidApiUrl, time.Second*5, &httpHeader{
		Host:    "login.wx2.qq.com",
		Referer: "https://wx2.qq.com/?&lang=zh_CN",
	})
	if err != nil {
		return "", err
	}

	reg, err := regexp.Compile(`window.QRLogin.code = 200; window.QRLogin.uuid = "(.+)"`)
	if err != nil {
		return "", err
	}

	uuidArr := reg.FindSubmatch([]byte(content))
	if len(uuidArr) != 2 {
		return "", errors.New("Uuid get failed")
	}

	return string(uuidArr[1]), nil
}

func (wx *WxChat) GetLoginQrcode(uuid string) error {

	getQrCodeUrl := wxChatApi["qrcodeApi"] + uuid
	content, err := wx.httpClient.get(getQrCodeUrl, time.Second*5, &httpHeader{
		Host: "login.weixin.qq.com",
	})
	if err != nil {
		return err
	}
	//fmt.Println(content)
	return ioutil.WriteFile("./qrcode.png", []byte(content), 0700)
}

// 判断是否已授权登陆,获取redirectUrl
func (wx *WxChat) isAuth(tip int) (int, string, error) {

	loginPollApi := strings.Replace(wxChatApi["loginApi"], "{uuid}", wx.Uuid, 1)
	loginPollApi = strings.Replace(loginPollApi, "{tip}", strconv.Itoa(tip), 1)
	loginPollApi = strings.Replace(loginPollApi, "{time}", utils.GetUnixMsTime(), 1)

	content, err := wx.httpClient.get(loginPollApi, time.Second*30, &httpHeader{
		Host:    "login.wx2.qq.com",
		Referer: "https://wx2.qq.com/?&lang=zh_CN",
	})
	if err != nil {
		return 0, "", err
	}

	regRedirectUri, err := regexp.Compile(`window.redirect_uri="(.+)";`)
	if err != nil {
		return 0, "", err
	}

	redirectUriArr := regRedirectUri.FindSubmatch([]byte(content))
	if len(redirectUriArr) == 2 {
		return 200, string(redirectUriArr[1]), nil
	}

	regScanCode, err := regexp.Compile(`window.code=201;window.userAvatar = '(.+)';`)
	if err != nil {
		return 0, "", err
	}

	userAvatarArr := regScanCode.FindSubmatch([]byte(content))
	if len(userAvatarArr) == 2 {
		return 201, string(userAvatarArr[1]), nil
	}

	return 0, "", nil
}

// 请求redirectUrl 登录
func (wx *WxChat) doLogin(redirectUrl string) error {
	content, err := wx.httpClient.get(redirectUrl+"&fun=new&version=v2&lang=zh_CN", time.Second*5, &httpHeader{
		Host:    wx.host,
		Referer: "https://" + wx.host + "/?&lang=zh_CN",
	})
	if err != nil {
		return err
	}

	var max int64 = 999999999999999
	var min int64 = 100000000000000
	wx.baseRequest.DeviceID = "e" + strconv.Itoa(int(rand.Int63n(max-min)+min))
	wx.baseRequest.Sid, wx.baseRequest.Uin, wx.baseRequest.Skey, wx.passTicket, err = utils.AnalysisLoginXml(content)

	return err
}

// 绑定登录
func (wx *WxChat) pushLogin() (string, error) {
	pushLoginApi := strings.Replace(wxChatApi["pushLoginApi"], "{uin}", wx.baseRequest.Uin, 1)
	content, err := wx.httpClient.get(pushLoginApi, time.Second*5, &httpHeader{
		Host:    wx.host,
		Referer: "https://" + wx.host + "/?&lang=zh_CN",
	})
	if err != nil {
		return "", err
	}

	var pushLoginResp pushLoginResponse
	err = json.Unmarshal([]byte(content), &pushLoginResp)
	if err != nil {
		return "", err
	}

	if pushLoginResp.Ret != "0" || "" == pushLoginResp.Uuid {
		return "", errors.New("Push Login Failed")
	}

	return pushLoginResp.Uuid, nil
}

// 退出
func (wx *WxChat) Logout() {

}
