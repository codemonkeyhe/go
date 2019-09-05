package applog

import (
	"bytes"
	"runtime"

	//	"log/syslog"
	"fmt"
	"os"
	"path"

	"github.com/op/go-logging"
)

type LogWatcher interface {
	OnErrorLog(file string, line int, fn string, text string)
}

var watchers []LogWatcher
var watcherFmt logging.Formatter

func AddLogWatcher(w LogWatcher) {
	watchers = append(watchers, w)
}

type AppLogBackend struct {
	backend  logging.Backend
	formater logging.Formatter
}

func (this *AppLogBackend) Log(level logging.Level, calldepth int, rec *logging.Record) error {
	if level <= logging.ERROR {
		if watchers != nil {
			pc, file, line, _ := runtime.Caller(4)
			fn := runtime.FuncForPC(pc)

			var buf bytes.Buffer
			watcherFmt.Format(calldepth+1, rec, &buf)
			for _, watcher := range watchers {
				watcher.OnErrorLog(file, line, fn.Name(), buf.String())
			}
		}
	}
	return this.backend.Log(level, calldepth+1, rec)
}

func newApplogBackend(b logging.Backend) *AppLogBackend {
	f := logging.MustStringFormatter("%{shortfile} %{longfunc} %{message}")
	return &AppLogBackend{b, f}
}

func init() {
	fmt.Println("%+v", os.Args)
	process_name := path.Base(os.Args[0])

	// 定义非syslog输出的日志格式
	watcherFmt = logging.MustStringFormatter("%{module} %{longfunc} %{message}")

	//这样改造支持跨平台编译
	backend1, err := MyNewSyslogBackendPriority(process_name)
	if err == nil {
		format := logging.MustStringFormatter("%{color} %{module} %{shortfile} %{longfunc} %{color:reset} %{message}")
		backend1Formatter := newApplogBackend(logging.NewBackendFormatter(backend1, format))
		logging.SetBackend(backend1Formatter)

		//log := MustGetLogger("applog")
		SetLogLevel("DEBUG")
		//log.Infof("syslog installed and %v start!", process_name)
	} else {
		panic(err)
	}
}
func SetLogLevel(logLevelStr string) error {
	var appLogLevel logging.Level
	if level, err := logging.LogLevel(logLevelStr); err != nil {
		appLogLevel = logging.DEBUG
	} else {
		appLogLevel = level
	}
	logging.SetLevel(appLogLevel, "")
	return nil
}

func MustGetLogger(module string) *logging.Logger {
	log := logging.MustGetLogger(module)
	return log
}
