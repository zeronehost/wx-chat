package wxchat

import (
	"encoding/json"
	"errors"
	"strings"
	"time"
	"wxchat/utils"
)

type initRequest struct {
	BaseRequest baseRequest
}

type initResp struct {
	Response
	User    Contact
	Skey    string
	SyncKey syncKey
}

// init
func (wx *WxChat) init() error {
	wxInitApi := strings.Replace(wxChatApi["initApi"], "{r}", utils.GetUnixTime(), 1)
	wxInitApi = strings.Replace(wxInitApi, "{host}", wx.host, 1)
	wxInitApi = strings.Replace(wxInitApi, "{pass_ticket}", wx.passTicket, 1)

	postData, err := json.Marshal(initRequest{
		BaseRequest: wx.baseRequest,
	})
	if err != nil {
		return err
	}

	content, err := wx.httpClient.post(wxInitApi, postData, time.Second*5, &httpHeader{
		Accept:      "application/json, text/plain, */*",
		ContentType: "application/json;charset=UTF-8",
		Origin:      "https://" + wx.host,
		Host:        wx.host,
		Referer:     "https://" + wx.host + "/?&lang=zh_CN",
	})
	if err != nil {
		return err
	}

	var initRes initResp
	err = json.Unmarshal([]byte(content), &initRes)
	if err != nil {
		return err
	}

	if initRes.Response.BaseResponse.Ret != 0 {
		wx.logger.Error("Init Failed. Res.Ret=" + string(initRes.Response.BaseResponse.Ret))
		return errors.New("Init Failed")
	}

	wx.me = initRes.User
	wx.baseRequest.Skey = initRes.Skey
	wx.syncKey = initRes.SyncKey

	return nil
}
