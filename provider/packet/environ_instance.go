package packet

import (
	"github.com/juju/errors"
	"github.com/juju/juju/constraints"
	"github.com/juju/juju/environs"
	"github.com/juju/juju/environs/imagemetadata"
	"github.com/juju/juju/instance"
)

// AllInstances returns all instances currently known to the broker.
func (env *environ) AllInstances() ([]instance.Instance, error) {
	// Please note that this must *not* return instances that have not been
	// allocated as part of this environment -- if it does, juju will see they
	// are not tracked in state, assume they're stale/rogue, and shut them down.

	logger.Tracef("environ.AllInstances...")

	instances, err := env.client.envInstances()
	if err != nil {
		logger.Tracef("environ.AllInstances failed: %v", err)
		return nil, err
	}

	return instances, nil
}

func (env *environ) Instances(ids []instance.Id) ([]instance.Instance, error) {
	logger.Tracef("environ.Instances %#v", ids)
	// Please note that this must *not* return instances that have not been
	// allocated as part of this environment -- if it does, juju will see they
	// are not tracked in state, assume they're stale/rogue, and shut them down.
	// This advice applies even if an instance id passed in corresponds to a
	// real instance that's not part of the environment -- the Environ should
	// treat that no differently to a request for one that does not exist.

	instances, err := env.client.envInstances(ids...)
	if err != nil {
		logger.Tracef("environ.AllInstances failed: %v", err)
		return nil, err
	}

	if len(instances) == 0 {
		err := environs.ErrNoInstances
		return nil, err
	}
	return instances, nil
}

func (env *environ) MaintainInstance(args environs.StartInstanceParams) error {
	return nil
}

func (env *environ) PrecheckInstance(series string, cons constraints.Value, placement string) error {
	return nil
}

func (env *environ) PrepareForBootstrap(ctx environs.BootstrapContext) error {
	logger.Infof("preparing model %q", env.name)
	return nil
}

var findInstanceImage = func(matchingImages []*imagemetadata.ImageMetadata) (*imagemetadata.ImageMetadata, error) {
	if len(matchingImages) == 0 {
		return nil, errors.New("no matching image meta data")
	}
	return matchingImages[0], nil
}

func (env *environ) StartInstance(args environs.StartInstanceParams) (*environs.StartInstanceResult, error) {
	logger.Infof("starting Packet instance")
	res, err := env.client.newInstance(args)
	if err != nill {
		return nil, errors.Trace(err)
	}
	return res, nil
}

func (env *environ) StopInstances(instances ...instance.Id) error {
	logger.Debugf("stop instances %+v", instances)

	var err error

	for _, instance := range instances {
		if e := env.client.stopInstance(instance); e != nil {
			err = e
		}
	}

	return err
}
