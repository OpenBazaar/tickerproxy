package events

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppSyncResolverTemplate_invoke(t *testing.T) {
	inputJSON, err := ioutil.ReadFile("./testdata/appsync-invoke.json")
	if err != nil {
		t.Errorf("could not open test file. details: %v", err)
	}

	var inputEvent AppSyncResolverTemplate
	if err = json.Unmarshal(inputJSON, &inputEvent); err != nil {
		t.Errorf("could not unmarshal event. details: %v", err)
	}
	assert.Equal(t, OperationInvoke, inputEvent.Operation)

	outputJSON, err := json.Marshal(inputEvent)
	if err != nil {
		t.Errorf("could not marshal event. details: %v", err)
	}

	assert.JSONEq(t, string(inputJSON), string(outputJSON))
}

func TestAppSyncResolverTemplate_batchinvoke(t *testing.T) {
	inputJSON, err := ioutil.ReadFile("./testdata/appsync-batchinvoke.json")
	if err != nil {
		t.Errorf("could not open test file. details: %v", err)
	}

	var inputEvent AppSyncResolverTemplate
	if err = json.Unmarshal(inputJSON, &inputEvent); err != nil {
		t.Errorf("could not unmarshal event. details: %v", err)
	}
	assert.Equal(t, OperationBatchInvoke, inputEvent.Operation)

	outputJSON, err := json.Marshal(inputEvent)
	if err != nil {
		t.Errorf("could not marshal event. details: %v", err)
	}

	assert.JSONEq(t, string(inputJSON), string(outputJSON))
}
