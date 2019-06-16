package wxchat

import (
	"fmt"
	"strings"
	"time"
	"wxchat/utils"
)

// 事件类型
type EventType int

const (
	_                    EventType = iota
	GEN_UUID_EVENT                 // 生成Uuid
	SCAN_CODE_EVENT                // 已扫码，未确认
	CONFIRM_AUTH_EVENT             // 已确认授权登录
	LOGIN_EVENT                    // 已登录
	INIT_EVENT                     // 初始化完成
	CONTACTS_INIT_EVENT            // 联系人初始化完
	LISTEN_FAILED_EVENT            // 同步微信失败,可能为客户端已退出 | 被微信反爬虫
	CONTACT_MODIFY_EVENT           // 联系人改变了
	CONTACT_DELETE_EVENT           // 联系人删除事件
	MESSAGE_EVENT                  // 消息
)

// 事件体
type Event struct {
	Time      int64
	EventType EventType
	Data      interface{}
}

// 生成Uuid事件的数据
type GenUuidEventData struct {
	Uuid string
}

// 扫码事件数据
type ScanCodeEventData struct {
	UserAvatar string
}

// 授权事件数据
type ConfirmAuthEventData struct {
	RedirectUrl string
}

// 登录事件数据
type LoginEventData struct {
	DeviceID string
}

// 初始化事件数据
type InitEventData struct {
	Me Contact
}

// 通讯录初始化事件数据
type ContactsInitEventData struct {
	ContactsCount int
}

// 同步微信失败事件数据
type ListenFailedEventData struct {
	ListenFailedCount int
	Host              string
}

// 联系人修改事件数据
type ContactModifyEventData struct {
	UserNames []string
}

// 联系人删除事件数据
type ContactDeleteEventData struct {
	UserNames []string
}

// 消息事件数据
type MessageEventData struct {
	MessageType    MessageType
	IsGroupMessage bool
	IsSendByMySelf bool
	IsAtMe         bool
	MediaUrl       string
	Content        string
	FromUserName   string
	FromUserInfo   Contact
	SenderUserInfo SenderUserInfo
	SenderUserId   string // 根据SendUserName生成ID
	ToUserName     string
	ToUserInfo     Contact
	RecommendInfo  map[string]interface{}
	LocationInfo   LocationInfo
	OriginalMsg    map[string]interface{}
}

// 消息类型
type MessageType int

const (
	_ MessageType = iota
	TextMessage
	ImgMessage
	VoiceMessage
	VideoMessage
	CardMessage
	LocationMessage
	FriendReqMessage
)

// 发送人信息
type SenderUserInfo struct {
	UserName   string
	NickName   string
	RemarkName string
}

// 位置信息数据
type LocationInfo struct {
	X     string
	Y     string
	Label string
	Img   string
}

// 设置事件监听器
func (wx *WxChat) SetListener(eventType EventType, listener func(Event)) {
	wx.listeners[eventType] = listener
}

// 处理从微信服务器拉过来的响应数据
func (wx *WxChat) handleSyncResponse(resp *syncMessageResponse) {

	if resp.ModContactCount > 0 {
		userNames := []string{}
		for _, v := range resp.ModContactList {
			userNames = append(userNames, v["UserName"].(string))
		}
		go wx.triggerContactModifyEvent(userNames)
	}

	if resp.DelContactCount > 0 {
		userNames := []string{}
		for _, v := range resp.ModContactList {
			userNames = append(userNames, v["UserName"].(string))
		}
		go wx.triggerContactDeleteEvent(userNames)
	}

	if resp.AddMsgCount > 0 {
		for _, v := range resp.AddMsgList {
			go wx.triggerMessageEvent(v)
		}
	}
}

// 触发生成uuid的事件
func (wx *WxChat) triggerGenUuidEvent(uuid string) {
	listener, isReg := wx.listeners[GEN_UUID_EVENT]
	if isReg {
		listener(Event{
			Time:      time.Now().Unix(),
			EventType: GEN_UUID_EVENT,
			Data: GenUuidEventData{
				Uuid: uuid,
			},
		})
	}
}

// 触发扫码事件(未确认)
func (wx *WxChat) triggerScanCodeEvent(userAvatar string) {
	listener, isReg := wx.listeners[SCAN_CODE_EVENT]
	if isReg {
		listener(Event{
			Time:      time.Now().Unix(),
			EventType: SCAN_CODE_EVENT,
			Data: ScanCodeEventData{
				UserAvatar: userAvatar,
			},
		})
	}
}

