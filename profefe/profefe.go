package profefe

import (
	"net/http"
	"net/http/cookiejar"
	"runtime"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/mohammed90/caddy_profiling"
	"github.com/profefe/profefe/agent"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(new(App))
	caddy.RegisterModule(new(ProfilingApp))
}

const defaultDuration = 10 * time.Second

// The `profefe` app collects profiling data during the life-time of the process
// and uploads them to the profefe server.
type App struct {
	// The URL of the Profefe service. The config value may be a [placeholder](https://caddyserver.com/docs/conventions#placeholders).
	Address string `json:"address,omitempty"`

	// The service name reported to Profefe. The config value may be a [placeholder](https://caddyserver.com/docs/conventions#placeholders).
	Service string `json:"service,omitempty"`

	// The timeout for the upload call. Setting the value to `0` disables the timeout and the call waits indefinitely until the upload is finished.
	Timeout caddy.Duration `json:"timeout,omitempty"`

	// TODO: similar to `tags` in Pyroscope, decide on this field
	// Labels  []string       `json:"labels,omitempty"`

	// The profiling parameters to be reported to Profefe.
	// The paramters cpu_profile_rate, block_profile_rate, and mutex_profile_fraction are inherited from the `profiling` app if `profefe`
	// is configured as a child module. The `profile_types` field is inherited if not configured explicitly.
	// If `profefe` is configured as an app, all the parameters are instated as-is.
	Parameters *caddy_profiling.Parameters `json:"parameters,omitempty"`

	profefeOptions []agent.Option

	ctx        caddy.Context
	httpClient *http.Client
	agent      *agent.Agent
	logger     *zap.Logger
}

// ProfilingApp is the container of the `profefe` profiler if configured as a guest module of the `profiling` app
type ProfilingApp struct {
	App
}

// CaddyModule implements caddy.Module
func (*App) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID: "profefe",
		New: func() caddy.Module {
			return new(App)
		},
	}
}
func (*ProfilingApp) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID: "profiling.profiler.profefe",
		New: func() caddy.Module {
			return new(ProfilingApp)
		},
	}
}

// SetProfilingParameter sets the enabled Profefe profile types as configured by the `profiling` app.
// If the profefe app is configured with `profile_types`, then the ones specific to profefe take priority and the
// ones passed from the `profiling` app are ignored.
func (a *App) SetProfilingParameter(parameters caddy_profiling.Parameters) {
	if a.Parameters != nil {
		parameters = *a.Parameters
	}
	for _, p := range parameters.ProfileTypes {
		switch p {
		case caddy_profiling.CPU:
			a.profefeOptions = append(a.profefeOptions, agent.WithCPUProfile(defaultDuration))
		case caddy_profiling.Goroutine:
			a.profefeOptions = append(a.profefeOptions, agent.WithGoroutineProfile())
		case caddy_profiling.Heap, caddy_profiling.Allocs:
			a.profefeOptions = append(a.profefeOptions, agent.WithHeapProfile())
		case caddy_profiling.Threadcreate:
			a.profefeOptions = append(a.profefeOptions, agent.WithThreadcreateProfile())
		case caddy_profiling.Block:
			a.profefeOptions = append(a.profefeOptions, agent.WithBlockProfile())
		case caddy_profiling.Mutex:
			a.profefeOptions = append(a.profefeOptions, agent.WithMutexProfile())
		}
	}
}

// Provision implements caddy.Provisioner
func (p *App) Provision(ctx caddy.Context) error {
	p.logger = ctx.Logger()
	jar, _ := cookiejar.New(nil)
	p.httpClient = &http.Client{
		Jar:     jar,
		Timeout: time.Duration(p.Timeout),
	}
	// if len(p.Labels)%2 != 0 {
	// 	return fmt.Errorf("uneven number of labels: %d", len(p.Labels))
	// }
	repl := caddy.NewReplacer()

	p.Address = repl.ReplaceKnown(p.Address, p.Address)
	p.Service = repl.ReplaceKnown(p.Service, p.Service)

	p.profefeOptions = append(p.profefeOptions,
		// agent.WithLabels(p.Labels...),
		agent.WithHTTPClient(p.httpClient),
		agent.WithLogger(p.logger.Sugar().Infof),
	)
	if p.Parameters != nil {
		runtime.SetCPUProfileRate(p.Parameters.CPUProfileRate)
		runtime.SetBlockProfileRate(p.Parameters.BlockProfileRate)
		runtime.SetMutexProfileFraction(p.Parameters.MutexProfileFraction)
		p.SetProfilingParameter(*p.Parameters)
	}

	p.ctx = ctx
	return nil
}

// Start implements caddy.App
func (p *App) Start() error {
	a := agent.New(p.Address, p.Service, p.profefeOptions...)
	p.agent = a

	return p.agent.Start(p.ctx)
}

// Stop implements caddy.App
func (p *App) Stop() error {
	return p.agent.Stop()
}

var _ caddy.Module = (*App)(nil)
var _ caddy.App = (*App)(nil)
var _ caddy.Module = (*ProfilingApp)(nil)
var _ caddy.Provisioner = (*App)(nil)
var _ caddy_profiling.Profiler = (*App)(nil)
var _ caddy_profiling.ProfilingParameterSetter = (*App)(nil)
