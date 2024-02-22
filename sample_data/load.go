package main

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"math/rand"
	"time"
)

type Reading struct {
	ID          string
	Temperature float64
	DeviceID    string
	RecordedAt  time.Time
}

func main() {

	// Load .env file to as environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Define and parse flags
	devices := flag.Int("devices", 1, "How many mock devices would you like to create?")
	readingsPerDevice := flag.Int("readingsPerDevice", 10, "How many readings per device would you like to create?")
	flag.Parse()

	fmt.Println("Starting mock data generation.")
	fmt.Printf("Creating %d devices with %d readings each.\n", *devices, *readingsPerDevice)

	// Seed the random number generator
	source := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(source)

	// Define the range of temperatures within 0 and 100 for mock purposes
	minTemp := 0.0
	maxTemp := 100.0

	readings := []Reading{}

	i := 0
	j := 0

	// Iterate over each device and create mock readings
	// Hashing the device (i) will result in a consistent device ID
	// Hashing the reading (j) will result in a unique reading ID
	for i = 0; i < *devices; i++ {
		for j = 0; j < *readingsPerDevice; j++ {
			exampleReading := Reading{
				ID:          HashInt(j),
				Temperature: minTemp + rng.Float64()*(maxTemp-minTemp),
				DeviceID:    HashInt(i),
				RecordedAt:  time.Now(),
			}
			fmt.Println(toJSON(exampleReading))
			readings = append(readings, exampleReading)
		}
	}
}

// HashInt takes an integer and returns a hashed string
func HashInt(i int) string {
	hasher := sha256.New()
	idBuf := make([]byte, binary.MaxVarintLen64)
	binary.PutVarint(idBuf, int64(i))
	hasher.Write(idBuf)
	hashedIdBytes := hasher.Sum(nil)
	hashedId := fmt.Sprintf("%x", hashedIdBytes)
	return hashedId
}

// toJSON takes a struct and returns its JSON representation as a string.
// If marshaling fails, it logs the error and returns an empty string.
func toJSON(v interface{}) string {
	jsonData, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Printf("Error marshaling to JSON: %s", err)
		return ""
	}
	return string(jsonData)
}