// 触发授权登录事件
func (wx *WxChat) triggerConfirmAuthEvent(redirectUrl string) {
	listener, isReg := wx.listeners[CONFIRM_AUTH_EVENT]
	if isReg {
		listener(Event{
			Time:      time.Now().Unix(),
			EventType: CONFIRM_AUTH_EVENT,
			Data: ConfirmAuthEventData{
				RedirectUrl: redirectUrl,
			},
		})
	}
}

// 触发登录事件
func (wx *WxChat) triggerLoginEvent(deviceID string) {
	listener, isReg := wx.listeners[LOGIN_EVENT]
	if isReg {
		listener(Event{
			Time:      time.Now().Unix(),
			EventType: LOGIN_EVENT,
			Data: LoginEventData{
				DeviceID: deviceID,
			},
		})
	}
}

// 触发初始化事件
func (wx *WxChat) triggerInitEvent(me Contact) {
	listener, isReg := wx.listeners[INIT_EVENT]
	if isReg {
		listener(Event{
			Time:      time.Now().Unix(),
			EventType: INIT_EVENT,
			Data: InitEventData{
				Me: me,
			},
		})
	}
}

// 触发通讯录初始化事件
func (wx *WxChat) triggerContactsInitEvent(contactsCount int) {
	listener, isReg := wx.listeners[CONTACTS_INIT_EVENT]
	if isReg {
		listener(Event{
			Time:      time.Now().Unix(),
			EventType: CONTACTS_INIT_EVENT,
			Data: ContactsInitEventData{
				ContactsCount: contactsCount,
			},
		})
	}
}

func (wx *WxChat) triggerListenFailedEvent(listenFailedCount int, host string) {
	listener, isReg := wx.listeners[LISTEN_FAILED_EVENT]
	if isReg {
		listener(Event{
			Time:      time.Now().Unix(),
			EventType: LISTEN_FAILED_EVENT,
			Data: ListenFailedEventData{
				ListenFailedCount: listenFailedCount,
				Host:              host,
			},
		})
	}
}

// 触发通讯录修改事件
func (wx *WxChat) triggerContactModifyEvent(userNames []string) {
	listener, isReg := wx.listeners[CONTACT_MODIFY_EVENT]
	if isReg {
		listener(Event{
			Time:      time.Now().Unix(),
			EventType: CONTACT_MODIFY_EVENT,
			Data: ContactModifyEventData{
				UserNames: userNames,
			},
		})
	}
}

// 触发通讯录删除事件
func (wx *WxChat) triggerContactDeleteEvent(userNames []string) {
	listener, isReg := wx.listeners[CONTACT_DELETE_EVENT]
	if isReg {
		listener(Event{
			Time:      time.Now().Unix(),
			EventType: CONTACT_DELETE_EVENT,
			Data: ContactDeleteEventData{
				UserNames: userNames,
			},
		})
	}
}

