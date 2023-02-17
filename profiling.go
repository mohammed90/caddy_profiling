package caddy_profiling

import (
	"encoding/json"
	"errors"
	"fmt"
	"runtime"

	"github.com/caddyserver/caddy/v2"
)

func init() {
	caddy.RegisterModule(new(App))
}

type App struct {
	Parameters
	ProfilersRaw []json.RawMessage `json:"profilers,omitempty" caddy:"namespace=profiling.profiler inline_key=profiler"`

	profilers []Profiler
}

// CaddyModule implements caddy.Module
func (*App) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID: "profiling",
		New: func() caddy.Module {
			return new(App)
		},
	}
}

// Provision implements caddy.Provisioner
func (a *App) Provision(ctx caddy.Context) error {
	mods, err := ctx.LoadModule(a, "ProfilersRaw")
	if err != nil {
		return fmt.Errorf("loading profiler module: %v", err)
	}
	for _, mod := range mods.([]any) {
		if m, ok := mod.(ProfilingParameterSetter); ok {
			m.SetProfilingParameter(a.Parameters)
		}
		a.profilers = append(a.profilers, mod.(Profiler))
	}

	// set the values here in case any of the child profilers changed them
	runtime.SetCPUProfileRate(a.CPUProfileRate)
	runtime.SetBlockProfileRate(a.BlockProfileRate)
	runtime.SetMutexProfileFraction(a.MutexProfileFraction)
	return nil
}

// Starts all the child profilers to initiate the periodic push
func (a *App) Start() (err error) {
	var startedProfilers []Profiler
	for _, p := range a.profilers {
		e := p.Start()
		if e != nil {
			err = errors.Join(err, e)
			for _, sp := range startedProfilers {
				err = errors.Join(err, sp.Stop())
			}
			return err
		}
	}
	return err
}

// Stops all the child profilers to halt the periodic push
func (a *App) Stop() (err error) {
	for _, p := range a.profilers {
		err = errors.Join(err, p.Stop())
	}
	return err
}

type ProfilingParameterSetter interface {
	SetProfilingParameter(Parameters)
}

type Profiler interface {
	Start() error
	Stop() error
}

var _ caddy.Module = (*App)(nil)
var _ caddy.App = (*App)(nil)
var _ caddy.Provisioner = (*App)(nil)
