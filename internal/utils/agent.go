package utils

import (
	"fmt"
	"strings"
)

func NormaliseAgent(agent string) string {
	agentLower := strings.ToLower(agent)
	switch agentLower {
	default:
		return agentLower
	}
}

func MakeGuid(agent string, id string) string {
	return fmt.Sprintf("%s://%s", agent, id)
}
