package XRay

import (
	"math"
	"time"

	runtimeDebug "runtime/debug"
)

func forceFree(interval time.Duration) {
	go func() {
		for {
			time.Sleep(interval)
			runtimeDebug.FreeOSMemory()
		}
	}()
}

func InitForceFree(maxMemory int64, interval int) {
	runtimeDebug.SetGCPercent(10)
	runtimeDebug.SetMemoryLimit(maxMemory)
	if interval > 0 {
		duration := time.Duration(interval) * time.Second
		forceFree(duration)
	}
}
