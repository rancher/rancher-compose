package main

import (
	"flag"
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/rancher/go-machine-service/events"
	"github.com/rancher/go-machine-service/handlers"
)

var (
	GITCOMMIT = "HEAD"
)

func main() {
	processCmdLineFlags()

	log.WithFields(log.Fields{
		"gitcommit": GITCOMMIT,
	}).Info("Starting go-machine-service...")
	eventHandlers := map[string]events.EventHandler{
		"physicalhost.create":    handlers.CreateMachine,
		"physicalhost.bootstrap": handlers.ActivateMachine,
		"physicalhost.remove":    handlers.PurgeMachine,
		"ping":                   handlers.PingNoOp,
	}

	apiUrl := os.Getenv("CATTLE_URL")
	accessKey := os.Getenv("CATTLE_ACCESS_KEY")
	secretKey := os.Getenv("CATTLE_SECRET_KEY")

	router, err := events.NewEventRouter("goMachineService", 2000, apiUrl, accessKey, secretKey,
		nil, eventHandlers, "physicalhost", 10)
	if err != nil {
		log.WithFields(log.Fields{
			"Err": err,
		}).Error("Unable to create EventRouter")
	} else {
		err := router.Start(nil)
		if err != nil {
			log.WithFields(log.Fields{
				"Err": err,
			}).Error("Unable to start EventRouter")
		}
	}
	log.Info("Exiting go-machine-service...")
}

func processCmdLineFlags() {
	// Define command line flags
	logLevel := flag.String("loglevel", "info", "Set the default loglevel (default:info) [debug|info|warn|error]")
	version := flag.Bool("v", false, "read the version of the go-machine-service")
	output := flag.String("o", "", "set the output file to write logs into, default is stdout")

	flag.Parse()

	if *output != "" {
		var f *os.File
		if _, err := os.Stat(*output); os.IsNotExist(err) {
			f, err = os.Create(*output)
			if err != nil {
				fmt.Printf("could not create file=%s for logging, err=%v\n", *output, err)
				os.Exit(1)
			}
		} else {
			var err error
			f, err = os.OpenFile(*output, os.O_RDWR|os.O_APPEND, 0)
			if err != nil {
				fmt.Printf("could not open file=%s for writing, err=%v\n", *output, err)
				os.Exit(1)
			}
		}
		log.SetFormatter(&log.JSONFormatter{})
		log.SetOutput(f)
	}

	if *version {
		fmt.Printf("go-machine-service\t gitcommit=%s\n", GITCOMMIT)
		os.Exit(0)
	}

	// Process log level.  If an invalid level is passed in, we simply default to info.
	if parsedLogLevel, err := log.ParseLevel(*logLevel); err == nil {
		log.WithFields(log.Fields{
			"logLevel": *logLevel,
		}).Info("Setting log level")
		log.SetLevel(parsedLogLevel)
	}
}
