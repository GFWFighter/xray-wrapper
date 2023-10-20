package XRay

import (
	"math"

	runtimeDebug "runtime/debug"
)

func SetMemoryLimit(enabled bool) {
	const memoryLimit = 45 * 1024 * 1024
	const memoryLimitGo = memoryLimit / 1.5
	if enabled {
		runtimeDebug.SetGCPercent(10)
		runtimeDebug.SetMemoryLimit(memoryLimitGo)
	} else {
		runtimeDebug.SetGCPercent(100)
		runtimeDebug.SetMemoryLimit(math.MaxInt64)
	}
}
