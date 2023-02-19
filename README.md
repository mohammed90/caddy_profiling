Continuous Profiling for Caddy
========================

The package contains 3 Caddy modules for push-mode continuous profiling support in Caddy.

[![Go Reference](https://pkg.go.dev/badge/github.com/mohammed90/caddy_profiling.svg)](https://pkg.go.dev/github.com/mohammed90/caddy_profiling)

## Modules

### `profiling`

The module can only serve as an app that is the host for multiple other profilers to configure and propagate common profiling parameters, i.e. [CPU profile rate](https://pkg.go.dev/runtime#SetCPUProfileRate), [block profile rate](https://pkg.go.dev/runtime#SetBlockProfileRate), [mutex profile fraction](https://pkg.go.dev/runtime#SetMutexProfileFraction), and the enabled profile types.

Documentation: [https://caddyserver.com/docs/json/apps/profiling/](https://caddyserver.com/docs/json/apps/profiling/)

#### Sample config

```json
{
	"apps": {
		"profiling": {
			"cpu_profile_rate": 2,
			"block_profile_rate": 2,
			"mutex_profile_fraction": 2,
			"profile_types": ["cpu","heap","allocs","goroutine"],
			"profilers": [
				{
					"profiler": "pyroscope",
					"server_address": "http://localhost:4040"
				},
				{
					"profiler": "pyroscope",
					"server_address": "http://another-instance.local:4040",
					"profile_types": ["heap"]
				},
				{
					"profiler": "profefe",
					// rest of config
				}
			]
		}
	}
}
```

### `pyroscope`

Configures and enables the push-mode [Pyroscope](https://pyroscope.io/) agent. May serve as an app or a child module of the `profiling` app. It may be configured as a child profiler of the `profilig` app or as first-level app within Caddy. Configuring the `pyroscope` app may look like this:

```json
{
	"apps": {
		"pyroscope": {
			"server_address": "http://localhost:4040",
			"application_name": "my_cool_app",
			"auth_token": "{env.PYROSCOPE_AUTH_TOKEN}",
			"parameters": {
				"cpu_profile_rate": 2,
				"block_profile_rate": 2,
				"mutex_profile_fraction": 2,
				"profile_types": ["cpu","heap","allocs","goroutine"]
			}
		}
	}
}
```

Documentation: [https://caddyserver.com/docs/json/apps/pyroscope/](https://caddyserver.com/docs/json/apps/pyroscope/)

### `profefe`

Similar to the `pyroscope` module, the `profefe` module configures and pushes data to [Profefe](https://github.com/profefe/profefe) server. May serve as an app or a child module of the `profiling` app. It may be configured as a child profiler of the `profilig` app or as first-level app within Caddy. Configuring the `profefe` app may look like this:

```json
{
	"apps": {
		"profefe": {
			"address": "http://localhost:4040",
			"service": "my_cool_app",
			"timeout": "10m",
			"parameters": {
				"cpu_profile_rate": 2,
				"block_profile_rate": 2,
				"mutex_profile_fraction": 2,
				"profile_types": ["cpu","heap","allocs","goroutine"]
			}
		}
	}
}
```

Documentation: [https://caddyserver.com/docs/json/apps/profefe/](https://caddyserver.com/docs/json/apps/profefe/)

## Available Profile Types

The `ProfileTypes` field in `Parameters` accepts the following values:

- `goroutine`: stack traces of all current goroutines
- `heap`: a sampling of memory allocations of live objects
- `allocs`: a sampling of all past memory allocations
- `threadcreate`: stack traces that led to the creation of new OS threads
- `block`: stack traces that led to blocking on synchronization primitives
- `mutex`: stack traces of holders of contended mutexes

They may be overridden by the individual guest profilers, if the guest module chooses to do so. The guest modules may not support the same types and they may have different names for the same types. The guest profilers own the translation logic.
