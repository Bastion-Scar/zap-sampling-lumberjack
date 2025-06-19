package main

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"net/http"
	"os"
	"time"
)

var logger *zap.Logger

func customLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		logger.Info("request",
			zap.Duration("latency", latency),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", status),
			zap.String("ip", c.ClientIP()),
		)
	}
}

func initLogger() (*zap.Logger, error) {
	lumberjackLogger := &lumberjack.Logger{
		Filename:   "./app.log",
		MaxSize:    5,
		MaxBackups: 3,
		MaxAge:     28,
		Compress:   true,
	}

	writeSyncer := zapcore.AddSync(lumberjackLogger)

	consoleEncoder := zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
		TimeKey:      "time",
		LevelKey:     "level",
		CallerKey:    "caller",
		MessageKey:   "msg",
		EncodeTime:   zapcore.ISO8601TimeEncoder,
		EncodeLevel:  zapcore.CapitalColorLevelEncoder,
		EncodeCaller: zapcore.ShortCallerEncoder,
	})

	jsonEncoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		TimeKey:      "time",
		LevelKey:     "level",
		CallerKey:    "caller",
		MessageKey:   "msg",
		EncodeTime:   zapcore.ISO8601TimeEncoder,
		EncodeLevel:  zapcore.LowercaseLevelEncoder,
		EncodeCaller: zapcore.ShortCallerEncoder,
	})

	consoleCore := zapcore.NewCore(consoleEncoder, os.Stdout, zapcore.InfoLevel)
	jsonCore := zapcore.NewCore(jsonEncoder, writeSyncer, zapcore.DebugLevel)

	core := zapcore.NewTee(consoleCore, jsonCore)

	sampledCore := zapcore.NewSamplerWithOptions(core, time.Second, 100, 100)

	return zap.New(sampledCore, zap.AddCaller()), nil

}

func main() {
	var err error

	logger, err = initLogger()
	if err != nil {
		panic(err)
	}
	defer func() {
		err = logger.Sync()
		if err != nil {
			panic(err)
		}
	}()

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(customLogger())
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.Run(":8080")

}
