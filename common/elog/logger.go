package elog

import (
	"fmt"
	"github.com/spf13/viper"
	"io"
	"net/http"
	"os"
	"strconv"
	"github.com/ecoball/go-ecoball/common/elog/logbunny"
	"github.com/ecoball/go-ecoball/common/elog/bunnystub"
	"runtime/debug"
	"github.com/ecoball/go-ecoball/common/config"
)

const (
	appendedFormat = "%s %s"
	configFileName = "log_config.toml"
)
const (
	NoticeLog = iota
	DebugLog
	InfoLog
	WarnLog
	ErrorLog
	FatalLog
	MaxLevelLog
)


type loggerOpt struct {
	debugLevel         logbunny.LogLevel
	loggerType         logbunny.LogType
	withCaller         bool
	loggerEncoder      logbunny.EncoderType
	timePattern        string
	debugLogFilename   string
	infoLogFilename    string
	warnLogFilename    string
	errorLogFilename   string
	fatalLogFilename   string
	httpPort           string
	rollingTimePattern string
	skip               int
	logger             logbunny.Logger
}

var Log Logger //default logger
var logOpt *loggerOpt
var levelHandler *logbunny.HTTPHandler

func init() {
	fmt.Println("init log, dir:", config.LogDir)
	var err error
	logOpt, err = newLoggerOpt()
	if err != nil {
		panic(err)
	}

	logFilename := map[logbunny.LogLevel]string{
		logbunny.DebugLevel: config.LogDir+logOpt.debugLogFilename,
		logbunny.InfoLevel:  config.LogDir+logOpt.infoLogFilename,
		logbunny.WarnLevel:  config.LogDir+logOpt.warnLogFilename,
		logbunny.ErrorLevel: config.LogDir+logOpt.errorLogFilename,
		logbunny.FatalLevel: config.LogDir+logOpt.fatalLogFilename,
	}

	outputWriter := make(map[logbunny.LogLevel][]io.Writer)
	for level, file := range logFilename {
		logFileWriter, err := newLogFile(file, logOpt.rollingTimePattern)
		if err != nil {
			panic(err)
		}
		if config.OutputToTerminal {
			outputWriter[level] = []io.Writer{logFileWriter, os.Stdout}
		} else {
			outputWriter[level] = []io.Writer{logFileWriter}
		}
	}

	zapCfg := &logbunny.Config{
		Type:        logOpt.loggerType,
		Level:       logOpt.debugLevel,
		Encoder:     logOpt.loggerEncoder,
		WithCaller:  logOpt.withCaller,
		Out:         nil,
		WithNoLock:  false,
		TimePattern: logOpt.timePattern,
		Skip:        logOpt.skip,
	}
	logOpt.logger, err = logbunny.FilterLogger(zapCfg, outputWriter)
	if err != nil {
		panic(err)
	}

	logbunny.SetCallerSkip(3)
	// log.Warp()

	levelHandler = logbunny.NewHTTPHandler(logOpt.logger)
	http.HandleFunc("/logoutLevel", func(w http.ResponseWriter, r *http.Request) {
		levelHandler.ServeHTTP(w, r)
	})
	go http.ListenAndServe(logOpt.httpPort, nil)
	Log = logOpt
}

func newLoggerOpt() (*loggerOpt, error) {
	/*conf, err := ioutil.ReadFile("." + "/" + configFileName)
	if err != nil {
		return nil, err
	}
	viper.SetConfigType("toml")

	err = viper.ReadConfig(bytes.NewBuffer(conf))
	if err != nil {
		return nil, err
	}*/

	return &loggerOpt{
		debugLevel:         logbunny.LogLevel(viper.GetInt("logbunny.debug_level")),
		loggerType:         logbunny.LogType(viper.GetInt("logbunny.loggerType")),
		withCaller:         viper.GetBool("logbunny.with_caller"),
		loggerEncoder:      logbunny.EncoderType(viper.GetInt("logbunny.logger_encoder")),
		timePattern:        viper.GetString("logbunny.time_pattern"),
		httpPort:           viper.GetString("logbunny.http_port"),
		debugLogFilename:   viper.GetString("logbunny.debug_log_filename"),
		infoLogFilename:    viper.GetString("logbunny.info_log_filename"),
		warnLogFilename:    viper.GetString("logbunny.warn_log_filename"),
		errorLogFilename:   viper.GetString("logbunny.error_log_filename"),
		fatalLogFilename:   viper.GetString("logbunny.fatal_log_filename"),
		rollingTimePattern: viper.GetString("logbunny.rolling_time_pattern"),
		skip:               viper.GetInt("logbunny.skip"),
	}, nil
}
func newLogFile(logPath string, rollingTimePattern string) (io.Writer, error) {
	if file := stdOutput(logPath); file != nil {
		return file, nil
	} else {
		file, err := bunnystub.NewIOWriter(logPath, bunnystub.TimeRotate, bunnystub.WithTimePattern(rollingTimePattern))
		//file, err := bunnystub.NewIOWriter(logPath, bunnystub.VolumeRotate, bunnystub.WithTimePattern(rollingTimePattern))
		if err != nil {
			return nil, err
		}
		return file, nil
	}
}
func stdOutput(logPath string) *os.File {
	if logPath == "stdout" {
		return os.Stdout
	}
	if logPath == "stderr" {
		return os.Stderr
	}
	return nil
}

