package integration_test

import (
	"reflect"
	"regexp"
	"testing"

	"github.com/cjsaylor/chessbot/integration"
)

func TestParse(t *testing.T) {
	commandPatterns := []integration.CommandPattern{
		integration.CommandPattern{
			Type:    integration.Challenge,
			Pattern: regexp.MustCompile("^<@([\\w|\\d]+)>.*challenge <@([\\w\\d]+)>.*$"),
		},
		integration.CommandPattern{
			Type:    integration.Help,
			Pattern: regexp.MustCompile(".*help.*"),
		},
	}
	parser := integration.NewCommandParser(commandPatterns)

	for _, input := range []struct {
		text            string
		expectedCommand integration.CommandType
		expectedParams  []string
	}{
		{
			text:            "<@U29109> challenge <@U391099>",
			expectedCommand: integration.Challenge,
			expectedParams:  []string{"U29109", "U391099"},
		},
		{
			text:            "help",
			expectedCommand: integration.Help,
			expectedParams:  []string{},
		},
		{
			text:            "don't know what this is",
			expectedCommand: integration.Unknown,
			expectedParams:  []string{},
		},
	} {
		match := parser.ParseInput(input.text)
		if match.Type != input.expectedCommand {
			t.Errorf("Expected command type of %v, got %v", input.expectedCommand, match.Type)
		}
		if !reflect.DeepEqual(match.Params, input.expectedParams) {
			t.Errorf("Expected parsed command parameters %v, got %v", input.expectedParams, match.Params)
		}
	}

}
