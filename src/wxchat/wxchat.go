package wxchat

import (
	"fmt"
	"io"
	"log"
	"os"
)

type WxChat struct {
	Uuid        string
	baseRequest baseRequest
	passTicket  string
	syncKey     syncKey
	syncHost    string
	host        string
	me          Contact
	contacts    map[string]*Contact
	httpClient  *httpClient
	storage     *storage
	logger      *log.Logger
	listeners   map[EventType]func(Event)
}

// New A WxChat
func NewWxChat(storageFilePath string, logFile io.Writer) *WxChat {
	storage := storage{
		filePath: storageFilePath,
	}

	if logFile == nil {
		logFile = os.Stdout
	}
	logger := log.New(logFile, "", log.Ldate|log.Ltime)

	return &WxChat{
		httpClient: &httpClient{},
		storage:    &storage,
		listeners:  map[EventType]func(Event){},
		logger:     logger,
	}
}

// Login And Init
func (wx *WxChat) Login() error {

	err := wx.beginLogin()
	if err != nil {
		return err
	}

	err = wx.init()
	if err != nil {
		wx.storage.delData()
		err = wx.beginLogin()
		if err != nil {
			return err
		}
		err = wx.init()
		if err != nil {
			return err
		}
	}

	wx.triggerInitEvent(wx.me)
	wx.logger.Println("[Info] WxChat Init.")

	err = wx.initContact()
	if err != nil {
		return err
	}

	wx.triggerContactsInitEvent(len(wx.contacts))
	wx.logger.Println("[Info] Contacts Init.")

	return nil
}

func (wx *WxChat) Run() error {
	err := wx.beginListen()
	return err
}

func (wx *WxChat) skeyKV() string {
	return fmt.Sprintf(`skey=%s`, wx.baseRequest.Skey)
}
