package pyroscope

import (
	"runtime"

	"github.com/caddyserver/caddy/v2"
	"github.com/mohammed90/caddy_profiling"
	"github.com/pyroscope-io/client/pyroscope"
)

func init() {
	caddy.RegisterModule(new(App))
	caddy.RegisterModule(new(ProfilingApp))
}

// The `pyroscope` app collects profiling data during the life-time of the process
// and uploads them to the Pyroscope server.
type App struct {
	// The application name reported to Pyroscope. The config value may be a [placeholder](https://caddyserver.com/docs/conventions#placeholders).
	ApplicationName string `json:"application_name,omitempty"`

	// TODO: decide no the inclusion of tags and whether they're beneficial
	// Custom tags to be included. The config value may be a [placeholder](https://caddyserver.com/docs/conventions#placeholders).
	// Tags            map[string]string `json:"tags,omitempty"`

	// The URL of the Pyroscope service. The config value may be a [placeholder](https://caddyserver.com/docs/conventions#placeholders).
	ServerAddress string `json:"server_address,omitempty"`

	// The token for Pyroscope Cloud. The config value may be a [placeholder](https://caddyserver.com/docs/conventions#placeholders).
	AuthToken string `json:"auth_token,omitempty"`

	// Disable automatic runtime.GC runs between getting the heap profiles
	DisableGCRuns bool `json:"disable_gc_runs,omitempty"`

	// The profiling parameters to be reported to Pyroscope.
	// The paramters cpu_profile_rate, block_profile_rate, and mutex_profile_fraction are inherited from the `profiling` app if `pyroscope`
	// is configured as a child module. The `profile_types` field is inherited if not configured explicitly.
	// If `pyroscope` is configured as an app, all the parameters are instated as-is.
	// Note: Pyroscope agent does not support `threadcreate` profile type, hence ignored.
	Parameters *caddy_profiling.Parameters `json:"parameters,omitempty"`

	// The pprof profiles to be collected. The accepted set of values is:
	// "cpu", "inuse_objects", "alloc_objects", "inuse_space", "alloc_space", "goroutines", "mutex_count", "mutex_duration", "block_count", "block_duration".
	// An empty set defaults to: "cpu", "alloc_objects", "alloc_space", "inuse_objects", "inuse_space".
	profileTypes []pyroscope.ProfileType

	profiler *pyroscope.Profiler
	logger   pyroscope.Logger
}

// ProfilingApp is the container of the `pyroscope` profiler if configured as a guest module of the `profiling` app
type ProfilingApp struct {
	App
}

func (*ProfilingApp) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID: "profiling.profiler.pyroscope",
		New: func() caddy.Module {
			return new(ProfilingApp)
		},
	}
}

// SetProfilingParameter sets the enabled Pyroscope profile types as configured by the `profiling` app.
// If the pyroscope app is configured with `profile_types`, then the ones specific to pyroscope take priority and the
// ones passed from the `profiling` app are ignored.
func (a *App) SetProfilingParameter(parameters caddy_profiling.Parameters) {
	if a.Parameters != nil {
		parameters = *a.Parameters
	}
	if a.Parameters != nil && len(a.Parameters.ProfileTypes) > 0 {
		return
	}
	for _, p := range parameters.ProfileTypes {
		switch p {
		case caddy_profiling.Goroutine:
			a.profileTypes = append(a.profileTypes, pyroscope.ProfileGoroutines)
		case caddy_profiling.Heap, caddy_profiling.Allocs:
			a.profileTypes = append(a.profileTypes, pyroscope.ProfileAllocObjects, pyroscope.ProfileInuseSpace, pyroscope.ProfileAllocSpace)
		case caddy_profiling.Threadcreate:
			a.logger.Infof("unsupported ProfileType: %s", p)
		case caddy_profiling.Block:
			a.profileTypes = append(a.profileTypes, pyroscope.ProfileBlockCount, pyroscope.ProfileBlockDuration)
		case caddy_profiling.Mutex:
			a.profileTypes = append(a.profileTypes, pyroscope.ProfileMutexCount, pyroscope.ProfileMutexDuration)
		}
	}
}

// CaddyModule implements caddy.Module
func (*App) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID: "pyroscope",
		New: func() caddy.Module {
			return new(App)
		},
	}
}

// Provision sets the profiling paramters per the configuration
func (a *App) Provision(ctx caddy.Context) error {
	logger := ctx.Logger()
	a.logger = logger.Sugar()
	repl := caddy.NewReplacer()

	if a.ApplicationName == "" {
		a.ApplicationName = "caddy"
	}
	a.ApplicationName = repl.ReplaceKnown(a.ApplicationName, a.ApplicationName)
	a.ServerAddress = repl.ReplaceKnown(a.ServerAddress, a.ServerAddress)

	// for k, v := range a.Tags {
	// 	a.Tags[repl.ReplaceKnown(k, k)] = repl.ReplaceKnown(v, v)
	// }
	a.AuthToken = repl.ReplaceKnown(a.AuthToken, a.AuthToken)

	if a.Parameters != nil {
		runtime.SetCPUProfileRate(a.Parameters.CPUProfileRate)
		runtime.SetBlockProfileRate(a.Parameters.BlockProfileRate)
		runtime.SetMutexProfileFraction(a.Parameters.MutexProfileFraction)
		a.SetProfilingParameter(*a.Parameters)
	}
	return nil
}

// Starts the Pyroscope session and the upload background routine
func (a *App) Start() (err error) {
	a.profiler, err = pyroscope.Start(pyroscope.Config{
		ApplicationName: a.ApplicationName,
		// Tags:            a.Tags,
		ServerAddress: a.ServerAddress,
		AuthToken:     a.AuthToken,
		Logger:        a.logger,
		ProfileTypes:  a.profileTypes,
		DisableGCRuns: a.DisableGCRuns,
	})
	return err
}

// Stops the Pyroscope session
func (a *App) Stop() error {
	return a.profiler.Stop()
}

var _ caddy.Module = (*App)(nil)
var _ caddy.Module = (*ProfilingApp)(nil)
var _ caddy.Provisioner = (*App)(nil)
var _ caddy.App = (*App)(nil)
var _ caddy_profiling.Profiler = (*App)(nil)
var _ caddy_profiling.ProfilingParameterSetter = (*App)(nil)
