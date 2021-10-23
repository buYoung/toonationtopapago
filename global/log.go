package global

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type logs struct {
	logger       *zap.Logger
	Sugar        *zap.SugaredLogger
	config       zap.Config
	path         string
	filename     string
	pullfilename string
}

var (
	Log_toonation_message *logs
	Log_Error             *logs
)

func init() {
	os.Mkdir("log", 0755)

	Log_toonation_message = &logs{
		path:     "./log/tonation/",
		filename: "message",
	}
	Log_Error = &logs{
		path:     "./log/error/",
		filename: "error",
	}

	Log_toonation_message.init()
	Log_Error.init()

	go func() {
		for {
			Findoldfile("./log")
			time.Sleep(time.Hour)
		}
	}()
}

func Dateformats() string {
	t := time.Now()
	datez := fmt.Sprintf("%d-%02d-%02d.log", t.Year(), t.Month(), t.Day())
	return datez
}
func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
func Join(strs ...string) string {
	var sb = &strings.Builder{}
	defer func() {
		sb = nil
	}()

	for _, str := range strs {
		sb.WriteString(str)
	}
	return sb.String()
}

func NewLogger(filename string) (*zap.Logger, error) {
	if !FileExists(fmt.Sprintf("./log/%s", filename)) {
		os.Create(fmt.Sprintf("./log/%s", filename))
	}
	cfg := zap.NewProductionConfig()
	encoderconfigz := zapcore.EncoderConfig{
		TimeKey:        "date",
		LevelKey:       "level",
		NameKey:        "logger",
		MessageKey:     "msg",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
	}
	cfg.EncoderConfig = encoderconfigz
	cfg.OutputPaths = []string{
		fmt.Sprintf("./log/%s", filename),
	}
	return cfg.Build()
}

func Findoldfile(filepathname string) {
	df := time.Now()
	filepath.Walk(filepathname, func(pathi string, infoi os.FileInfo, err error) error {
		reg := regexp.MustCompile(`-(\w+-\w+-\w+).log`)
		s := reg.FindStringSubmatch(pathi)
		if len(s) > 0 {
			t, _ := time.ParseInLocation("2006-01-02", s[1], df.Location())
			if int(df.Sub(t).Hours()) > 168 {
				os.Remove(pathi)
			}
		}
		return err
	})
}

func (l *logs) init() {
	os.MkdirAll(l.path, 0755)
	os.Chmod(l.path, 0755)
	l.config = zap.NewProductionConfig()
	encoderconfigz := zapcore.EncoderConfig{
		TimeKey:        "date",
		LevelKey:       "level",
		NameKey:        "logger",
		MessageKey:     "msg",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
	}
	l.config.EncoderConfig = encoderconfigz

	l.create()

	go func() {
		for {
			time.Sleep(1 * time.Minute)
			camse := Join(l.filename, "_", Dateformats())
			if l.pullfilename != camse {
				log.Println("로그_상태_새로운파일생성")
				l.create()
			}
		}
	}()
	go func() {
		for {
			Findoldfile(l.path)
			time.Sleep(time.Hour)
		}
	}()
}
func (l *logs) create() {
	filename := Join(l.path, l.filename, "_", Dateformats())
	l.pullfilename = Join(l.filename, "_", Dateformats())
	l.config.OutputPaths = []string{
		filename,
	}
	logger, err := l.config.Build()
	if err != nil {
		log.Println("로그_상태_생성실패", err)
	}
	l.logger = logger
	l.Sugar = l.logger.Sugar()
	err = l.logger.Sync()
	if err != nil {
		log.Println("로그_상태_오류 : ", err)
	}
}
