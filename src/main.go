package main

import (
	"fmt"
	"wxchat"
	logs "wxchat/log"
)

var (
	cmdFlag = false
	addFlag = false
	delFlag = false
	//nameList = map[string]bool{}
)

type AutoSend struct {
	auto   bool
	tuling bool
	name   string
}

var list = map[string]*AutoSend{}

func main() {
	logger := logs.NewLogger()
	wx := wxchat.NewWxChat("./db.json", logger)
	GetContactsList(wx)
	MessageListener(wx)
	err := wx.Login()
	if err != nil {
		logger.Error(err.Error())
	}
	err = wx.Run()
	if err != nil {
		logger.Error(err.Error())
	}

}

func GetContactsList(wx *wxchat.WxChat) {
	wx.SetListener(wxchat.CONTACTS_INIT_EVENT, func(event wxchat.Event) {
		lists := wx.AllContact()
		for k, v := range lists {
			if v.RemarkName == "" {
				list[k] = &AutoSend{
					name:   v.NickName,
					auto:   false,
					tuling: false,
				}
			} else {
				list[k] = &AutoSend{
					name:   v.RemarkName,
					auto:   false,
					tuling: false,
				}
			}
		}
	})
}

func MessageListener(wx *wxchat.WxChat) {
	wx.SetListener(wxchat.MESSAGE_EVENT, func(event wxchat.Event) {
		eventData, ok := event.Data.(wxchat.MessageEventData)

		if ok {
			if eventData.IsSendByMySelf && wxchat.TextMessage == eventData.MessageType {
				if "cmd" == eventData.Content {
					cmdFlag = true
					wx.SendTextMsg("1. 添加自动应答好友\n2. 删除自动应答好友\n3. 已添加的好友", "filehelper")
				} else if "over" == eventData.Content {
					cmdFlag = false
					wx.SendTextMsg("操作结束", "filehelper")
				} else if cmdFlag {
					_ = cmd(wx, eventData.Content)
				}

			} else if list[eventData.SenderUserInfo.UserName].auto && wxchat.TextMessage == eventData.MessageType {
				_, _ = wx.SendTextMsg("[自动回复]对方暂时不想理你，等会再说(^_^)", eventData.SenderUserInfo.UserName)
			}
		}
	})
}

func cmd(wx *wxchat.WxChat, msg string) error {
	var err error = nil
	if "1" == msg {
		addFlag = true
		delFlag = false
		_, err = wx.SendTextMsg("请输入添加自动应答好友的备注名", "filehelper")
	} else if "2" == msg {
		delFlag = true
		addFlag = false
		_, err = wx.SendTextMsg("请输入删除自动应答好友的备注名", "filehelper")
	} else if "3" == msg {
		var names = "当前添加的好友有"
		if len(list) > 0 {
			for _, v := range list {
				if v.auto {
					names = fmt.Sprintf("%s\n%s", names, v.name)
				}
			}
		}

		_, err = wx.SendTextMsg(names, "filehelper")
	} else if addFlag {
		userName, err := wx.SearchContact(msg)
		if err != nil {
			_, err = wx.SendTextMsg(err.Error(), "filehelper")
		} else {
			list[userName].auto = true
		}

	} else if delFlag {
		userName, err := wx.SearchContact(msg)
		if err != nil {
			_, err = wx.SendTextMsg(err.Error(), "filehelper")
		} else {
			list[userName].auto = false
		}
	} else {
		_, err = wx.SendTextMsg(fmt.Sprintf("未知命令[%s]", msg), "filehelper")
	}

	return err
}

//func MessageListener(wx *wxchat.WxChat)  {
//	wx.SetListener(wxchat.MESSAGE_EVENT, func(event wxchat.Event){
//		eventData, ok := event.Data.(wxchat.MessageEventData)
//
//		if ok {
//			//if eventData.IsGroupMessage {
//			//	if eventData.IsAtMe {
//			//		res, err := tuling(eventData.Content, "青岛",eventData.SenderUserId)
//			//		if err != nil {
//			//			wx.SendTextMsg("@"+ eventData.SenderUserInfo.NickName +" "+"短路了...快通知我主人修修我...", eventData.FromUserName)
//			//		} else {
//			//			wx.SendTextMsg("@"+ eventData.SenderUserInfo.NickName +" "+res, eventData.FromUserName)
//			//		}
//			//	}
//			//} else {
//
//				if "AI小号" == eventData.SenderUserInfo.RemarkName && wxchat.TextMessage == eventData.MessageType{
//					m := map[string]string{}
//					b:= m[eventData.SenderUserInfo.RemarkName]
//					//b, _:= json.Marshal(eventData.SenderUserInfo)
//					fmt.Println("m|"+b+"|")
//					wx.SendTextMsg("test 成功", eventData.SenderUserInfo.UserName)
//				//}else if wxchat.FriendReqMessage != eventData.MessageType {
//				//	fmt.Println(eventData.SenderUserInfo.UserName)
//				//	wx.SendTextMsg("[微笑]这是自动回复", eventData.SenderUserInfo.UserName)
//				}
//			//}
//		}
//	})
//}
