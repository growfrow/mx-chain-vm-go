package vmjsonintegrationtest

import (
	"testing"
)

func TestDistribution_v0_1(t *testing.T) {
	if testing.Short() {
		t.Skip("not a short test")
	}

	runAllTestsInFolder(t, "distribution/v0_1")
}

func TestDistribution_v0_1_single(t *testing.T) {
	if testing.Short() {
		t.Skip("not a short test")
	}

	runSingleTest(t, "distribution/v0_1/mandos", "claim_rewards_proxy_after_enter_with_lock_after_mint_rewards.scen.json")
}