const (
	colorRed = iota + 91
	colorGreen
	colorYellow
	colorBlue
	colorMagenta
)

type Logger interface {
	Notice(a ...interface{})
	Debug(a ...interface{})
	Info(a ...interface{})
	Warn(a ...interface{})
	Error(a ...interface{})
	Fatal(a ...interface{})
	Panic(a ...interface{})
	GetLogger() logbunny.Logger
	ErrStack()
	SetLogLevel(level int) error
	//GetLogLevel() int
}

func NewLogger(moduleName string, level int) Logger {
	return logOpt
}

func (l *loggerOpt) Notice(a ...interface{}) {
	if l.loggerEncoder == 0 {
		l.logger.Debug(fmt.Sprintln(a...))
	} else {
		msg := "\x1b[" + strconv.Itoa(colorGreen) + "m" + "▶ " + fmt.Sprintln(a...) + "\x1b[0m"
		l.logger.Debug(msg)
	}
}

func (l *loggerOpt) Debug(a ...interface{}) {
	if l.loggerEncoder == 0 {
		l.logger.Debug(fmt.Sprintln(a...))
	} else {
		msg := "\x1b[" + strconv.Itoa(colorBlue) + "m" + "▶ " + fmt.Sprintln(a...) + "\x1b[0m"
		l.logger.Debug(msg)
	}
}

func (l *loggerOpt) Info(a ...interface{}) {
	if l.loggerEncoder == 0 {
		l.logger.Info(fmt.Sprintln(a...))
	} else {
		msg := "\x1b[" + strconv.Itoa(colorYellow) + "m" + "▶ " + fmt.Sprintln(a...) + "\x1b[0m"
		l.logger.Info(msg)
	}
}

func (l *loggerOpt) Warn(a ...interface{}) {
	if l.loggerEncoder == 0 {
		l.logger.Warn(fmt.Sprintln(a...))
	} else {
		msg := "\x1b[" + strconv.Itoa(colorMagenta) + "m" + "▶ " + fmt.Sprintln(a...) + "\x1b[0m"
		l.logger.Warn(msg)
	}
}

func (l *loggerOpt) Error(a ...interface{}) {
	if l.loggerEncoder == 0 {
		l.logger.Error(fmt.Sprintln(a...))
	} else {
		msg := "\x1b[" + strconv.Itoa(colorRed) + "m" + "▶ " + fmt.Sprintln(a...) + "\x1b[0m"
		l.logger.Error(msg)
	}
}

func (l *loggerOpt) Fatal(a ...interface{}) {
	if l.loggerEncoder == 0 {
		l.logger.Fatal(fmt.Sprintln(a...))
	} else {
		msg := "\x1b[" + strconv.Itoa(colorYellow) + "m" + "▶ " + fmt.Sprintln(a...) + "\x1b[0m"
		l.logger.Fatal(msg)
	}
}

func (l *loggerOpt) Panic(a ...interface{}) {
	if l.loggerEncoder == 0 {
		l.logger.Panic(fmt.Sprintln(a...))
	} else {
		msg := "\x1b[" + strconv.Itoa(colorYellow) + "m" + "▶ " + fmt.Sprintln(a...) + "\x1b[0m"
		l.logger.Panic(msg)
	}
	panic(fmt.Sprintln(a...))
}

func (l *loggerOpt) ErrStack() {
	l.Error(string(debug.Stack()))

}

func (l *loggerOpt) SetLogLevel(level int) error {
	l.logger.SetLevel(logbunny.LogLevel(level))
	return nil
}

func (l *loggerOpt) GetLogger() logbunny.Logger {
	return l.logger
}

func NewZapLogger() logbunny.Logger {
	return logOpt.logger
}
