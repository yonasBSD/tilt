package cisettings

import (
	"time"

	"go.starlark.net/starlark"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/tilt-dev/tilt/internal/tiltfile/starkit"
	"github.com/tilt-dev/tilt/internal/tiltfile/value"
	"github.com/tilt-dev/tilt/pkg/apis/core/v1alpha1"
	"github.com/tilt-dev/tilt/pkg/model"
)

// Implements functions for dealing with ci settings.
type Plugin struct {
	ciTimeoutFlag model.CITimeoutFlag
}

func NewPlugin(ciTimeoutFlag model.CITimeoutFlag) Plugin {
	return Plugin{
		ciTimeoutFlag: ciTimeoutFlag,
	}
}

func (e Plugin) NewState() interface{} {
	return model.DefaultSessionCISpec(e.ciTimeoutFlag)
}

func (e Plugin) OnStart(env *starkit.Environment) error {
	return env.AddBuiltin("ci_settings", e.ciSettings)
}

func (e *Plugin) ciSettings(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var k8sGracePeriod value.Duration = -1
	var timeout value.Duration = -1
	var readinessTimeout value.Duration = -1
	if err := starkit.UnpackArgs(thread, fn.Name(), args, kwargs,
		"k8s_grace_period?", &k8sGracePeriod,
		"timeout?", &timeout,
		"readiness_timeout?", &readinessTimeout); err != nil {
		return nil, err
	}

	err := starkit.SetState(thread, func(settings *v1alpha1.SessionCISpec) *v1alpha1.SessionCISpec {
		if k8sGracePeriod != -1 {
			settings = settings.DeepCopy()
			settings.K8sGracePeriod = &metav1.Duration{Duration: time.Duration(k8sGracePeriod)}
		}
		if timeout != -1 {
			settings = settings.DeepCopy()
			settings.Timeout = &metav1.Duration{Duration: time.Duration(timeout)}
		}
		if readinessTimeout != -1 {
			settings = settings.DeepCopy()
			settings.ReadinessTimeout = &metav1.Duration{Duration: time.Duration(readinessTimeout)}
		}
		return settings
	})

	return starlark.None, err
}

var _ starkit.StatefulPlugin = Plugin{}

func MustState(model starkit.Model) *v1alpha1.SessionCISpec {
	state, err := GetState(model)
	if err != nil {
		panic(err)
	}
	return state
}

func GetState(m starkit.Model) (*v1alpha1.SessionCISpec, error) {
	state := &v1alpha1.SessionCISpec{}
	err := m.Load(&state)
	return state, err
}
