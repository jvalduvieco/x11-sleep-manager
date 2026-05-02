package matcher

import (
	"strings"

	"github.com/jvalduvieco/x11_sleep_manager/internal/config"
	"github.com/jvalduvieco/x11_sleep_manager/internal/observe"
)

func Filter(inhibitors []observe.Inhibitor, match config.MatchConfig, selfUID int) []observe.Inhibitor {
	matched := make([]observe.Inhibitor, 0, len(inhibitors))
	for _, inhibitor := range inhibitors {
		if Matches(inhibitor, match, selfUID) {
			matched = append(matched, inhibitor)
		}
	}
	return matched
}

func Matches(inhibitor observe.Inhibitor, match config.MatchConfig, selfUID int) bool {
	if match.UID == "self" && inhibitor.UID != selfUID {
		return false
	}
	if len(match.Who) > 0 && !containsExact(match.Who, inhibitor.Who) {
		return false
	}
	if len(match.WhatAny) > 0 && !containsAnyToken(match.WhatAny, inhibitor.What) {
		return false
	}
	return true
}

func containsExact(candidates []string, value string) bool {
	for _, candidate := range candidates {
		if candidate == value {
			return true
		}
	}
	return false
}

func containsAnyToken(candidates []string, what string) bool {
	tokens := strings.Split(what, ":")
	for _, token := range tokens {
		for _, candidate := range candidates {
			if token == candidate {
				return true
			}
		}
	}
	return false
}
