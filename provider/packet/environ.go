package packet

import (
	"sync"

	"github.com/juju/errors"
	"github.com/juju/juju/constraints"
	"github.com/juju/juju/environs"
	"github.com/juju/juju/environs/config"
	"github.com/juju/juju/environs/instances"
	"github.com/juju/juju/instance"
	"github.com/juju/juju/provider/common"
	"github.com/juju/version"
)

type environ struct {
	name   string
	cloud  environs.CloudSpec
	client *environClient
	lock   sync.Mutex
	ecfg   *environConfig
}

// Name returns the Environ's name.
func (env *environ) Name() string {
	return env.name
}

// Provider returns the EnvironProvider that created this Environ.
func (*environ) Provider() environs.EnvironProvider {
	return providerInstance
}

// SetConfig updates the Environ's configuration.
//
// Calls to SetConfig do not affect the configuration of values previously obtained
// from Storage.
func (env *environ) SetConfig(cfg *config.Config) error {
	env.lock.Lock()
	defer env.lock.Unlock()

	ecfg, err := validateConfig(cfg, env.ecfg)
	if err != nil {
		return errors.Trace(err)
	}
	env.ecfg = ecfg

	return nil
}

func (env *environ) Config() *config.Config {
	return env.ecfg.Config
}

// AdoptResources is part of the Environ interface.
func (e *environ) AdoptResources(controllerUUID string, fromVersion version.Number) error {
	// This provider doesn't track instance -> controller.
	return nil
}

func (env *environ) Bootstrap(ctx environs.BootstrapContext, params environs.BootstrapParams) (*environs.BootstrapResult, error) {
	return common.Bootstrap(ctx, env, params)
}

// ControllerInstances is part of the Environ interface.
func (e *environ) ControllerInstances(controllerUUID string) ([]instance.Id, error) {
	return e.client.getControllerIds()
}

func (env *environ) Create(environs.CreateParams) error {
	return nil
}

func (env *environ) Destroy() error {
	// You can probably ignore this method; the common implementation should work.
	return common.Destroy(env)
}

func (env *environ) DestroyController(controllerUUID string) error {
	// TODO(wallyworld): destroy hosted model resources
	return env.Destroy()
}

func (e *environ) InstanceTypes(c constraints.Value) (instances.InstanceTypesWithCostMetadata, error) {
	result := instances.InstanceTypesWithCostMetadata{}
	return result, errors.NotSupportedf("InstanceTypes")
}
