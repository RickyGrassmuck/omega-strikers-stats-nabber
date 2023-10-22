package tracker

import (
	"reflect"
	"time"
)

var Teams = map[string]string{
	"TeamOne": "Blue Team",
	"TeamTwo": "Red Team",
}

type LogLine struct {
	Timestamp      string
	Module         string
	RawMessage     string
	MsgType        MsgTypeName
	ParsedDataType reflect.Type
	ParsedData     interface{}
}

type SetResult struct {
	TeamOneScore int
	TeamTwoScore int
	Winner       string
}

type MatchResults struct {
	Winner string
	Sets   map[int]SetResult
}

type MatchState struct {
	Phase GamePhase
	Teams struct {
		TeamOne []Player
		TeamTwo []Player
	}
	Start      time.Time
	End        time.Time
	CurrentSet int
	Results    MatchResults
}

type Player struct {
	Name  string
	Stats struct {
		Goals   int
		Assists int
		Saves   int
		Shots   int
		KOs     int
	}
}
