//+build windows plan9

package applog

import (
	"github.com/op/go-logging"
)

//type Priority int

//type Priority logging.Priority

const severityMask = 0x07
const facilityMask = 0xf8

const (
	// Severity.

	// From /usr/include/sys/syslog.h.
	// These are the same on Linux, BSD, and OS X.
	LOG_EMERG logging.Priority = iota
	//LOG_EMERG int = 1
	LOG_ALERT
	LOG_CRIT
	LOG_ERR
	LOG_WARNING
	LOG_NOTICE
	LOG_INFO
	LOG_DEBUG
)

const (
	// Facility.

	// From /usr/include/sys/syslog.h.
	// These are the same up to LOG_FTP on Linux, BSD, and OS X.
	LOG_KERN logging.Priority = iota << 3
	//LOG_KERN int = 1 << 3
	LOG_USER
	LOG_MAIL
	LOG_DAEMON
	LOG_AUTH
	LOG_SYSLOG
	LOG_LPR
	LOG_NEWS
	LOG_UUCP
	LOG_CRON
	LOG_AUTHPRIV
	LOG_FTP
	_ // unused
	_ // unused
	_ // unused
	_ // unused
	LOG_LOCAL0
	LOG_LOCAL1
	LOG_LOCAL2
	LOG_LOCAL3
	LOG_LOCAL4
	LOG_LOCAL5
	LOG_LOCAL6
	LOG_LOCAL7
)

func MyNewSyslogBackendPriority(name string) (b *logging.SyslogBackend, err error) {
	return logging.NewSyslogBackendPriority(name, LOG_LOCAL0|LOG_DEBUG)
}
