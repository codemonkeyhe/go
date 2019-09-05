// +build !windows,!nacl,!plan9

package applog

import (
	"log/syslog"

	"github.com/op/go-logging"
)

type Priority int

//var (
//	LOG_EMERG   Priority = syslog.LOG_EMERG
//	LOG_ALERT   Priority = syslog.LOG_ALERT
//	LOG_CRIT    Priority = syslog.LOG_CRIT
//	LOG_ERR     Priority = syslog.LOG_ERR
//	LOG_WARNING Priority = syslog.LOG_WARNING
//	LOG_NOTICE  Priority = syslog.LOG_NOTICE
//	LOG_INFO    Priority = syslog.LOG_INFO
//	LOG_DEBUG   Priority = syslog.LOG_DEBUG

//	LOG_LOCAL0 Priority = syslog.LOG_LOCAL0
//)

func MyNewSyslogBackendPriority(name string) (b *logging.SyslogBackend, err error) {
	return logging.NewSyslogBackendPriority(name, syslog.LOG_LOCAL0|syslog.LOG_DEBUG)
}
