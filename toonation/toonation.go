package toonation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/gorilla/websocket"
	"go.uber.org/atomic"
	"io"
	"io/ioutil"
	"livteam/toonationpapago/util"
	"log"
	"os"
	"regexp"
	"strings"
	"time"
)

var (
	UserKey *ToonationKey
	Toona   = &ToonationInstance{}
	IsReady = atomic.NewBool(false)
)

func (i *ToonationInstance) Initialize() {
	go func() {
		if util.FileExists("user.json") {
			data, err := ioutil.ReadFile("user.json")
			if err != nil {
				util.Log_Error.Sugar.Errorf("UserInfo read Error: %s", err)
				return
			}
			err = json.Unmarshal(data, &UserKey)
			if err != nil {
				util.Log_Error.Sugar.Errorf("UserInfo read Error: %s", err)
				return
			}
		} else {
			IsReady.Store(true)
		}

		for {
			if UserKey != nil {
				break
			}
			time.Sleep(time.Second)
		}
		IsReady.Store(true)
		alertwindow, err := i.alertwindow()
		if err != nil {
			log.Println("Error", err)
			os.Exit(0)
		}
		parsepayload := i.parsepayload(alertwindow)

		i.toonationSync()
		i.connectWebsocket(parsepayload)
	}()
}
func (i *ToonationInstance) connectWebsocket(payload *ToonationAlertStruct) {
	isClose := atomic.NewBool(false)
	url := fmt.Sprintf("wss://toon.at:8071/%s", payload.Payload)
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Println("Websocket Connection Error", err)
		util.Log_Error.Sugar.Errorf("Websocket Connection Error : %v", err)
		return
	}
	defer c.Close()
	log.Println("Connect Success")
	go func() {
		for {
			if isClose.Load() {
				break
			}
			time.Sleep(time.Second * 5)
			err := c.WriteMessage(websocket.PingMessage, []byte("#ping"))
			if err != nil {
				log.Println("Send Errror", err)
				util.Log_Error.Sugar.Errorf("TonaionWssWrite : %v", err)
				return
			}
		}
	}()
	for {
		var jsonstruct ToonationType
		err := c.ReadJSON(&jsonstruct)
		if err != nil {
			log.Println("error", err)
			if strings.Contains(err.Error(), "abnormal closure") {
				isClose.Store(true)
				go i.connectWebsocket(payload)
				return
			}
			util.Log_Error.Sugar.Errorf("TonaionWssRead : %v", err)
		}
		i.L.Lock()
		i.Data = append(i.Data, jsonstruct)
		i.L.Unlock()
	}
}
func (i *ToonationInstance) alertwindow() (string, error) {

	url := UserKey.WigetUrl

	client := resty.New()
	req := client.R()
	resp, err := req.Get(url)
	if err != nil {
		util.Log_Error.Sugar.Errorf("TonaionPayloadParse : %v URL : %s", err, url)
		return "", nil
	}
	strresp := resp.String()

	resp.RawResponse.Body.Close()
	io.Copy(ioutil.Discard, resp.RawResponse.Body)
	return strresp, err
}

func (i *ToonationInstance) parsepayload(str string) *ToonationAlertStruct {
	jsonStruct := &ToonationAlertStruct{}

	var re = regexp.MustCompile(`(?m)\w+.\w+\s=\s(?P<test>.*?);`)
	if re.MatchString(str) {
		err := json.Unmarshal([]byte(re.FindAllStringSubmatch(str, -1)[0][1]), jsonStruct)
		if err != nil {
			log.Printf("TonaionPayloadParse : %v data : %s", err, str)
			util.Log_Error.Sugar.Errorf("TonaionPayloadParse : %v data : %s", err, str)
			return nil
		}
	}
	return jsonStruct
}
func (i *ToonationInstance) toonationSync() {
	go func() {
		client := resty.New()
		req := client.R()
		for {
			if len(i.Data) <= 0 {
				time.Sleep(time.Second)
				continue
			}
			i.L.RLock()
			data := i.Data
			i.L.RUnlock()

			switch data[0].Code {
			case 101:
				if data[0].Content.VideoInfo == nil {
					log.Printf("Donation name : %s Message : %s", data[0].Content.Name, data[0].Content.Message)
					Papago.Load(fmt.Sprintf("https://papago.naver.com/?sk=auto&tk=ja&st=%s", data[0].Content.Message))
				} else {
					log.Println("Video Donation Pass")
				}
			case 102:
				log.Printf("Subscription name : %s %s", data[0].Content.Name, data[0].Content.Message)
			case 103:
				log.Printf("fallow name : %s", data[0].Content.Name)
			case 104:
				log.Printf("Host name : %s", data[0].Content.Name)
			case 107:
				log.Printf("Donation vite! : %s Count : %d", data[0].Content.Name, data[0].Content.Count)
			case 115:
				log.Printf("Subscription gift count : %d", data[0].Content.Count)
			}
			if data[0].Content.TtsLink != "" {
				resp, err := req.Get(data[0].Content.TtsLink)
				if err != nil {
					util.Log_Error.Sugar.Errorf("ToonationNextEvent : %v URL : %s", err, data[0].Content.TtsLink)
				}

				duration, err := GetMP4Duration(bytes.NewReader(resp.Body()))
				if err != nil {
					log.Println("Get MP4 play Duration", err)
					time.Sleep(time.Second)
					continue
				}
				time.Sleep(time.Second * (time.Duration(duration) + 1))
			} else {
				time.Sleep(time.Second)
			}
			i.L.Lock()
			i.Data = i.Data[1:]
			i.L.Unlock()
		}
	}()
}
