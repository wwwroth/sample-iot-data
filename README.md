### Load Sample Data

Generate and storing mock IoT (Internet of Things) sensor data, specifically temperature readings, into a MongoDB database. This script creates sample data from multiple devices over a specified number of days, allowing for the creation of a sample dataset for testing or demonstration purposes.

```shell
# Mock data for 5 devices going back 180 days
go run sample_data/load.go -devices=5 -days=180 
```

#### Response
```
2024/03/02 19:46:36 Generating 10 devices with 365 days worth of data
2024/03/02 19:46:38 Connecting to MongoDB with connection mongodb://localhost:27017
2024/03/02 19:46:38 Truncating records (dropping collection): sample_iot_data/readings
2024/03/02 19:46:38 Inserting records into MongoDB: sample_iot_data/readings
2024/03/02 19:46:38 Inserting readings batch [0:500000]
2024/03/02 19:46:40 Inserting readings batch [500000:1000000]
...
2024/03/02 19:46:56 Inserting readings batch [4500000:5000000]
2024/03/02 19:46:58 Inserting readings batch [5000000:5256000]
2024/03/02 19:47:00 Inserted 5256000 records successfully
```