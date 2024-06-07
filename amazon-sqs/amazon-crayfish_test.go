package main

import (
	"context"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

func TestHandler(t *testing.T) {
	// Mock SQS event
	mockEvent := events.SQSEvent{
		Records: []events.SQSMessage{
			{ // Example of a single message content (a sub-population) in the SQS queue
				Body: `{"SubPopulation":[[0.1, 0.2, 3.2, 1.5]], "T":100, "F":"F16", "StartTime":"2020-01-01T01:01:01Z"}`,
			},
		},
	}

	ctx := context.Background()

	// Call the Lambda handler
	err := Handler(ctx, mockEvent)
	if err != nil {
		t.Fatalf("Handler returned an unexpected error: %v", err)
	}
}
