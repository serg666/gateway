package config

import (
	"os"
	"fmt"
	"time"
	"flag"
	"syscall"
	"context"
	"net"
	"net/http"
	"os/signal"
	"gopkg.in/yaml.v2"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/gin-contrib/requestid"
	"github.com/serg666/repository"
)

// Config struct for webapp config
type Config struct {
	Client struct {
		Timeout struct {
			// Read specifies a time limit for requests made by this
			// Client. The timeout includes connection time, any
			// redirects, and reading the response body.
			Read time.Duration `yaml:"read"`

			// Connect is the maximum amount of time a dial will wait for
			// a connect to complete.
			Connect time.Duration `yaml:"connect"`
		} `yaml:"timeout"`
	} `yaml:"client"`
	Server struct {
		// Host is the local machine IP Address to bind the HTTP Server to
		Host string `yaml:"host"`

		// Port is the local machine TCP Port to bind the HTTP Server to
		Port string `yaml:"port"`
		Timeout	struct {
			// Server is the general server timeout to use
			// for graceful shutdowns
			Server time.Duration `yaml:"server"`

			// Write is the amount of time to wait until an HTTP server
			// write opperation is cancelled
			Write time.Duration `yaml:"write"`

			// Read is the amount of time to wait until an HTTP server
			// read operation is cancelled
			Read time.Duration `yaml:"read"`

			// Read is the amount of time to wait
			// until an IDLE HTTP session is closed
			Idle time.Duration `yaml:"idle"`
		} `yaml:"timeout"`
	} `yaml:"server"`
	LogRus struct {
		Level logrus.Level `yaml:"level"`
	} `yaml:"logrus"`
	Alfabank struct {
		Ecom struct {
			Url string `yaml:"url"`
		} `yaml:"ecom"`
	} `yaml:"alfabank"`
	Databases struct {
		Default struct {
			Dsn string `yaml:"dsn"`
		} `yaml:"default"`
	} `yaml:"databases"`
}

func (cfg *Config) LogRusLogger(c interface{}) logrus.FieldLogger {
	logger := logrus.New()
	// @todo: somehow to configure logger from config
	logger.SetLevel(cfg.LogRus.Level)
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		DisableLevelTruncation: true,
		ForceColors: true,
	})

	var rid string
	if c != nil {
		rid = requestid.Get(c.(*gin.Context))
	}

	return logger.WithFields(logrus.Fields{
		"request_id": rid,
	})
}

// Will return the HTTP Client
func (cfg *Config) HttpClient() *http.Client {
	return &http.Client{
		Timeout: cfg.Client.Timeout.Read * time.Second,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout: cfg.Client.Timeout.Connect * time.Second,
			}).DialContext,
		},
	}
}

// Run will run the HTTP Server
func (cfg *Config) RunServer(handler *gin.Engine, loggerFunc repository.LoggerFunc) {
	log := loggerFunc(nil)

	// Set up a channel to listen to for interrupt signals
	runChan := make(chan os.Signal, 1)

	// Set up a context to allow for graceful server shutdowns in the event
	// of an OS interrupt (defers the cancel just in case)
	ctx, cancel := context.WithTimeout(
		context.Background(),
		cfg.Server.Timeout.Server * time.Second,
	)
	defer cancel()

	// Define server options
	server := &http.Server{
		Addr:           cfg.Server.Host + ":" + cfg.Server.Port,
		Handler:        handler,
		ReadTimeout:    cfg.Server.Timeout.Read * time.Second,
		WriteTimeout:   cfg.Server.Timeout.Write * time.Second,
		IdleTimeout:    cfg.Server.Timeout.Idle * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// Handle ctrl+c/ctrl+x interrupt and some other signals
	signal.Notify(runChan, os.Interrupt, os.Kill, syscall.SIGTSTP, syscall.SIGSTOP)

	// Alert the user that the server is starting
	log.Printf("Server is starting on %s", server.Addr)

	// Run the server on a new goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil {
			if err == http.ErrServerClosed {
				// Normal interrupt operation, ignore
			} else {
				log.Fatalf("Server failed to start due to err: %v", err)
			}
		}
	}()

	// Block on this channel listening for those previously defined syscalls assign
	// to variable so we can let the user know why the server is shutting down
	interrupt := <-runChan

	// If we get one of the pre-prescribed syscalls, gracefully terminate the server
	// while alerting the user
	log.Printf("Server is shutting down due to %+v", interrupt)
	// @todo: check shutdown here. it does not work after some time while runing server
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server was unable to gracefully shutdown due to err: %+v", err)
	}
}

// NewConfig returns a new decoded Config struct
func NewConfig(configPath *string) (*Config, error) {
	// Create config structure
	config := &Config{}

	// Open config file
	file, err := os.Open(*configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Init new YAML decode
	d := yaml.NewDecoder(file)

	// Start YAML decoding from file
	if err := d.Decode(&config); err != nil {
		return nil, fmt.Errorf("can not decode yaml config: %v", err)
	}

	return config, nil
}

// ValidateConfigPath just makes sure, that the path provided is a file,
// that can be read
func ValidateConfigPath(path *string) error {
	s, err := os.Stat(*path)

	if err != nil {
		return err
	}

	if s.IsDir() {
		return fmt.Errorf("'%s' is a directory, not a normal file", *path)
	}

	return nil
}

// ParseFlags will create and parse the CLI flags
// and return the path to be used elsewhere
func ParseFlags() (*string, error) {
	// Set up a CLI flag called "-config" to allow users
	// to supply the configuration file
	configPath := flag.String("config", "./config.yml", "path to config file")

	// Actually parse the flags
	flag.Parse()

	// Validate the path first
	if err := ValidateConfigPath(configPath); err != nil {
		return nil, err
	}

	// Return the configuration path
	return configPath, nil
}
