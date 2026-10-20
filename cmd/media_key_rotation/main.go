package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"media-key-rotation/internal/rotation"
)

func main() {
	assetIngestions := flag.Int("asset-ingestions", 0, "active asset ingestion runs")
	processingJobs := flag.Int("processing-jobs", 0, "active media processing jobs")
	creatorDeliveries := flag.Int("creator-deliveries", 0, "active creator delivery runs")
	flag.Parse()

	client, err := rotation.NewClientFromEnv()
	if err != nil {
		log.Fatal(err)
	}
	result, err := rotation.RotateMediaKey(context.Background(), client, rotation.PipelineLoad{
		AssetIngestions: *assetIngestions, ProcessingJobs: *processingJobs, CreatorDeliveries: *creatorDeliveries,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result.Summary())
}
