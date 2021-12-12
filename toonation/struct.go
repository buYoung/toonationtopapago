package toonation

import "sync"

type ToonationKey struct {
	WigetUrl string `json:"WigetUrl"`
}
type ToonationAlertStruct struct {
	FilterCode int         `json:"filterCode"`
	RemoteConf interface{} `json:"-"`
	UID        string      `json:"uid"`
	Payload    string      `json:"payload"`
	Cm         bool        `json:"cm"`
	LocaleCode string      `json:"LocaleCode"`
}

type ToonationInstance struct {
	Data []ToonationType
	L    sync.RWMutex
}
type ToonationType struct {
	Test    int                  `json:"test"`
	Code    int                  `json:"code"`
	Content ToonationTypeContent `json:"content"`
}

type ToonationTypeContent struct {
	Amount      int         `json:"amount"`
	UID         string      `json:"uid"`
	Account     string      `json:"account"`
	Name        string      `json:"name"`
	Image       string      `json:"image"`
	Acctype     int         `json:"acctype"`
	Count       int         `json:"count,omitempty"`
	TestNoti    int         `json:"test_noti"`
	Level       int         `json:"level"`
	TtsLocale   string      `json:"tts_locale"`
	TtsProvider string      `json:"tts_provider"`
	Message     string      `json:"message"`
	ConfIdx     int         `json:"conf_idx"`
	RecLink     string      `json:"rec_link"`
	VideoInfo   interface{} `json:"video_info"`
	VideoBegin  int         `json:"video_begin"`
	VideoLength int         `json:"video_length"`
	TtsLink     string      `json:"tts_link"`
}
