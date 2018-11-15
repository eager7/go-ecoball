// Copyright 2018 The go-ecoball Authors
// This file is part of the go-ecoball library.
//
// The go-ecoball library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ecoball library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ecoball library. If not, see <http://www.gnu.org/licenses/>.

package elog

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"runtime/debug"
	"sync"
	"github.com/ecoball/go-ecoball/common/config"
)

const (
	colorRed = iota + 91
	colorGreen
	colorYellow
	colorBlue
	colorMagenta
)

const defaultSize = 8500000 * 5 //50M

var Log = NewLogger("default", NoticeLog)

const (
	NoticeLog = iota
	DebugLog
	InfoLog
	WarnLog
	ErrorLog
	FatalLog
	MaxLevelLog
)

type Logger interface {
	Notice(a ...interface{})
	Debug(a ...interface{})
	Info(a ...interface{})
	Warn(a ...interface{})
	Error(a ...interface{})
	ErrStack(a ...interface{})
	Fatal(a ...interface{})
	Panic(a ...interface{})
	GetLogger() *log.Logger
	SetLogLevel(level int) error
	GetLogLevel() int
}

type loggerModule struct {
	logger      *log.Logger
	fd          *os.File
	fileName    string
	backupCount int
	moduleName  string
	level       int
	maxSize     int
	curSize     int
	mutex       sync.Mutex
}

func fileOpen(path string) (*os.File, string, error) {
	if fi, err := os.Stat(path); err == nil {
		if !fi.IsDir() {
			return nil, "", fmt.Errorf("open %s: not a directory", path)
		}
	} else if os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0766); err != nil {
			return nil, "", err
		}
	} else {
		return nil, "", err
	}

	var currentTime = time.Now().Format("2006-01-02_15.04.05")
	fileName := path + currentTime + "_LOG.log"
	logfile, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, "", err
	}
	return logfile, fileName, nil
}

func NewLogger(moduleName string, level int) Logger {
	var filePath string
	if config.LogDir == "" {
		filePath = "./"
	} else {
		filePath = config.LogDir
	}
	file, Writer, fileName := initFile(filePath)
	logger := log.New(Writer, "", log.Ldate|log.Lmicroseconds|log.LstdFlags)
	module := loggerModule{
		logger:      logger,
		fd:          file,
		fileName:    fileName,
		backupCount: 10,
		moduleName:  moduleName,
		level:       level,
		maxSize:     defaultSize,
		curSize:     0,
		mutex:       sync.Mutex{},
	}
	return &module
}

func initFile(filePath string) (*os.File, io.Writer, string) {
	if filePath == "" {
		filePath = "/tmp/Log/"
	}
	var fileAndStdoutWrite io.Writer
	logFile, fileName, err := fileOpen(filePath)
	if err != nil {
		fmt.Println("open log file failed: ", err)
		os.Exit(1)
	}

	fileAndStdoutWrite = io.MultiWriter(os.Stdout, logFile)
	return logFile, fileAndStdoutWrite, fileName
}

func (l *loggerModule) GetLogger() *log.Logger {
	return l.logger
}

func (l *loggerModule) SetLogLevel(level int) error {
	if level > MaxLevelLog || level < 0 {
		return errors.New("invalid log level")
	}
	l.level = level
	return nil
}

func (l *loggerModule) GetLogLevel() int {
	return l.level
}

