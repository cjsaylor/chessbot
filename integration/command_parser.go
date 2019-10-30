package integration

import (
	"errors"
	"regexp"
)

// CommandType is the kind of command a user wishes to execute.
type CommandType uint8

const (
	// Unknown represents a command that isn't represented in the enum.
	Unknown CommandType = iota + 1
	// Challenge represents a challenge command for initiating a chess match.
	Challenge
	// Move represents a specific move by a player.
	Move
	// Resign represents a player's intention of resignation.
	Resign
	// Takeback represents a player's request to take back a previous move.
	Takeback
	// Help represents a player's need for help (UI or otherwise).
	Help
)

// CommandPattern maps a regular expression pattern to a specific command type.
type CommandPattern struct {
	Type    CommandType
	Pattern *regexp.Regexp
}

// CommandMatch represents a matched command pattern with parsed parameters.
type CommandMatch struct {
	Type           CommandType
	MatchedPattern *CommandPattern
	Params         []string
}

// ChallengeCommand represents a challenge to propose
type ChallengeCommand struct {
	ChallengedID string
}

// MoveCommand represents a single long algebraic notation move.
type MoveCommand struct {
	LAN string
}

// ToChallenge converts this command match to a proper challenge command
func (c *CommandMatch) ToChallenge() (*ChallengeCommand, error) {
	if c.Type != Challenge || len(c.Params) < 1 {
		return nil, errors.New("match is not a valid challenge command")
	}
	return &ChallengeCommand{
		ChallengedID: c.Params[0],
	}, nil
}

// ToMove converts this command match to a proper move command
func (c *CommandMatch) ToMove() (*MoveCommand, error) {
	if c.Type != Move || len(c.Params) < 1 {
		return nil, errors.New("match is not a valid move command")
	}
	return &MoveCommand{
		LAN: c.Params[0],
	}, nil
}

// CommandParser will parse and return a CommandMatch.
type CommandParser struct {
	patterns []CommandPattern
}

// NewCommandParser gets a new instance operating on provided CommandMap.
func NewCommandParser(patterns []CommandPattern) CommandParser {
	return CommandParser{patterns: patterns}
}

// ParseInput will attempt to match a command.
// An unknown command will still match with a type of Unknown
func (c *CommandParser) ParseInput(input string) CommandMatch {
	match := CommandMatch{
		Type:   Unknown,
		Params: []string{},
	}
	for _, pattern := range c.patterns {
		results := pattern.Pattern.FindStringSubmatch(input)
		if len(results) > 0 {
			return CommandMatch{
				Type:           pattern.Type,
				MatchedPattern: &pattern,
				Params:         results[1:],
			}
		}
	}
	return match
}
