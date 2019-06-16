package wxchat

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
)

type storage struct {
	filePath string
}

type storageData struct {
	Uuid        string
	BaseRequest baseRequest
	PassTicket  string
	Cookies     []*http.Cookie
	Host        string
}

func (storage *storage) setData(Uuid string, baseRequest baseRequest, passTicket string, cookies []*http.Cookie, host string) error {
	storageStr, _ := json.Marshal(storageData{
		Uuid:        Uuid,
		BaseRequest: baseRequest,
		PassTicket:  passTicket,
		Cookies:     cookies,
		Host:        host,
	})

	err := ioutil.WriteFile(storage.filePath, storageStr, 0700)

	return err
}

func (storage *storage) getData() (string, baseRequest, string, []*http.Cookie, string, error) {

	bs, err := ioutil.ReadFile(storage.filePath)
	if err != nil {
		return "", baseRequest{}, "", nil, "", err
	}

	var storageData storageData
	err = json.Unmarshal(bs, &storageData)
	if err != nil {
		return "", baseRequest{}, "", nil, "", err
	}

	if "" == storageData.Uuid || "" == storageData.PassTicket || "" == storageData.Host {
		return storageData.Uuid, storageData.BaseRequest, storageData.PassTicket, storageData.Cookies, storageData.Host, errors.New("Storage Is Nil")
	}

	return storageData.Uuid, storageData.BaseRequest, storageData.PassTicket, storageData.Cookies, storageData.Host, nil
}

func (this *storage) delData() error {
	fileName := this.filePath
	err := os.Remove(fileName)
	return err
}
