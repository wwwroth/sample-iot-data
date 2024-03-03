package main

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Reading struct {
	ReadingID   string    `bson:"reading_id,omitempty"`
	Temperature float64   `bson:"temperature"`
	DeviceID    string    `bson:"device_id"`
	RecordedAt  time.Time `bson:"recorded_at"`
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}

	// Parse flags
	devices, daysOfData := parseFlags()

	log.Printf("Generating %d devices with %d days worth of data", devices, daysOfData)

	// Generate readings
	readings := generateReadings(devices, daysOfData)

	mongoUrl := os.Getenv("MONGO_URI")
	mongoDb := os.Getenv("MONGO_DB")
	mongoCollection := os.Getenv("MONGO_COLLECTION")

	// Connect to MongoDB
	client, err := connectToMongoDB(mongoUrl)
	if err != nil {
		return err
	}
	defer client.Disconnect(context.TODO())

	// Insert readings into MongoDB
	if err := insertReadingsIntoMongoDB(client, mongoDb, mongoCollection, readings); err != nil {
		return err
	}

	return nil
}

func parseFlags() (int, int) {
	devices := flag.Int("devices", 1, "Number of mock devices to create.")
	daysOfData := flag.Int("days", 365, "Number of days of data per device to mock.")
	flag.Parse()
	return *devices, *daysOfData
}

func generateReadings(devices, daysOfData int) []Reading {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	minTemp, maxTemp := 0.0, 100.0
	var readings []Reading

	startTime := time.Now().AddDate(0, 0, -daysOfData)
	totalMinutes := daysOfData * 1440

	for i := 0; i < totalMinutes; i++ {
		for j := 1; j <= devices; j++ {
			timestamp := startTime.Add(time.Duration(i) * time.Minute)
			readings = append(readings, Reading{
				ReadingID:   hashInt(i + j),
				Temperature: minTemp + rng.Float64()*(maxTemp-minTemp),
				DeviceID:    hashInt(j),
				RecordedAt:  timestamp,
			})
		}
	}

	return readings
}

func connectToMongoDB(mongoUrl string) (*mongo.Client, error) {
	log.Printf("Connecting to MongoDB with connection %s", mongoUrl)
	clientOptions := options.Client().ApplyURI(mongoUrl)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}
	if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}
	return client, nil
}

func insertReadingsIntoMongoDB(client *mongo.Client, mongoDb string, mongoCollection string, readings []Reading) error {

	log.Printf("Truncating records (dropping collection): %s/%s", mongoDb, mongoCollection)

	collection := client.Database(mongoDb).Collection(mongoCollection)
	if err := dropCollection(collection); err != nil {
		return err
	}

	log.Printf("Inserting records into MongoDB: %s/%s", mongoDb, mongoCollection)

	batchSize := 500000
	for i := 0; i < len(readings); i += batchSize {
		end := i + batchSize
		if end > len(readings) {
			end = len(readings)
		}
		log.Printf("Inserting readings batch [%d:%d]", i, end)
		if err := insertReadings(collection, readings[i:end]); err != nil {
			return fmt.Errorf("failed to insert batch %d-%d: %w", i, end, err)
		}
	}

	count, err := countReadings(collection)
	if err != nil {
		return err
	}

	log.Printf("Inserted %d records successfully", count)

	return nil
}

func dropCollection(collection *mongo.Collection) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := collection.Drop(ctx); err != nil {
		return fmt.Errorf("failed to drop collection: %w", err)
	}
	return nil
}

func insertReadings(collection *mongo.Collection, readings []Reading) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var documents []interface{}
	for _, reading := range readings {
		documents = append(documents, reading)
	}
	if _, err := collection.InsertMany(ctx, documents); err != nil {
		return fmt.Errorf("failed to insert readings: %w", err)
	}
	return nil
}

func countReadings(collection *mongo.Collection) (int64, error) {
	count, err := collection.CountDocuments(context.Background(), map[string]interface{}{})
	if err != nil {
		return 0, fmt.Errorf("failed to count readings: %w", err)
	}
	return count, nil
}

// hashInt takes an integer and returns a hashed string
func hashInt(i int) string {
	hasher := sha256.New()
	idBuf := make([]byte, binary.MaxVarintLen64)
	binary.PutVarint(idBuf, int64(i))
	hasher.Write(idBuf)
	hashedIdBytes := hasher.Sum(nil)
	return fmt.Sprintf("%x", hashedIdBytes)
}
