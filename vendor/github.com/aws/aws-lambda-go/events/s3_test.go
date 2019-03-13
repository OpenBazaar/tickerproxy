// Copyright 2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
package events

import (
	"encoding/json"
	"testing"

	"github.com/aws/aws-lambda-go/events/test"
	"github.com/stretchr/testify/assert"
)

func TestS3EventMarshaling(t *testing.T) {

	// 1. read JSON from file
	inputJSON := readJsonFromFile(t, "./testdata/s3-event.json")

	// 2. de-serialize into Go object
	var inputEvent S3Event
	if err := json.Unmarshal(inputJSON, &inputEvent); err != nil {
		t.Errorf("could not unmarshal event. details: %v", err)
	}

	// 3. serialize to JSON
	outputJSON, err := json.Marshal(inputEvent)
	if err != nil {
		t.Errorf("could not marshal event. details: %v", err)
	}

	// 4. check result
	assert.JSONEq(t, string(inputJSON), string(outputJSON))
}

func TestS3TestEventMarshaling(t *testing.T) {
	inputJSON := []byte(`{
	    "Service" :"Amazon S3",
	    "Event": "s3:TestEvent",
	    "Time": "2019-02-04T19:34:46.985Z",
	    "Bucket": "bmoffatt",
	    "RequestId": "7BA1940DC6AF888B",
	    "HostId": "q1YDbiaMjllP0m+Lcy6cKKgxNrMLFJ9zCrZUFBqHGTG++0nXvnTDIGC7q2/QPAsJg86E8gI7y9U="
	}`)
	var inputEvent S3TestEvent
	if err := json.Unmarshal(inputJSON, &inputEvent); err != nil {
		t.Errorf("could not marshal event. details: %v", err)
	}
	outputJSON, err := json.Marshal(inputEvent)
	if err != nil {
		t.Errorf("could not marshal event. details: %v", err)
	}
	assert.JSONEq(t, string(inputJSON), string(outputJSON))
}

func TestS3MarshalingMalformedJSON(t *testing.T) {
	test.TestMalformedJson(t, S3Event{})
}
