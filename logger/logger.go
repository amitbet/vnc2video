package logger

import "fmt"

var simpleLogger = SimpleLogger{level: LogLevelNone, errorCh: nil}

type Logger interface {
	Trace(v ...interface{})
	Tracef(format string, v ...interface{})
	Debug(v ...interface{})
	Debugf(format string, v ...interface{})
	Info(v ...interface{})
	Infof(format string, v ...interface{})
	DebugfNoCR(format string, v ...interface{})
}
type LogLevel int

const (
	LogLevelTrace LogLevel = iota
	LogLevelDebug
	LogLevelInfo
	LogLevelNone
)

type SimpleLogger struct {
	level   LogLevel
	errorCh chan error
}

func (sl *SimpleLogger) Trace(v ...interface{}) {
	if sl.level <= LogLevelTrace {
		arr := []interface{}{"[Trace]"}
		for _, item := range v {
			arr = append(arr, item)
		}

		fmt.Println(arr...)
	}
}
func (sl *SimpleLogger) Tracef(format string, v ...interface{}) {
	if sl.level <= LogLevelTrace {
		fmt.Printf("[Trace] "+format+"\n", v...)
	}
}

func (sl *SimpleLogger) Debug(v ...interface{}) {
	if sl.level <= LogLevelDebug {
		arr := []interface{}{"[Debug]"}
		for _, item := range v {
			arr = append(arr, item)
		}

		fmt.Println(arr...)
	}
}
func (sl *SimpleLogger) Debugf(format string, v ...interface{}) {
	if sl.level <= LogLevelDebug {
		fmt.Printf("[Debug] "+format+"\n", v...)
	}
}
func (sl *SimpleLogger) Info(v ...interface{}) {
	if sl.level <= LogLevelInfo {
		arr := []interface{}{"[Info ]"}
		for _, item := range v {
			arr = append(arr, item)
		}
		fmt.Println(arr...)
	}
}
func (sl *SimpleLogger) DebugfNoCR(format string, v ...interface{}) {
	if sl.level <= LogLevelDebug {
		fmt.Printf("[Info ] "+format, v...)
	}
}

func (sl *SimpleLogger) Infof(format string, v ...interface{}) {
	if sl.level <= LogLevelInfo {
		fmt.Printf("[Info ] "+format+"\n", v...)
	}
}

func Trace(v ...interface{}) {
	simpleLogger.Trace(v...)
}
func Tracef(format string, v ...interface{}) {
	simpleLogger.Tracef(format, v...)
}

func Debug(v ...interface{}) {
	simpleLogger.Debug(v...)
}
func Debugf(format string, v ...interface{}) {
	simpleLogger.Tracef(format, v...)
}

func Info(v ...interface{}) {
	simpleLogger.Info(v...)
}
func Infof(format string, v ...interface{}) {
	simpleLogger.Infof(format, v...)
}
func DebugfNoCR(format string, v ...interface{}) {
	simpleLogger.DebugfNoCR(format, v...)
}
