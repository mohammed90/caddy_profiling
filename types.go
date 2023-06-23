package caddy_profiling

// Common profiling paramters
type Parameters struct {
	// The hertz rate for CPU profiling, as accepted by the [`SetCPUProfileRate`](https://pkg.go.dev/runtime#SetCPUProfileRate) function.
	CPUProfileRate int `json:"cpu_profile_rate,omitempty"`

	// The hertz rate for CPU profiling, as accepted by the [`SetBlockProfileRate`](https://pkg.go.dev/runtime#SetBlockProfileRate) function.
	BlockProfileRate int `json:"block_profile_rate,omitempty"`

	// The the fraction of mutex contention events, as accepted by the [`SetMutexProfileFraction`](https://pkg.go.dev/runtime#SetMutexProfileFraction) function.
	MutexProfileFraction int `json:"mutex_profile_fraction,omitempty"`

	// The enabled runtime profile types. The accepted values are: cpu, goroutine, heap, allocs, threadcreate, block, mutex.
	ProfileTypes []ProfileType `json:"profile_types,omitempty"`
}

type ProfileType string

const (
	CPU          ProfileType = "cpu"
	Goroutine    ProfileType = "goroutine"
	Heap         ProfileType = "heap"
	Allocs       ProfileType = "allocs"
	Threadcreate ProfileType = "threadcreate"
	Block        ProfileType = "block"
	Mutex        ProfileType = "mutex"
)

// Signals whether a guest profiling module accepts inheriting the profiling parameters
type ProfilingParameterSetter interface {
	SetProfilingParameter(Parameters)
}

// Guest moduels of the `profiling` app are expected to implement this interface
// and be registered in the `profiling.profiler` caddy namespace
type Profiler interface {
	Start() error
	Stop() error
}
