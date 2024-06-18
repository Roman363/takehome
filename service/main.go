package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/pelletier/go-toml"
)

// Default endpoint for querying the Simpsons quotes API
var defaultEndpoint = "http://thesimpsonsquoteapi.glitch.me/quotes"

const configFile = "config.toml"

// Config structure for storing the endpoint
type Config struct {
	Endpoint string `toml:"endpoint"`
}

// SimpsonsQuote structure to map the JSON response from the API
type SimpsonsQuote struct {
	Quote              string `json:"quote"`
	Character          string `json:"character"`
	Image              string `json:"image"`
	CharacterDirection string `json:"characterDirection"`
}

// loadConfig function loads the configuration from a config file located in the SNAP_DATA directory
func loadConfig() (Config, error) {
	config := Config{}
	// Get the SNAP_DATA environment variable
	snapDir := os.Getenv("SNAP_DATA")

	// If SNAP_DATA is not set, use the default endpoint
	if snapDir == "" {
		log.Printf("SNAP_DATA environment variable is not set, using default endpoint")
		config.Endpoint = defaultEndpoint
		return config, nil
	}

	// Construct the path to the config file
	configPath := filepath.Join(snapDir, configFile)

	// Check if the config file exists
	_, err := os.Stat(configPath)
	if os.IsNotExist(err) {
		log.Printf("config.toml not found in %s, using default endpoint", snapDir)
		config.Endpoint = defaultEndpoint
		return config, nil
	} else if err != nil {
		return config, err
	}

	// Read the config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return config, err
	}

	// Unmarshal the TOML data into the Config struct
	err = toml.Unmarshal(data, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}

// queryAPI function queries the given endpoint and prints a quote from the Simpsons
func queryAPI(endpoint string) {
	// Perform a GET request to the endpoint
	resp, err := http.Get(endpoint)
	if err != nil {
		log.Printf("error querying API: %v", err)
		return
	}
	if resp.StatusCode > 399 {
		log.Fatalf("unexpected response %v - %v", resp.StatusCode, resp.Status)
		return
	}
	defer resp.Body.Close()

	var q []SimpsonsQuote

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("error reading response body: %v", err)
		return
	}

	// Unmarshal the JSON data into a slice of SimpsonsQuote structs
	err = json.Unmarshal(body, &q)
	if err != nil {
		log.Printf("error reading quotes API - %v", err)
		return
	}

	// Print the first quote from the response
	quote := q[0]
	fmt.Printf("\"%s\" - %s\n", quote.Quote, quote.Character)
}

// main function is the entry point of the application
func main() {
	// Load the configuration
	config, err := loadConfig()
	if err != nil {
		log.Printf("error loading config: %v", err)
		log.Printf("using default endpoint: %v", defaultEndpoint)
		config.Endpoint = defaultEndpoint
	}

	// Create a ticker that triggers every 10 seconds
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	// Create a timer for 100 seconds
	timer := time.NewTimer(100 * time.Second)
	defer timer.Stop()

	// Query the API initially
	queryAPI(config.Endpoint)
	log.Printf("Printing out of loop\n") // Debug log

	// Loop to query the API every 10 seconds for 100 seconds
	for {
		select {
		case <-ticker.C:
			queryAPI(config.Endpoint)
			log.Printf("Printing in loop\n") // Debug log
		case <-timer.C:
			log.Printf("100 seconds have passed, stopping the loop")
			return
		}
	}
}
