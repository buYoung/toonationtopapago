package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/gorilla/websocket"
	"github.com/zserge/lorca"
	"go.uber.org/atomic"
	"io"
	"io/ioutil"
	"livteam/toonationpapago/global"
	"log"
	"os"
	"regexp"
	"strings"
	"time"
)

var (
	papago   lorca.UI
	userData global.UserData
)

func main() {
	if FileExists("user.json") {
		data, err := ioutil.ReadFile("user.json")
		if err != nil {
			global.Log_Error.Sugar.Errorf("UserInfo read Error: %s", err)
			return
		}
		err = json.Unmarshal(data, &userData)
		if err != nil {
			global.Log_Error.Sugar.Errorf("UserInfo read Error: %s", err)
			return
		}
	} else {
		fmt.Println("처음사용시 투네이션 위젯 URL을 입력해야합니다.")
		fmt.Println("모든 도네이션(후원, 구독, 팔로우, 호스팅, 비트, 구독선물, 구독선물갯수)를 얻을꺼면 기봇위젯 URL의 '톱합위젯' URL을 적어주세요")
		fmt.Println("후원 알림 번역기능만 쓸꺼면 '세부 위젯 URL'을 누르고 '후원 알림 위젯'의 URL을 아래의 URL : 에 적어주세요")
		fmt.Println("はじめてご使用の際にはトゥネーションウィジェットURLを入力してください。")
		fmt.Println("すべてのドネーション(後援、購読、フォロー、ホスティング、ビット、購読プレゼント、購読プレゼントの数)を得る場合は、基本ウィジェットURLの「トップ合ウィジェット」URLを記入してください。(コピー貼り付け)")
		fmt.Println("後援通知の翻訳機能を使う場合は、「詳細ウィジェットURL」をクリックして「後援通知ウィジェット」のURLを以下の「URL:」にご記入ください。(コピー貼り付け)")

		var strscan string
		fmt.Print("URL : ")
		fmt.Scan(&strscan)
		userData.WigetUrl = strscan

		data, err := json.Marshal(userData)
		if err != nil {
			return
		}
		ioutil.WriteFile("user.json", data, 0644)
	}

	s, err := alertwindow()
	if err != nil {
		log.Println("Error", err)
		return
	}
	alertStruct := parsepayload(s)
	if alertStruct == nil {
		return
	}
	papagoWindow()
	connectWebsocket(alertStruct)
}

func papagoWindow() {
	if !DirectoryExists("./cache") {
		os.Mkdir("cache", 0755)
	}

	papago, _ = lorca.New("https://papago.naver.com/", "cache", 1280, 960)
	if papago != nil {
		global.Log_Error.Sugar.Errorf("papago Window Create Error %#v", papago)
		return
	}

	go func() {
		<-papago.Done()
		os.Exit(0)
	}()
}
func alertwindow() (string, error) {

	url := userData.WigetUrl

	client := resty.New()
	req := client.R()
	resp, err := req.Get(url)
	if err != nil {
		global.Log_Error.Sugar.Errorf("TonaionPayloadParse : %v URL : %s", err, url)
		return "", nil
	}
	strresp := resp.String()

	resp.RawResponse.Body.Close()
	io.Copy(ioutil.Discard, resp.RawResponse.Body)

	return strresp, err
}
func parsepayload(str string) *global.ToonationAlertStruct {
	jsonStruct := &global.ToonationAlertStruct{}

	var re = regexp.MustCompile(`(?m)\w+.\w+\s=\s(?P<test>.*?);`)
	if re.MatchString(str) {
		err := json.Unmarshal([]byte(re.FindAllStringSubmatch(str, -1)[0][1]), jsonStruct)
		if err != nil {
			global.Log_Error.Sugar.Errorf("TonaionPayloadParse : %v data : %s", err, str)
			return nil
		}
	}
	return jsonStruct
}
func connectWebsocket(payload *global.ToonationAlertStruct) {
	isClose := atomic.NewBool(false)
	url := fmt.Sprintf("wss://toon.at:8071/%s", payload.Payload)
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Println("Websocket Connection Error", err)
		global.Log_Error.Sugar.Errorf("Websocket Connection Error : %v", err)
		return
	}
	defer c.Close()

	go func() {
		for {
			if isClose.Load() {
				break
			}
			time.Sleep(time.Second * 5)
			err := c.WriteMessage(websocket.PingMessage, []byte("#ping"))
			if err != nil {
				log.Println("Send Errror", err)
				global.Log_Error.Sugar.Errorf("TonaionWssWrite : %v", err)
				return
			}
		}
	}()
	for {
		var jsonstruct global.ToonationAlertWebsocket
		err := c.ReadJSON(&jsonstruct)
		if err != nil {
			log.Println("error", err)
			if strings.Contains(err.Error(), "abnormal closure") {
				isClose.Store(true)
				go connectWebsocket(payload)
				return
			}
			global.Log_Error.Sugar.Errorf("TonaionWssRead : %v", err)
		}
		global.Log_toonation_message.Sugar.Infof("%+v", jsonstruct)
		switch jsonstruct.Code {
		case 101:
			if jsonstruct.Content.VideoInfo == nil {
				log.Printf("Donation name : %s Message : %s", jsonstruct.Content.Name, jsonstruct.Content.Message)
				papago.Load(fmt.Sprintf("https://papago.naver.com/?sk=auto&tk=ja&st=%s", jsonstruct.Content.Message))
			} else {
				log.Println("Video Donation Pass")
			}
		case 102:
			log.Printf("Subscription name : %s %s", jsonstruct.Content.Name, jsonstruct.Content.Message)
		case 103:
			log.Printf("fallow name : %s", jsonstruct.Content.Name)
		case 104:
			log.Printf("Host name : %s", jsonstruct.Content.Name)
		case 107:
			log.Printf("Donation vite! : %s Count : %d", jsonstruct.Content.Name, jsonstruct.Content.Count)
		case 115:
			log.Printf("Subscription gift count : %d", jsonstruct.Content.Count)
		}
	}
}

func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func DirectoryExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}
