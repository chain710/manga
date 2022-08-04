package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

//goland:noinspection ALL
var (
	slogger *zap.SugaredLogger
)

type logConfig struct {
	config  zap.Config
	options []zap.Option
}

type Option func(*logConfig)

func WithLogLevel(l zapcore.Level) Option {
	return func(c *logConfig) {
		c.config.Level.SetLevel(l)
	}
}

func WithLogEncoding(name string) Option {
	return func(c *logConfig) {
		c.config.Encoding = name
	}
}

func Init(opts ...Option) {
	zapConfig := zap.NewProductionConfig()
	zapConfig.Encoding = "console"
	zapConfig.EncoderConfig.StacktraceKey = ""
	zapConfig.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	c := logConfig{
		config:  zapConfig,
		options: []zap.Option{},
	}
	for _, opt := range opts {
		opt(&c)
	}

	if l, err := c.config.Build(c.options...); err != nil {
		panic(err)
	} else {
		initLogger(l)
	}
}

func init() {
	Init()
}

func With(args ...interface{}) *zap.SugaredLogger {
	return slogger.With(args...)
}

func Debugf(template string, args ...interface{}) {
	slogger.Debugf(template, args...)
}

func Infof(template string, args ...interface{}) {
	slogger.Infof(template, args...)
}

func Warnf(template string, args ...interface{}) {
	slogger.Warnf(template, args...)
}

// Errorf uses fmt.Sprintf to log a templated message.
func Errorf(template string, args ...interface{}) {
	slogger.Errorf(template, args...)
}

func Panicf(template string, args ...interface{}) {
	slogger.Panicf(template, args...)
}

func Debugw(msg string, keysAndValues ...interface{}) {
	slogger.Debugw(msg, keysAndValues...)
}

func Infow(msg string, keysAndValues ...interface{}) {
	slogger.Infow(msg, keysAndValues...)
}

func Warnw(msg string, keysAndValues ...interface{}) {
	slogger.Warnw(msg, keysAndValues...)
}

func Errorw(msg string, keysAndValues ...interface{}) {
	slogger.Errorw(msg, keysAndValues...)
}

func Panicw(msg string, keysAndValues ...interface{}) {
	slogger.Panicw(msg, keysAndValues...)
}

func initLogger(l *zap.Logger) {
	slogger = l.Sugar()
}

func Logger() *zap.Logger {
	return slogger.Desugar()
}