func GetGID() uint64 {
	var buf [64]byte
	b := buf[:runtime.Stack(buf[:], false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}

func getFunctionName() string {
	pc := make([]uintptr, 10)
	runtime.Callers(3, pc)
	f := runtime.FuncForPC(pc[0])

	file, line := f.FileLine(pc[0])
	fileName := filepath.Base(file)

	nameFull := f.Name()
	nameEnd := filepath.Ext(nameFull)

	funcName := strings.TrimPrefix(nameEnd, ".")

	return fileName + ":" + strconv.Itoa(line) + "-" + funcName

}

func (l *loggerModule) Notice(a ...interface{}) {
	if l.level > NoticeLog {
		return
	}
	prefix := []interface{}{"\x1b[" + strconv.Itoa(colorGreen) + "m" + "▶ NOTI " + "[" + l.moduleName + "] " + getFunctionName() + "():" + "\x1b[0m "}
	a = append(prefix, a...)
	l.CheckLogFile(fmt.Sprintln(a...))
	l.logger.Output(2, fmt.Sprintln(a...))

}

func (l *loggerModule) Debug(a ...interface{}) {
	if l.level > DebugLog {
		return
	}
	prefix := []interface{}{"\x1b[" + strconv.Itoa(colorBlue) + "m" + "▶ DEBU " + "[" + l.moduleName + "] " + getFunctionName() + "():" + "\x1b[0m "}
	a = append(prefix, a...)
	l.CheckLogFile(fmt.Sprintln(a...))
	l.logger.Output(2, fmt.Sprintln(a...))
}

func (l *loggerModule) Info(a ...interface{}) {
	if l.level > InfoLog {
		return
	}
	prefix := []interface{}{"\x1b[" + strconv.Itoa(colorYellow) + "m" + "▶ INFO " + "[" + l.moduleName + "] " + getFunctionName() + "():" + "\x1b[0m "}
	a = append(prefix, a...)
	l.CheckLogFile(fmt.Sprintln(a...))
	l.logger.Output(2, fmt.Sprintln(a...))
}

func (l *loggerModule) Warn(a ...interface{}) {
	if l.level > WarnLog {
		return
	}
	prefix := []interface{}{"\x1b[" + strconv.Itoa(colorMagenta) + "m" + "▶ WARN " + "[" + l.moduleName + "] " + getFunctionName() + "():" + "\x1b[0m "}
	a = append(prefix, a...)
	l.CheckLogFile(fmt.Sprintln(a...))
	l.logger.Output(2, fmt.Sprintln(a...))
}

func (l *loggerModule) Error(a ...interface{}) {
	if l.level > ErrorLog {
		return
	}
	prefix := []interface{}{"\x1b[" + strconv.Itoa(colorRed) + "m" + "▶ ERRO " + "[" + l.moduleName + "] " + getFunctionName() + "():" + "\x1b[0m "}
	a = append(prefix, a...)
	l.CheckLogFile(fmt.Sprintln(a...))
	l.logger.Output(2, fmt.Sprintln(a...))
}

func (l *loggerModule) ErrStack(a ...interface{}) {
	if l.level > ErrorLog {
		return
	}
	prefix := []interface{}{"\x1b[" + strconv.Itoa(colorRed) + "m" + "▶ ERRO " + "[" + l.moduleName + "] " + getFunctionName() + "():" + "\x1b[0m "}
	a = append(prefix, a...)
	l.logger.Output(3, fmt.Sprintln(a...))
	debug.PrintStack()
}

func (l *loggerModule) Fatal(a ...interface{}) {
	if l.level > FatalLog {
		return
	}
	prefix := []interface{}{"\x1b[" + strconv.Itoa(colorRed) + "m" + "▶ FATAL " + "[" + l.moduleName + "] " + getFunctionName() + "():" + "\x1b[0m "}
	a = append(prefix, a...)
	debug.PrintStack()
	l.logger.Fatal(a...)
}

func (l *loggerModule) Panic(a ...interface{}) {
	if l.level > FatalLog {
		return
	}
	prefix := []interface{}{"\x1b[" + strconv.Itoa(colorRed) + "m" + "▶ PANIC " + "[" + l.moduleName + "] " + getFunctionName() + "():" + "\x1b[0m "}
	a = append(prefix, a...)
	l.logger.Panic(a...)
}

func (l *loggerModule) CheckLogFile(s string) {
	if l.curSize > l.maxSize {
		l.mutex.Lock()
		defer l.mutex.Unlock()
		l.fd.Close()
		for i := l.backupCount - 1; i > 0; i-- {
			sfn := fmt.Sprintf("%s.%d", l.fileName, i)
			dfn := fmt.Sprintf("%s.%d", l.fileName, i+1)

			os.Rename(sfn, dfn)
		}

		dfn := fmt.Sprintf("%s.1", l.fileName)
		os.Rename(l.fileName, dfn)

		l.fd, _ = os.OpenFile(l.fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		l.curSize = 0
		f, err := l.fd.Stat()
		if err != nil {
			return
		}
		l.curSize = int(f.Size())
		Writer := io.MultiWriter(os.Stdout, l.fd)
		l.logger = log.New(Writer, "", log.Ldate|log.Lmicroseconds|log.LstdFlags)

	} else {
		l.curSize += len(s)
	}
}
