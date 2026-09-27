package domain

import "testing"

func TestBootstrapConfigProposalFieldsAcceptsGeneratedSecretStrategies(t *testing.T) {
	for _, strategy := range []string{"generated", "generated credentials", "generated database credentials"} {
		t.Run(strategy, func(t *testing.T) {
			fields := BootstrapConfigProposalFields{SecretStrategies: map[string]string{"Secret/app/backend-secret": strategy}}
			if err := fields.Validate(); err != nil {
				t.Fatalf("strategy %q rejected: %v", strategy, err)
			}
		})
	}
}

func TestBootstrapConfigProposalFieldsRejectsUnknownSecretStrategy(t *testing.T) {
	fields := BootstrapConfigProposalFields{SecretStrategies: map[string]string{"Secret/app/backend-secret": "copy arbitrary source"}}
	if err := fields.Validate(); err == nil {
		t.Fatal("unknown secret strategy was accepted")
	}
}
