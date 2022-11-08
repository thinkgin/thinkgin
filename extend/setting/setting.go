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

	PageSize      int
	JwtSecret     string
	LOG_FILE_PATH string
	LOG_FILE_NAME string
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
	LOG_FILE_PATH = sec.Key("LOG_FILE_PATH").MustString(logpath)
	LOG_FILE_NAME = sec.Key("LOG_FILE_NAME").MustString(logname)
}

func LoadBase() {
	RunMode = Cfg.Section("").Key("RUN_MODE").MustString("debug")
}

func LoadServer() {
	sec, err := Cfg.GetSection("server")
	if err != nil {
		log.Fatalf("Fail to get section 'server': %v", err)
	}

	HTTPPort = sec.Key("HTTP_PORT").MustInt(8000)
	ReadTimeout = time.Duration(sec.Key("READ_TIMEOUT").MustInt(60)) * time.Second
	WriteTimeout = time.Duration(sec.Key("WRITE_TIMEOUT").MustInt(60)) * time.Second
}

func LoadApp() {
	sec, err := Cfg.GetSection("app")
	if err != nil {
		log.Fatalf("Fail to get section 'app': %v", err)
	}

	JwtSecret = sec.Key("JWT_SECRET").MustString("!@)*#)!@U#@*!@!)")
	PageSize = sec.Key("PAGE_SIZE").MustInt(10)
}
