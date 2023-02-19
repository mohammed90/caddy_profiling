package caddy_profiling

type Parameters struct {
	CPUProfileRate       int           `json:"cpu_profile_rate,omitempty"`
	BlockProfileRate     int           `json:"block_profile_rate,omitempty"`
	MutexProfileFraction int           `json:"mutex_profile_fraction,omitempty"`
	ProfileTypes         []ProfileType `json:"profile_types,omitempty"`
}

type ProfileType string

const (
	Goroutine    ProfileType = "goroutine"
	Heap         ProfileType = "heap"
	Allocs       ProfileType = "allocs"
	Threadcreate ProfileType = "threadcreate"
	Block        ProfileType = "block"
	Mutex        ProfileType = "mutex"
)

type ProfilingParameterSetter interface {
	SetProfilingParameter(Parameters)
}

type Profiler interface {
	Start() error
	Stop() error
}
