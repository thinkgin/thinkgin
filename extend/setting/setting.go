package setting

import (
	log "github.com/sirupsen/logrus"
	"time"

	"github.com/go-ini/ini"
)

var (
	Cfg *ini.File

	RunMode string

	HTTPPort     int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration

	PageSize    int
	JwtSecret   string
	LogFilePath string
	LogFileName string
)

func init() {
	var err error
	Cfg, err = ini.Load("config.ini")
	if err != nil {
		log.Fatalf("Fail to parse 'config/config.ini': %v", err)
	}
	LoadLog()
	LoadBase()
	LoadServer()
	LoadApp()
}

func LoadLog() {
	sec, err := Cfg.GetSection("log")
	if err != nil {
		log.Fatalf("Fail to get section 'app': %v", err)
	}
	logpath := "runtime/log"
	logname := "system"
	LogFilePath = sec.Key("LogFilePath").MustString(logpath)
	LogFileName = sec.Key("LogFileName").MustString(logname)
}

func LoadBase() {
	RunMode = Cfg.Section("").Key("RunMode").MustString("debug")
}

func LoadServer() {
	sec, err := Cfg.GetSection("server")
	if err != nil {
		log.Fatalf("Fail to get section 'server': %v", err)
	}

	HTTPPort = sec.Key("HttpPort").MustInt(8000)
	ReadTimeout = time.Duration(sec.Key("ReadTimeout").MustInt(60)) * time.Second
	WriteTimeout = time.Duration(sec.Key("WriteTimeout").MustInt(60)) * time.Second
}

func LoadApp() {
	sec, err := Cfg.GetSection("app")
	if err != nil {
		log.Fatalf("Fail to get section 'app': %v", err)
	}

	JwtSecret = sec.Key("JwtSecret").MustString("!@)*#)!@U#@*!@!)")
	PageSize = sec.Key("PageSize").MustInt(10)
}
