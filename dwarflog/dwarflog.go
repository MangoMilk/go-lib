package dwarflog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"
)

const logDateFormat = "2006-01-02"

type LogFormat string
type LogLevel string
type LogLevelDate string

const (
	Dot                  = "."
	Slash                = "/"
	FileSuffix           = ".log"
	RollingInAdvance     = time.Second * 5 // must greater than 0
	RollingCheckInterval = time.Second * 3 // must greater than 0
	CloseFileDelay       = time.Second * 2 // must greater than 0

	PaleFormat = LogFormat("pale")
	JsonFormat = LogFormat("json")

	PanicLevel = LogLevel("panic")
	FatalLevel = LogLevel("fatal")
	ErrorLevel = LogLevel("error")
	InfoLevel  = LogLevel("info")
)

var logLevels = []LogLevel{ErrorLevel, InfoLevel}

type lg struct {
	lgr        *log.Logger
	fm         LogFormat
	rolling    bool
	path       string
	files      map[LogLevelDate]*os.File
	filePrefix string
	fileDate   string
}

var l *lg

func NewDwarfLog() *lg {
	return &lg{
		lgr: log.New(os.Stderr, "", log.LstdFlags),
	}
}

func (l *lg) Open(path string) {

	path = strings.TrimRight(path, Slash)
	files := make(map[LogLevelDate]*os.File)

	for _, level := range logLevels {
		filePath := path + Slash + l.filePrefix + l.fileDate + string(level) + FileSuffix
		file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			panic("【dwarflog error】open file [" + filePath + "] fall!")
		}
		// os.Chmod(file, 0644)
		files[l.getLevelDate(level)] = file
	}

	l.path, l.files = path, files
}

func (l *lg) SetRolling(rolling bool) {
	l.rolling = rolling
	if l.rolling {
		l.setFileDate(time.Now().Format(logDateFormat))
		go l.runRolling()
	}
}

func (l *lg) runRolling() {
	ticker := time.NewTicker(RollingCheckInterval)
	defer ticker.Stop()
	for {
		curTime := <-ticker.C
		oldDate := l.getPaleFileDate()
		loc := time.Local
		oldTime, err := time.ParseInLocation(logDateFormat, oldDate, loc)
		if err != nil {
			Panic(err)
			return
		}

		// Create log file in X seconds in advance, notice the time zone
		if curTime.Sub(oldTime) >= time.Hour*24-RollingInAdvance {
			l.setFileDate(curTime.Add(RollingInAdvance + 1).Format(logDateFormat))
			l.Open(l.path)
			// sleep and close the old log file handler
			<-time.After(CloseFileDelay)
			err := l.Close(oldDate)
			if err != nil {
				Panic(err)
				return
			}
		}
	}
}

func (l *lg) setFileDate(date string) {
	l.fileDate = strings.TrimRight(date, Dot) + Dot
}

func (l *lg) getPaleFileDate() string {
	return strings.TrimRight(l.fileDate, Dot)
}

func (l *lg) getLevelDate(level LogLevel) LogLevelDate {
	return LogLevelDate(string(level) + l.fileDate)
}

func (l *lg) SetFormat(fm LogFormat) {
	l.fm = fm
	switch l.fm {
	case JsonFormat:
		l.lgr.SetFlags(0)
		break
	case PaleFormat:
		l.lgr.SetFlags(log.LstdFlags | log.Llongfile)
		break
	}
}

type LogJsonFormat struct {
	Level  LogLevel `json:"level"`
	Date   string   `json:"date"`
	Caller string   `json:"caller"`
	Msg    string   `json:"msg"`
}

func (l *lg) ProcessLogFormat(level LogLevel, v ...interface{}) interface{} {
	if l.fm == JsonFormat {
		_, file, line, _ := runtime.Caller(2)
		jsonLog := LogJsonFormat{
			Level:  level,
			Date:   time.Now().Format("2006/01/02/ 15:04:05"),
			Caller: fmt.Sprintf("%s %d", file, line),
			Msg:    fmt.Sprintf(string(bytes.Repeat([]byte("%+v "), len(v))), v...),
		}

		dataByte, _ := json.Marshal(jsonLog)

		return string(dataByte)
	}

	return v
}

func (l *lg) SetFilePrefix(filePrefix string) {
	if filePrefix != "" {
		l.filePrefix = strings.TrimRight(filePrefix, Dot) + Dot
	}
}

func (l *lg) Close(date string) error {
	for _, level := range logLevels {
		file, exist := l.files[LogLevelDate(string(level)+date)]
		if exist {
			err := file.Close()
			if err != nil {
				return err
			}
		}
	}
	return nil
}
