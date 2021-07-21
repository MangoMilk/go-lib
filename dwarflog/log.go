package dwarflog

import (
	"os"
)

type Config struct {
	Format     LogFormat
	Rolling    bool
	Path       string
	FilePrefix string
}

func Setup(cfg *Config) {

	// singleton
	if l == nil {

		if isExist(cfg.Path) != true {
			mkdirErr := os.Mkdir(cfg.Path, 0755)

			if mkdirErr != nil {
				panic("【dwarflog error】Mkdir [" + cfg.Path + "] fail!")
			}
		}

		l = NewDwarfLog()
		l.SetFilePrefix(cfg.FilePrefix)
		l.SetFormat(cfg.Format)
		l.SetRolling(cfg.Rolling)
		l.Open(cfg.Path)
	}
}

func isExist(path string) bool {
	_, err := os.Stat(path)

	if err != nil {
		if os.IsExist(err) {
			return true
		}

		if os.IsNotExist(err) {
			return false
		}
		return false
	}

	return true
}

func Info(v ...interface{}) {
	l.lgr.SetOutput(l.files[l.getLevelDate(InfoLevel)])
	log := l.ProcessLogFormat(InfoLevel, v...)
	l.lgr.Print(log)
}

func Infof(format string, v ...interface{}) {
	l.lgr.SetOutput(l.files[l.getLevelDate(InfoLevel)])
	l.lgr.Printf(format, v...)
}

func Infoln(v ...interface{}) {
	l.lgr.SetOutput(l.files[l.getLevelDate(InfoLevel)])
	log := l.ProcessLogFormat(InfoLevel, v...)
	l.lgr.Println(log)
}

func Error(v ...interface{}) {
	l.lgr.SetOutput(l.files[l.getLevelDate(ErrorLevel)])
	log := l.ProcessLogFormat(ErrorLevel, v...)
	l.lgr.Print(log)
}

func Errorf(format string, v ...interface{}) {
	l.lgr.SetOutput(l.files[l.getLevelDate(ErrorLevel)])
	l.lgr.Printf(format, v...)
}

func Errorrln(v ...interface{}) {
	l.lgr.SetOutput(l.files[l.getLevelDate(ErrorLevel)])
	log := l.ProcessLogFormat(ErrorLevel, v...)
	l.lgr.Println(log)
}

func Fatal(v ...interface{}) {
	l.lgr.SetOutput(l.files[l.getLevelDate(ErrorLevel)])
	log := l.ProcessLogFormat(FatalLevel, v...)
	l.lgr.Fatal(log)
}

func Fatalf(format string, v ...interface{}) {
	l.lgr.SetOutput(l.files[l.getLevelDate(ErrorLevel)])
	l.lgr.Fatalf(format, v...)
}

func Fatalln(v ...interface{}) {
	l.lgr.SetOutput(l.files[l.getLevelDate(ErrorLevel)])
	log := l.ProcessLogFormat(FatalLevel, v...)
	l.lgr.Fatalln(log)
}

func Panic(v ...interface{}) {
	//l.lgr.SetOutput(l.files[l.getLevelDate(ErrorLevel)])
	l.lgr.SetOutput(os.Stderr)
	//log := l.ProcessLogFormat(PanicLevel, v...)
	l.lgr.Panic(v...)
}

func Panicf(format string, v ...interface{}) {
	//l.lgr.SetOutput(l.files[l.getLevelDate(ErrorLevel)])
	l.lgr.SetOutput(os.Stderr)
	l.lgr.Panicf(format, v...)
}

func Panicln(v ...interface{}) {
	//l.lgr.SetOutput(l.files[l.getLevelDate(ErrorLevel)])
	l.lgr.SetOutput(os.Stderr)
	//log := l.ProcessLogFormat(PanicLevel, v...)
	l.lgr.Panicln(v...)
}
