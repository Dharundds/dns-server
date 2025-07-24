package logger

import (
	"os"
	"path/filepath"
	"strconv"

	"github.com/natefinch/lumberjack"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type LoggerOption func(l *LoggerConfig)

// LoggerConfig holds the configuration for the logger
type LoggerConfig struct {
	LogFilePath string
	MaxSize     int
	MaxBackups  int
	Compress    bool
	Level       string
}

// InitLogger initializes the logger with the provided configuration
func initLogger(config LoggerConfig) {
	rotatingLogger := &lumberjack.Logger{
		Filename:   config.LogFilePath,
		MaxSize:    config.MaxSize,
		MaxBackups: config.MaxBackups,
		Compress:   config.Compress,
	}
	level, err := zerolog.ParseLevel(config.Level)
	if err != nil {
		log.Error().Msgf("Error while parsing level: %v", err)
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(level)
	}

	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		return filepath.Base(file) + ":" + strconv.Itoa(line)
	}
	consoleLogger := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: ("2006-01-02T15:04:05Z")}
	fileLogger := zerolog.ConsoleWriter{Out: rotatingLogger, TimeFormat: ("2006-01-02T15:04:05Z"), NoColor: true}
	log.Logger = log.Output(zerolog.MultiLevelWriter(fileLogger, consoleLogger)).With().Caller().Logger()
	zerolog.TimeFieldFormat = ("2006-01-02T15:04:05Z")
}

func NewLogger(opts ...LoggerOption) {
	logger := &LoggerConfig{}
	for _, opt := range opts {
		opt(logger)
	}
	initLogger(*logger)
}

func WithLogFilePath(logFilePath string) LoggerOption {
	return func(l *LoggerConfig) {
		l.LogFilePath = logFilePath + "/dns-server.log"
	}
}

func WithMaxSize(maxSize int) LoggerOption {
	return func(l *LoggerConfig) {
		l.MaxSize = maxSize
	}
}

func WithLevel(level string) LoggerOption {
	return func(l *LoggerConfig) {
		l.Level = level
	}
}

func WithMaxBackups(maxBackups int) LoggerOption {
	return func(l *LoggerConfig) {
		l.MaxBackups = maxBackups
	}
}

func WithCompress(compress bool) LoggerOption {
	return func(l *LoggerConfig) {
		l.Compress = compress
	}
}
