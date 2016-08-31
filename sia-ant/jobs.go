package main

import (
	"github.com/NebulousLabs/Sia/api"
	"github.com/NebulousLabs/Sia/sync"
)

// A JobRunner is used to start up jobs on the running Sia node.
type JobRunner struct {
	client         *api.Client
	walletPassword string
	siaDirectory   string
	tg             sync.ThreadGroup
}

// NewJobRunner creates a new job runner, using the provided api address,
// authentication password, and sia directory.  It expects the connected api to
// be newly initialized, and initializes a new wallet, for usage in the jobs.
// siadirectory is used in logging to identify the job runner.
func NewJobRunner(apiaddr string, authpassword string, siadirectory string) (*JobRunner, error) {
	jr := &JobRunner{
		client:       api.NewClient(apiaddr, authpassword),
		siaDirectory: siadirectory,
	}
	var walletParams api.WalletInitPOST
	err := jr.client.Post("/wallet/init", "", &walletParams)
	if err != nil {
		return nil, err
	}
	jr.walletPassword = walletParams.PrimarySeed
	return jr, nil
}

// Stop signals all running jobs to stop and blocks until the jobs have
// finished stopping.
func (j *JobRunner) Stop() {
	j.tg.Stop()
}
