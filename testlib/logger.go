package testlib

import (
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"
	logrusTest "github.com/sirupsen/logrus/hooks/test"
)

type testingWriter struct {
	tb testing.TB
}

func (tw *testingWriter) Write(b []byte) (int, error) {
	tw.tb.Log(strings.TrimSpace(string(b)))
	return len(b), nil
}

func MakeLogger(tb testing.TB) log.FieldLogger {
	logger := log.New()
	logger.SetOutput(&testingWriter{tb})
	logger.SetLevel(log.TraceLevel)

	return logger
}

func MakeLoggerWithHooks(tb testing.TB) (log.FieldLogger, *logrusTest.Hook) {
	logger, hook := logrusTest.NewNullLogger()
	logger.SetOutput(&testingWriter{tb})
	logger.SetLevel(log.TraceLevel)

	return logger, hook
}
