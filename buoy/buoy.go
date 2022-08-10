package buoy

import (
	"fmt"
	"time"
)

type (
	TBuoyParams struct {
		// Minimum BuoyFunc run time in seconds to be considered a successful run
		// If zero, any run time exit will trigger failure counter.
		MinimumSuccessTime uint32
		// Timeout in seconds before restarting the BuoyFunc. Zero means no delay.
		RestartDelay uint32
		// This many failures in a row will case buoy loop to terminate.
		// Zero means it will terminate on the first failure
		GeneralFailureCount uint
		// Debug callback
		DebugCallback func(message string)
	}

	TBuoyFunc func()
)

func (BPar TBuoyParams) DebugSay(message string) {
	if BPar.DebugCallback != nil {
		BPar.DebugCallback(message)
	}
}

func (BPar TBuoyParams) KeepFloating(bfun TBuoyFunc) {
	var failCount uint
	for {
		BPar.DebugSay("Function starts")
		timeStart := time.Now()
		bfun()
		elapsed := time.Since(timeStart)
		BPar.DebugSay(fmt.Sprintf("Function exited: %+v elapsed", elapsed))
		if (uint32(elapsed.Seconds()) < BPar.MinimumSuccessTime) || (BPar.MinimumSuccessTime == 0) {
			failCount++
			BPar.DebugSay(fmt.Sprintf("Failure counter: %d", failCount))
			if failCount >= BPar.GeneralFailureCount {
				BPar.DebugSay("General failure, exiting the loop")
				break
			} else {
				BPar.DebugSay(fmt.Sprintf("Delaying restart by %d seconds", BPar.RestartDelay))
				time.Sleep(time.Duration(BPar.RestartDelay) * time.Second)
			}
		} else {
			if failCount > 0 {
				BPar.DebugSay("Failure counter reset")
			}
			failCount = 0
		}
	}
}
