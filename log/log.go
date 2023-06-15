package log

import (
	"sync"

	"github.com/xtls/xray-core/common/log"
)

type RawLogger interface {
	OnAccessLog(message string)
	OnDNSLog(message string)
	OnGeneralMessage(severity int, message string)
}

type Logger struct {
	sync.RWMutex
	raw RawLogger
}

func New(raw RawLogger) *Logger {
	return &Logger{
		raw: raw,
	}
}

func (l *Logger) Handle(msg log.Message) {
	l.RLock()
	defer l.RUnlock()
	if l.raw == nil {
		return
	}
	switch msg := msg.(type) {
	case *log.AccessMessage:
		l.raw.OnAccessLog(msg.String())
	case *log.DNSLog:
		l.raw.OnDNSLog(msg.String())
	case *log.GeneralMessage:
		l.raw.OnGeneralMessage(int(msg.Severity), msg.String())
	default:
	}
}
