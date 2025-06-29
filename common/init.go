// ./common/init.go
package common

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

var (
	Port                = flag.Int("port", 3000, "the listening port")
	SessionSecret       = uuid.New().String()
	CryptoSecret        = uuid.New().String()
	SQLitePath          = "my-site.db?_busy_timeout=5000"
	LogDir              = flag.String("log-dir", "./logs", "specify the log directory")
	DebugEnabled        bool
	MemoryCacheEnabled  bool
	SyncFrequency       int
	BatchUpdateInterval int
	BatchUpdateEnabled  = false
	RelayTimeout        int
)

func LoadEnv() {

	if os.Getenv("SESSION_SECRET") != "" {
		ss := os.Getenv("SESSION_SECRET")
		if ss == "random_string" {
			log.Println("WARNING: SESSION_SECRET is set to the default value 'random_string', please change it to a random string.")
			log.Fatal("Please set SESSION_SECRET to a random string.")
		} else {
			SessionSecret = ss
		}
	}
	if os.Getenv("CRYPTO_SECRET") != "" {
		CryptoSecret = os.Getenv("CRYPTO_SECRET")
	} else {
		CryptoSecret = SessionSecret
	}
	if os.Getenv("SQLITE_PATH") != "" {
		SQLitePath = os.Getenv("SQLITE_PATH")
	}
	if *LogDir != "" {
		var err error
		*LogDir, err = filepath.Abs(*LogDir)
		if err != nil {
			log.Fatal(err)
		}
		if _, err := os.Stat(*LogDir); os.IsNotExist(err) {
			err = os.Mkdir(*LogDir, 0777)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	// Initialize variables from constants.go that were using environment variables
	DebugEnabled = os.Getenv("DEBUG") == "true"
	MemoryCacheEnabled = os.Getenv("MEMORY_CACHE_ENABLED") == "true"

	// Initialize variables with GetEnvOrDefault
	SyncFrequency = GetEnvOrDefault("SYNC_FREQUENCY", 60)
	BatchUpdateInterval = GetEnvOrDefault("BATCH_UPDATE_INTERVAL", 5)
	RelayTimeout = GetEnvOrDefault("RELAY_TIMEOUT", 0)
}
