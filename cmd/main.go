package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gsakun/alarmtransfer/config"
	"github.com/gsakun/alarmtransfer/db"
	models "github.com/gsakun/alarmtransfer/model"
	colorable "github.com/mattn/go-colorable"
	log "github.com/sirupsen/logrus"
	"github.com/snowzach/rotatefilehook"
	"gopkg.in/alecthomas/kingpin.v2"
)

func init() {
	loglevel := os.Getenv("LOG_LEVEL")
	var logLevel log.Level
	log.Infof("loglevel env is %s", loglevel)
	if loglevel == "debug" {
		log.SetLevel(log.DebugLevel)
		logLevel = log.DebugLevel
		log.Infof("log level is %s", loglevel)
		log.SetReportCaller(true)
	} else {
		log.SetLevel(log.InfoLevel)
		logLevel = log.InfoLevel
		log.Infoln("log level is normal")
	}
	rotateFileHook, err := rotatefilehook.NewRotateFileHook(rotatefilehook.RotateFileConfig{
		Filename:   "logs/alarmtransfer.log",
		MaxSize:    50, // megabytes
		MaxBackups: 3,
		MaxAge:     28, //days
		Level:      logLevel,
		Formatter: &log.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
		},
	})
	log.SetOutput(colorable.NewColorableStdout())
	if err != nil {
		log.Fatalf("Failed to initialize file rotate hook: %v", err)
	}
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		ForceColors:     true,
	})
	log.AddHook(rotateFileHook)
}

func main() {
	var (
		listenAddress = kingpin.Flag(
			"web.listen-address",
			"The address to listen on for web interface.",
		).Default(":8080").String()
		configFile = kingpin.Flag(
			"config.file",
			"Path to the configuration file.",
		).Default("config.yml").String()
	)

	kingpin.Version("alarmtransfer v1.0")
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
	log.Infof("Config name %s", *configFile)
	conf, err := config.LoadFile(*configFile)
	if err != nil {
		log.Fatalf("Parse Config Failed, Please Check Config %v", err)
	}
	db.Init(*conf.DbConfig)

	go db.SyncMap()

	router := gin.Default()

	router.POST("/send", handler)
	router.GET("/health", health)
	router.Run(*listenAddress)
}

func health(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "health",
	})
}

func handler(c *gin.Context) {
	var promMessage models.WebhookMessage
	if err := json.NewDecoder(c.Request.Body).Decode(&promMessage); err != nil {
		log.Errorf("Cannot decode prometheus webhook JSON request %v", err)
		c.JSON(400, gin.H{
			"status":  "failed",
			"errinfo": "parse failed",
		})
	} else {
		//log.Infof("Prometheus Message %v", promMessage)
		err := db.HandleMessage(promMessage)
		if err != nil {
			c.JSON(400, gin.H{
				"status":     "failed",
				"errmessage": fmt.Sprintf("errinfo %v", err),
			})
		} else {
			c.JSON(200, gin.H{
				"status": "success",
			})
		}
	}
}
