package cmd

import (
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"
)

// Allowed values come from the Fizzy API documentation
var cardListStatusValues = stringSet(
	"all",
	"closed",
	"not_now",
	"stalled",
	"postponing_soon",
	"golden",
)

// Allowed values come from the Fizzy API documentation
var cardListSortValues = stringSet("latest", "newest", "oldest")

// Allowed values come from the Fizzy API documentation
var cardListDateValues = stringSet(
	"today",
	"yesterday",
	"thisweek",
	"lastweek",
	"thismonth",
	"lastmonth",
	"thisyear",
	"lastyear",
)

// Allowed values come from the Fizzy API documentation
var webhookEventValues = stringSet(
	"card_assigned",
	"card_closed",
	"card_postponed",
	"card_auto_postponed",
	"card_board_changed",
	"card_published",
	"card_reopened",
	"card_sent_back_to_triage",
	"card_triaged",
	"card_unassigned",
	"comment_created",
)

// The Fizzy API documentation limits reaction text to 16 characters.
const maxReactionContentRunes = 16

func stringSet(values ...string) map[string]struct{} {
	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		set[value] = struct{}{}
	}

	return set
}

func validateAllowedValue(flagName, value string, allowed map[string]struct{}) error {
	if _, ok := allowed[value]; ok {
		return nil
	}

	return fmt.Errorf("%s must be one of %s", flagName, strings.Join(sortedAllowedValues(allowed), ", "))
}

func sortedAllowedValues(allowed map[string]struct{}) []string {
	values := make([]string, 0, len(allowed))
	for value := range allowed {
		values = append(values, value)
	}

	sort.Strings(values)

	return values
}

func validateCardListFilters() error {
	if cardListStatus != "" {
		if err := validateAllowedValue("--status", cardListStatus, cardListStatusValues); err != nil {
			return err
		}
	}

	if cardListSort != "" {
		if err := validateAllowedValue("--sort", cardListSort, cardListSortValues); err != nil {
			return err
		}
	}

	if cardListCreated != "" {
		if err := validateAllowedValue("--created", cardListCreated, cardListDateValues); err != nil {
			return err
		}
	}

	if cardListClosed != "" {
		if err := validateAllowedValue("--closed", cardListClosed, cardListDateValues); err != nil {
			return err
		}
	}

	return nil
}

func validateReactionContent(flagName, value string) error {
	if utf8.RuneCountInString(value) > maxReactionContentRunes {
		return fmt.Errorf("%s must be at most %d characters", flagName, maxReactionContentRunes)
	}

	return nil
}

func parseWebhookEvents(raw string) ([]string, error) {
	if raw == "" {
		return nil, nil
	}

	parts := strings.Split(raw, ",")

	events := make([]string, 0, len(parts))
	for _, part := range parts {
		event := strings.TrimSpace(part)
		if event == "" {
			return nil, fmt.Errorf("--events must not contain empty values")
		}

		if err := validateAllowedValue("--events", event, webhookEventValues); err != nil {
			return nil, err
		}

		events = append(events, event)
	}

	return events, nil
}
