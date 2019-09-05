package main

import (
	"fmt"
	"os"
	"time"

	"common/applog/go-logging"
)

var log = logging.MustGetLogger("example")

// Example format string. Everything except the message has a custom color
// which is dependent on the log level. Many fields have a custom output
// formatting too, eg. the time returns the hour down to the milli second.
var format = logging.MustStringFormatter(
	"%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}",
)

// Password is just an example type implementing the Redactor interface. Any
// time this is logged, the Redacted() function will be called.
type Password string

func (p Password) Redacted() interface{} {
	return logging.Redact(string(p))
}

func main() {
	// For demo purposes, create two backend for os.Stderr.
	//be1是最原始的输出，但是限定了只打Error级别以上的日志
	backend1 := logging.NewLogBackend(os.Stdout, "", 0)
	//be2是带格式的输出，且输出所有级别的日志
	backend2 := logging.NewLogBackend(os.Stdout, "", 0)
	//be3打印到文件，格式同 be2
	logFile, err := os.OpenFile("log.txt", os.O_WRONLY, 0666)
	if err != nil {
		fmt.Println(err)
	}
	backend3 := logging.NewLogBackend(logFile, "", 0)
	backend3Formatter := logging.NewBackendFormatter(backend3, format)

	// For messages written to backend2 we want to add some additional
	// information to the output, including the used log level and the name of
	// the function.
	backend2Formatter := logging.NewBackendFormatter(backend2, format)

	// Only errors and more severe messages should be sent to backend1
	backend1Leveled := logging.AddModuleLevel(backend1)
	backend1Leveled.SetLevel(logging.ERROR, "")

	// Set the backends to be used.
	logging.SetBackend(backend1Leveled, backend2Formatter, backend3Formatter)

	log.Debug("debug %s", Password("secret"))
	log.Notice("notice")
	log.Warning("warning")
	log.Info("info")

	log.Error("err")
	log.Critical("crit")
	time.Sleep(time.Second * 2)
}
