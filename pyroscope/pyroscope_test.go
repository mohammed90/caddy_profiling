package pyroscope

import (
	"testing"

	"github.com/grafana/pyroscope-go"
	"github.com/mohammed90/caddy_profiling"
)

func TestApp_SetProfilingParameter(t *testing.T) {
	type fields struct {
		Parameters   *caddy_profiling.Parameters
		profileTypes []pyroscope.ProfileType
	}
	type args struct {
		parameters caddy_profiling.Parameters
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "",
			fields: fields{
				Parameters: &caddy_profiling.Parameters{
					ProfileTypes: []caddy_profiling.ProfileType{"goroutine", "allocs", "block", "mutex"},
				},
				profileTypes: []pyroscope.ProfileType{},
			},
			args: args{
				parameters: caddy_profiling.Parameters{
					ProfileTypes: []caddy_profiling.ProfileType{},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &App{
				Parameters:   tt.fields.Parameters,
				profileTypes: tt.fields.profileTypes,
			}
			a.SetProfilingParameter(tt.args.parameters)
			t.Logf("%+v", a.profileTypes)
		})
	}
}
