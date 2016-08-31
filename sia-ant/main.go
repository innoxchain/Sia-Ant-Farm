package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"time"

	"github.com/NebulousLabs/Sia/api"
)

// NewSiad spawns a new siad process using os/exec and waits for the api to
// become available.  siadPath is the path to Siad, passed directly to
// exec.Command.  An error is returned if starting siad fails, otherwise a
// pointer to siad's os.Cmd object is returned.  The data directory `datadir`
// is passed as siad's `--sia-directory`.
func NewSiad(siadPath string, datadir string, apiAddr string, rpcAddr string, hostAddr string) (*exec.Cmd, error) {
	cmd := exec.Command(siadPath, "--no-bootstrap", "--sia-directory", datadir, "--api-addr", apiAddr, "--rpc-addr", rpcAddr, "--host-addr", hostAddr)
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	// After starting siad, we must immediately listen for Interrupt signals sent
	// to the main process to avoid leaving orphaned siad processes when this
	// program is interrupted.
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, os.Interrupt)
	go func() {
		<-sigchan
		if err := api.NewClient(apiAddr, "").Get("/daemon/stop", nil); err != nil {
			// Call to /daemon/stop failed, maybe the API hasnt finished loading yet.
			// Fall back to issuing an interrupt signal.
			cmd.Process.Signal(os.Interrupt)
		}
	}()

	if err := waitForAPI(apiAddr); err != nil {
		return nil, err
	}

	return cmd, nil
}

// waitForAPI blocks until the Sia API at apiAddr becomes available.
func waitForAPI(apiAddr string) error {
	c := api.NewClient(apiAddr, "")

	// Wait for the Sia API to become available.
	success := false
	for start := time.Now(); time.Since(start) < 5*time.Minute; time.Sleep(time.Millisecond * 100) {
		if err := c.Get("/consensus", nil); err == nil {
			success = true
			break
		}
	}
	if !success {
		c.Get("/daemon/stop", nil)
		return errors.New("timeout: couldnt reach api after 5 minutes")
	}
	return nil
}

// runSiaAnt is the main entry point of the sia-ant program, and returns an
// exit code.
func runSiaAnt(siadPath, apiAddr, rpcAddr, hostAddr, siaDirectory string, runGateway bool, runMining bool) int {
	// Construct a new siad instance
	siad, err := NewSiad(siadPath, siaDirectory, apiAddr, rpcAddr, hostAddr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error starting siad: %v\n", err)
		return 1
	}

	defer func() {
		api.NewClient(apiAddr, "").Get("/daemon/stop", nil)
	}()

	// Construct the job runner
	j, err := NewJobRunner(apiAddr, "", siaDirectory)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating job runner: %v\n", err)
		return 1
	}

	// Catch os.Interrupt and trigger a clean close of sia-ant.  First, stop all
	// jobs using JobRunner's ThreadGroup, then stop siad using the API.
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, os.Interrupt)
	go func() {
		<-sigchan
		j.Stop()
		if err := api.NewClient(apiAddr, "").Get("/daemon/stop", nil); err != nil {
			// Call to /daemon/stop failed, maybe the API hasnt finished loading yet.
			// Fall back to issuing an interrupt signal.
			siad.Process.Signal(os.Interrupt)
		}
	}()

	// Start up selected jobs
	if runGateway {
		go j.gatewayConnectability()
	}
	if runMining {
		go j.blockMining()
	}

	siad.Wait()
	return 0
}

func main() {
	siadPath := flag.String("siad", "siad", "path to siad executable")
	apiAddr := flag.String("api-addr", "localhost:9980", "api address to bind siad")
	rpcAddr := flag.String("rpc-addr", "localhost:9981", "rpc address to bind siad")
	hostAddr := flag.String("host-addr", "localhost:9982", "host address to bind siad")
	siaDirectory := flag.String("sia-directory", "./", "sia data directory")
	runGateway := flag.Bool("gateway", false, "enable gateway test jobs")
	runMining := flag.Bool("mining", false, "enable mining test jobs")
	flag.Parse()

	os.Exit(runSiaAnt(*siadPath, *apiAddr, *rpcAddr, *hostAddr, *siaDirectory, *runGateway, *runMining))
}
