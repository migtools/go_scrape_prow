package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/theckman/yacspin"
)

func start_spinner() (*yacspin.Spinner, error) {
	// meh have some fun
	cfg := yacspin.Config{
		Frequency:         500 * time.Millisecond,
		Writer:            nil,
		ShowCursor:        false,
		HideCursor:        false,
		SpinnerAtEnd:      false,
		CharSet:           yacspin.CharSets[59],
		Prefix:            " ",
		Suffix:            " ",
		SuffixAutoColon:   true,
		Message:           " Getting your jobs",
		ColorAll:          true,
		Colors:            []string{"fgYellow"},
		StopCharacter:     "✓",
		StopColors:        []string{"fgGreen"},
		StopMessage:       "done",
		StopFailCharacter: "✗",
		StopFailColors:    []string{"fgRed"},
		StopFailMessage:   "failed",
		NotTTY:            false,
	}

	spinner, err := yacspin.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to make spinner from struct: %s", err)
	}

	err = spinner.Start()
	time.Sleep(1 * time.Second)
	// end fun
	return spinner, err
}

func stopSpinnerOnSignal(spinner *yacspin.Spinner) {
	// ensure we stop the spinner before exiting, otherwise cursor will remain
	// hidden and terminal will require a `reset`
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh

		spinner.StopFailMessage("interrupted")

		// ignoring error intentionally
		_ = spinner.StopFail()

		os.Exit(0)
	}()
}
