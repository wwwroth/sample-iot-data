package main

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
	"math/rand"
	"sync"
	"time"
)

type Reading struct {
	ReadingID   string    `bson:"reading_id,omitempty"`
	Temperature float64   `bson:"temperature"`
	DeviceID    string    `bson:"device_id"`
	RecordedAt  time.Time `bson:"recorded_at"`
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

	log.Println("Starting mock data generation.")
	log.Printf("Creating %d devices with %d readings each - %d total \n", *devices, *readingsPerDevice, *devices**readingsPerDevice)

	// Seed the random number generator
	source := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(source)

	// Define the range of temperatures within 0 and 100 for mock purposes
	minTemp := 0.0
	maxTemp := 100.0

	readings := []Reading{}

	i := 1
	j := 1

	// Iterate over each device and create mock readings
	// Hashing the device (i) will result in a consistent device ID
	// Hashing the reading (j) will result in a unique reading ID
	for i = 0; i < *devices; i++ {
		for j = 0; j < *readingsPerDevice; j++ {
			exampleReading := Reading{
				ReadingID:   HashInt(j),
				Temperature: minTemp + rng.Float64()*(maxTemp-minTemp),
				DeviceID:    HashInt(i),
				RecordedAt:  time.Now(),
			}
			readings = append(readings, exampleReading)
		}
	}

	// Set client options
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")

	// Connect to MongoDB
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// Check the connection
	err = client.Ping(context.TODO(), readpref.Primary())
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Truncating existing data")
	err = truncateCollection(client.Database("sample_iot_data").Collection("readings"))
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Inserting new data")

	batchSize := 500
	numGoroutines := len(readings) / batchSize

	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			start := i * batchSize
			end := start + batchSize
			if end > len(readings) {
				end = len(readings)
			}

			err = insertReadings(readings[start:end], client.Database("sample_iot_data").Collection("readings"))
			if err != nil {
				log.Printf("Failed to insert batch %d: %v", i, err)
			}
		}(i)
	}

	wg.Wait()

	log.Println("All data inserted.")
	count := countReadings(client.Database("sample_iot_data").Collection("readings"))
	log.Printf("Inserted %d records.", count)

	// Disconnect from MongoDB
	defer client.Disconnect(context.TODO())
}

func truncateCollection(collection *mongo.Collection) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// An empty filter matches all documents in the collection
	filter := bson.D{{}}

	result, err := collection.DeleteMany(ctx, filter)
	if err != nil {
		return err
	}

	log.Printf("Deleted %v documents\n", result.DeletedCount)
	return nil
}

func insertReadings(readings []Reading, collection *mongo.Collection) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var documents []interface{}
	for _, reading := range readings {
		documents = append(documents, reading)
	}

	_, err := collection.InsertMany(ctx, documents)
	if err != nil {
		return err
	}

	return nil
}

func countReadings(collection *mongo.Collection) int64 {
	count, err := collection.CountDocuments(context.Background(), map[string]interface{}{})
	if err != nil {
		log.Fatal(err)
	}

	return count
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
