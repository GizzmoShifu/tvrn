package logx

import (
  "fmt"
  "log"
)

type Logger struct { level string }

func New(level string) *Logger { return &Logger{level: level} }
func (l *Logger) Debugf(f string, a ...any) { if l.level == "debug" { log.Printf("DEBUG "+f, a...) } }
func (l *Logger) Infof(f string, a ...any)  { log.Printf("INFO  "+f, a...) }
func (l *Logger) Warnf(f string, a ...any)  { log.Printf("WARN  "+f, a...) }
func (l *Logger) Errorf(f string, a ...any) { log.Printf("ERROR "+f, a...) }
func (l *Logger) Println(v ...any)          { fmt.Println(v...) }
