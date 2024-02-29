### Load Sample Data

```shell
# Mock data for 5 devices going back 180 days
go run sample_data/load.go -devices=5 -days=180 
```

#### Response
```
2024/02/29 00:23:38 Starting mock data generation.
2024/02/29 00:23:38 Generating 10 devices with 365 days worth of data.
2024/02/29 00:23:40 Dropping existing collection
2024/02/29 00:23:40 Dropping collection complete
2024/02/29 00:23:40 Inserting new data
2024/02/29 00:23:40 Inserting batch 0:500000
2024/02/29 00:23:42 Inserting batch 500000:1000000
2024/02/29 00:23:44 Inserting batch 1000000:1500000
...
2024/02/29 00:23:57 Inserting batch 4500000:5000000
2024/02/29 00:23:59 Inserting batch 5000000:5256000
2024/02/29 00:24:00 All data inserted.
2024/02/29 00:24:02 Inserted 5256000 records.
```