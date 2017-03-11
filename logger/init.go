package logger

import (
	"fmt"

	"github.com/1046102779/wx_relay_server/conf"
	"github.com/astaxie/beego/logs"
)

var (
	LogId string
)

var (
	Logger     *logs.BeeLogger
	LoggerSMTP *logs.BeeLogger
)

func init() {
	loggerFileFilename := fmt.Sprintf("log/%s.log", conf.Cconfig.AppName)
	loggerFileMaxlines := 1000000
	loggerFileMaxsize := 256 * 1024 * 1024
	loggerFileDaily := true
	loggerFileMaxdays := 7
	loggerFileRotate := true
	loggerFileLevel := logs.LevelDebug
	loggerConf := fmt.Sprintf(`{
		        "filename": "%s",
		        "maxlines": %d,
		        "maxsize": %d,
		        "daily": %t,
		        "maxdays": %d,
		        "rotate": %t,
		        "level": %d
		    }`, loggerFileFilename, loggerFileMaxlines, loggerFileMaxsize, loggerFileDaily, loggerFileMaxdays, loggerFileRotate, loggerFileLevel)
	logFuncCallEnable := true
	Logger = logs.NewLogger(10000)
	Logger.EnableFuncCallDepth(logFuncCallEnable)
	Logger.SetLogFuncCallDepth(2)

	Logger.SetLogger("file", loggerConf)
}
