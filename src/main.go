package main

import (
	"fmt"
	"wxchat"
	logs "wxchat/log"
)

var (
	cmdFlag  = false
	addFlag  = false
	delFlag  = false
	nameList = make(map[string]bool)
)

func main() {
	logger := logs.NewLogger()
	wx := wxchat.NewWxChat("./db.json", logger)
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

func MessageListener(wx *wxchat.WxChat) {
	wx.SetListener(wxchat.MESSAGE_EVENT, func(event wxchat.Event) {
		eventData, ok := event.Data.(wxchat.MessageEventData)

		if ok {
			if eventData.IsSendByMySelf && wxchat.TextMessage == eventData.MessageType {
				if "cmd" == eventData.Content {
					cmdFlag = true
				}
				if cmdFlag {
					_ = cmd(wx, eventData.Content)
				}
				if "over" == eventData.Content {
					cmdFlag = false
				}
			}
			if nameList[eventData.SenderUserInfo.RemarkName] && wxchat.TextMessage == eventData.MessageType {
				_, _ = wx.SendTextMsg("[自动回复]对方暂时不想理你，等会再说(^_^)", eventData.SenderUserInfo.UserName)

			}
		}
	})
}

func cmd(wx *wxchat.WxChat, msg string) error {
	var err error = nil
	if "cmd" == msg {
		_, err = wx.SendTextMsg("1. 添加自动应答好友\n2. 删除自动应答好友\n", "filehelper")
	} else if "over" == msg {
		_, err = wx.SendTextMsg("操作结束", "filehelper")
	} else if "1" == msg {
		addFlag = true
		delFlag = false
		_, err = wx.SendTextMsg("请输入添加自动应答好友的备注名", "filehelper")
	} else if "2" == msg {
		delFlag = true
		addFlag = false
		_, err = wx.SendTextMsg("请输入删除自动应答好友的备注名", "filehelper")
	} else if addFlag {
		nameList[msg] = true
	} else if delFlag {
		delete(nameList, msg)
	}
	if addFlag || delFlag {
		var names = "当前自动应答好友\n"
		for name := range nameList {
			names = fmt.Sprintf("%s|%s", names, name)
		}
		_, err = wx.SendTextMsg(names, "filehelper")
	}

	return err
}

/*
func MessageListener(wx *wxchat.WxChat)  {
	wx.SetListener(wxchat.MESSAGE_EVENT, func(event wxchat.Event){
		eventData, ok := event.Data.(wxchat.MessageEventData)

		if ok {
			//if eventData.IsGroupMessage {
			//	if eventData.IsAtMe {
			//		res, err := tuling(eventData.Content, "青岛",eventData.SenderUserId)
			//		if err != nil {
			//			wx.SendTextMsg("@"+ eventData.SenderUserInfo.NickName +" "+"短路了...快通知我主人修修我...", eventData.FromUserName)
			//		} else {
			//			wx.SendTextMsg("@"+ eventData.SenderUserInfo.NickName +" "+res, eventData.FromUserName)
			//		}
			//	}
			//} else {

				if "AI小号" == eventData.SenderUserInfo.RemarkName && wxchat.TextMessage == eventData.MessageType{
					m := map[string]string{}
					b:= m[eventData.SenderUserInfo.RemarkName]
					//b, _:= json.Marshal(eventData.SenderUserInfo)
					fmt.Println("m|"+b+"|")
					wx.SendTextMsg("test 成功", eventData.SenderUserInfo.UserName)
				//}else if wxchat.FriendReqMessage != eventData.MessageType {
				//	fmt.Println(eventData.SenderUserInfo.UserName)
				//	wx.SendTextMsg("[微笑]这是自动回复", eventData.SenderUserInfo.UserName)
				}
			//}
		}
	})
}*/
