package redisadapter

import (
	"io/ioutil"
	"log"
	"os"
)

// The Logger interface generalizes the Entry and Logger types
type Logger interface {
	Debugf(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})

	Debug(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})

	Debugln(args ...interface{})
	Warnln(args ...interface{})
	Errorln(args ...interface{})
}

// NewStdLog return a new instance of "log" to stdout
func NewStdLog() Logger {
	return &stdLog{
		logger: log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

// NewNoLog return a no log instance
func NewNoLog() Logger {
	return &stdLog{
		logger: log.New(ioutil.Discard, "", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

type stdLog struct {
	logger *log.Logger
}

func (s *stdLog) Warnf(format string, args ...interface{}) {
	s.logger.Printf(format, args)
}

func (s *stdLog) Errorf(format string, args ...interface{}) {
	s.logger.Printf(format, args)
}

func (s *stdLog) Debug(args ...interface{}) {
	s.logger.Print(args)
}

func (s *stdLog) Warn(args ...interface{}) {
	s.logger.Print(args)
}

func (s *stdLog) Error(args ...interface{}) {
	panic("implement me")
}

func (s *stdLog) Debugln(args ...interface{}) {
	s.logger.Println(args)
}

func (s *stdLog) Warnln(args ...interface{}) {
	s.logger.Println(args)
}

func (s *stdLog) Errorln(args ...interface{}) {
	s.logger.Println(args)
}

func (s *stdLog) Debugf(format string, args ...interface{}) {
	s.logger.Printf(format, args)
}
