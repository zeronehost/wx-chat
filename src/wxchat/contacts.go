package wxchat

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"
	"wxchat/utils"
)

// 联系人类型
type ContactType int

const (
	_        ContactType = iota
	Official             // 公众号
	Friend               // 好友
	Group                // 群组
)

// 联系人
type Contact struct {
	Uin              float64
	UserName         string
	NickName         string
	HeadImgUrl       string
	ContactFlag      float64
	MemberCount      float64
	MemberList       []*Member
	MemberMap        map[string]*Member
	RemarkName       string
	HideInputBarFlag float64
	Sex              float64
	Signature        string
	VerifyFlag       float64
	PYInitial        string
	PYQuanPin        string
	RemarkPYInitial  string
	RemarkPYQuanPin  string
	StarFriend       float64
	AppAccountFlag   float64
	Statues          float64
	AttrStatus       float64
	Province         string
	City             string
	Alias            string
	SnsFlag          float64
	UniFriend        float64
	DisplayName      string
	ChatRoomId       float64
	KeyWord          string
	EncryChatRoomId  string
	IsOwner          float64
	Type             ContactType
}

// 群组成员
type Member struct {
	Uin             float64
	UserName        string
	NickName        string
	AttrStatus      float64
	PYInitial       string
	PYQuanPin       string
	RemarkPYInitial string
	RemarkPYQuanPin string
	MemberStatus    float64
	DisplayName     string
	KeyWord         string
}

type getContactResponse struct {
	Response
	MemberCount int
	MemberList  []*Contact
	Seq         float64
}

type batchGetContactResponse struct {
	Response
	Count       int
	ContactList []*Contact
}

// 初始化通讯录
func (wx *WxChat) initContact() error {
	seq := float64(-1)

	var cts = []*Contact{}
	wx.contacts = map[string]*Contact{}

	for seq != 0 {
		if -1 == seq {
			seq = 0
		}
		contactList, s, err := wx.getContacts(seq)
		if err != nil {
			return err
		}
		seq = s
		cts = append(cts, contactList...)
	}

	var groupUserNames []string

	for _, v := range cts {
		verifyFlag := v.VerifyFlag
		userName := v.UserName

		if verifyFlag/8 != 0 {
			v.Type = Official
		} else if strings.HasPrefix(userName, "@@") {
			v.Type = Group
			groupUserNames = append(groupUserNames, userName)
		} else {
			v.Type = Friend
		}
		wx.contacts[userName] = v
	}

	groups, _ := wx.fetchContacts(groupUserNames)
	for _, group := range groups {
		group.MemberMap = map[string]*Member{}
		for _, contact := range group.MemberList {
			group.MemberMap[contact.UserName] = contact
		}
		wx.contacts[group.UserName] = group
	}

	return nil
}

// 获取联系人
func (wx *WxChat) getContacts(seq float64) ([]*Contact, float64, error) {

	getContactsApiUrl := strings.Replace(wxChatApi["getContactApi"], "{pass_ticket}", wx.passTicket, 1)
	getContactsApiUrl = strings.Replace(getContactsApiUrl, "{seq}", strconv.FormatInt(int64(seq), 10), 1)
	getContactsApiUrl = strings.Replace(getContactsApiUrl, "{skey}", wx.baseRequest.Skey, 1)
	getContactsApiUrl = strings.Replace(getContactsApiUrl, "{r}", utils.GetUnixTime(), 1)
	getContactsApiUrl = strings.Replace(getContactsApiUrl, "{host}", wx.host, 1)

	content, err := wx.httpClient.get(getContactsApiUrl, time.Second*5, &httpHeader{
		Host:    wx.host,
		Referer: "https://" + wx.host + "/?&lang=zh_CN",
	})

	var resp getContactResponse
	err = json.Unmarshal([]byte(content), &resp)
	if err != nil {
		return nil, float64(0), err
	}

	return resp.MemberList, resp.Seq, nil
}

// 获取联系人详情, 群组获取成员
func (wx *WxChat) fetchContacts(userNames []string) ([]*Contact, error) {

	var list []map[string]string

	for _, u := range userNames {
		list = append(list, map[string]string{
			"UserName":   u,
			"ChatRoomId": "",
		})
	}

	data, err := json.Marshal(map[string]interface{}{
		"BaseRequest": wx.baseRequest,
		"Count":       len(list),
		"List":        list,
	})

	if err != nil {
		return nil, err
	}

	batchGetContactApi := strings.Replace(wxChatApi["batchGetContactApi"], "{r}", utils.GetUnixMsTime(), 1)
	batchGetContactApi = strings.Replace(batchGetContactApi, "{host}", wx.host, 1)
	content, err := wx.httpClient.post(batchGetContactApi, data, time.Second*5, &httpHeader{
		Host:    wx.host,
		Referer: "https://" + wx.host + "/?&lang=zh_CN",
	})

	var resp batchGetContactResponse
	err = json.Unmarshal([]byte(content), &resp)
	if err != nil {
		return nil, err
	}

	return resp.ContactList, nil
}

// 根据UserName更新联系人
func (wx *WxChat) updateContact(userNames []string) error {

	contacts, err := wx.fetchContacts(userNames)

	if err != nil || len(contacts) != 1 {
		return errors.New("Fetch contacts failed.")
	}

	for _, contact := range contacts {
		contact.MemberMap = map[string]*Member{}
		for _, member := range contact.MemberList {
			contact.MemberMap[member.UserName] = member
		}

		if contact.VerifyFlag/8 != 0 {
			contact.Type = Official
		} else if strings.HasPrefix(contact.UserName, "@@") {
			contact.Type = Group
		} else {
			contact.Type = Friend
		}

		wx.contacts[contact.UserName] = contact
	}

	return nil
}

// 更新联系人
func (wx *WxChat) contactsModify(cts []map[string]interface{}) error {
	userNames := []string{}
	userNamesStr := ""
	for _, newContact := range cts {
		userNames = append(userNames, newContact["UserName"].(string))
		userNamesStr += newContact["UserName"].(string) + ", "
	}

	wx.logger.Notice("Contacts Modify. UserNames: " + userNamesStr)

	return wx.updateContact(userNames)
}

// 删除联系人
func (wx *WxChat) contactsDelete(cts []map[string]interface{}) {
	userNamesStr := ""
	for _, contact := range cts {
		delete(wx.contacts, contact["UserName"].(string))
		userNamesStr += contact["UserName"].(string) + ", "
	}

	wx.logger.Notice("Contacts Delete. UserNames: " + userNamesStr)

}

// 查找联系人userName
func (wx *WxChat) SearchContact(remarkName string) (string, error) {
	userName := ""
	for k, v := range wx.contacts {
		if v.RemarkName == remarkName {
			userName = k
			break
		}
	}
	if userName == "" {
		wx.logger.Error("未找到该联系人")
		return "", errors.New("未找到该联系人")
	}
	return userName, nil
}

func (wx *WxChat) GetRemarkName(userName string) string {
	return wx.contacts[userName].RemarkName
}