// 触发消息事件
func (wx *WxChat) triggerMessageEvent(msg map[string]interface{}) {

	messageType := TextMessage
	isGroupMessage := false
	isSendByMySelf := false
	isAtMe := false
	mediaUrl := ""
	content := msg["Content"].(string)
	fromUserName := msg["FromUserName"].(string)
	senderUserInfo := SenderUserInfo{}
	senderUserId := ""
	toUserName := msg["ToUserName"].(string)
	recommendInfo := map[string]interface{}{}
	locationInfo := LocationInfo{}
	senderUserName := fromUserName

	var groupUserName string
	if strings.HasPrefix(fromUserName, "@@") {
		groupUserName = fromUserName
	} else if strings.HasPrefix(toUserName, "@@") {
		groupUserName = toUserName
	}

	if len(groupUserName) > 0 {
		isGroupMessage = true
	}

	msgType := msg["MsgType"].(float64)
	mid := msg["MsgId"].(string)

	path := ""
	switch msgType {
	case 3:
		{
			messageType = ImgMessage
			path = "webwxgetmsgimg"
		}
	case 47:
		{
			pid, _ := msg["HasProductId"].(float64)
			if pid == 0 {
				messageType = ImgMessage
				path = "webwxgetmsgimg"
			}
		}
	case 34:
		{
			messageType = VoiceMessage
			path = "webwxgetvoice"
		}
	case 43:
		{
			messageType = VideoMessage
			path = "webwxgetvideo"
		}
	case 37:
		{
			messageType = FriendReqMessage
			recommendInfo, _ = msg["RecommendInfo"].(map[string]interface{})
		}
	case 42:
		{
			messageType = CardMessage
		}
	}
	if len(path) > 0 {
		mediaUrl = fmt.Sprintf(`https://wx2.qq.com/%s?msgid=%v&%v`, path, mid, wx.skeyKV())
	}

	subMsgType, found := msg["SubMsgType"]
	if found && 48 == subMsgType.(float64) {
		messageType = LocationMessage
		locationX, locationY, locationLabel, err := utils.GetLocationInfoFromOriContent(msg["OriContent"].(string))
		if err == nil {
			locationInfo.X = locationX
			locationInfo.Y = locationY
			locationInfo.Label = locationLabel
		}

		locationImg, err := utils.GetLocationImgFromContent(content)
		if err == nil {
			locationInfo.Img = "https://" + wx.host + locationImg
		}
	}

	if isGroupMessage {
		atMe := "@"
		if len(wx.me.DisplayName) > 0 {
			atMe += wx.me.DisplayName
		} else {
			atMe += wx.me.NickName
		}
		isAtMe = strings.Contains(content, atMe) // 标识是否是@我

		infos := strings.Split(content, ":<br/>")
		if len(infos) != 2 {
			return
		}

		content = infos[1]
		if isAtMe && infos[0] == wx.me.UserName {
			isSendByMySelf = true
		}

		contact := &Member{}
		for {
			fromGroup, found := wx.contacts[fromUserName]
			if found {
				contact, found = fromGroup.MemberMap[infos[0]] // 根据content中UserName(消息发布人)找到详细数据
				if found {
					break
				}
			}

			err := wx.updateContact([]string{fromUserName})
			if err != nil {
				return
			}

			contact, found = wx.contacts[fromUserName].MemberMap[infos[0]]
			if !found {
				return
			}
		}

		if nil == contact {
			return
		}
		senderUserName = infos[0]
		senderUserInfo = SenderUserInfo{
			UserName:   infos[0],
			NickName:   contact.NickName,
			RemarkName: contact.DisplayName,
		}
	} else {

		isSendByMySelf = fromUserName == wx.me.UserName
		if isSendByMySelf {
			senderUserInfo = SenderUserInfo{
				UserName:   senderUserName,
				NickName:   wx.me.NickName,
				RemarkName: "mySelf",
			}
		} else {
			senderUserInfo = SenderUserInfo{
				UserName:   senderUserName,
				NickName:   "",
				RemarkName: "",
			}

			senderUser, found := wx.contacts[senderUserName]

			if found {
				senderUserInfo.NickName = senderUser.NickName
				senderUserInfo.RemarkName = senderUser.RemarkName
			}
		}
	}

	fromUserInfo := wx.me
	if !isSendByMySelf {
		fromUserInfoTemp, found := wx.contacts[fromUserName]
		if found {
			fromUserInfo = *fromUserInfoTemp
		}
	}

	toUserInfo := wx.me
	if toUserName != wx.me.UserName {
		toUserInfoTemp, found := wx.contacts[toUserName]
		if found {
			toUserInfo = *toUserInfoTemp
		}
	}

	senderUserId = utils.UserNameToId(senderUserName)

	event := Event{
		Time:      time.Now().Unix(),
		EventType: MESSAGE_EVENT,
		Data: MessageEventData{
			MessageType:    messageType,
			IsGroupMessage: isGroupMessage,
			IsSendByMySelf: isSendByMySelf,
			IsAtMe:         isAtMe,
			MediaUrl:       mediaUrl,
			Content:        content,
			FromUserName:   fromUserName,
			FromUserInfo:   fromUserInfo,
			SenderUserInfo: senderUserInfo,
			SenderUserId:   senderUserId,
			ToUserName:     toUserName,
			ToUserInfo:     toUserInfo,
			RecommendInfo:  recommendInfo,
			LocationInfo:   locationInfo,
			OriginalMsg:    msg,
		},
	}

	wx.logger.Println("[Info] Get Message. SenderNickName=[" + senderUserInfo.NickName + "], Content=[" + content + "]")
	listener, isReg := wx.listeners[MESSAGE_EVENT]
	if isReg {
		listener(event)
	}
}
