package domain

import "testing"

func TestSubscriptionTransitions(t *testing.T) {
	for _, test := range []struct {
		from, to SubscriptionState
		valid    bool
	}{
		{"", SubscriptionActive, true},
		{SubscriptionActive, SubscriptionPastDue, true},
		{SubscriptionGrace, SubscriptionActive, true},
		{SubscriptionCanceled, SubscriptionActive, false},
		{SubscriptionActive, SubscriptionTrialing, false},
		{SubscriptionState("unknown"), SubscriptionState("unknown"), false},
		{"", SubscriptionState("unknown"), false},
	} {
		err := ValidateSubscriptionTransition(test.from, test.to)
		if (err == nil) != test.valid {
			t.Fatalf("transition %q -> %q valid=%v err=%v", test.from, test.to, test.valid, err)
		}
	}
}
