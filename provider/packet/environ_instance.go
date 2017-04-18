package packet

import (
	"github.com/juju/errors"
	"github.com/juju/juju/cloudconfig/instancecfg"
	"github.com/juju/juju/cloudconfig/providerinit"
	"github.com/juju/juju/constraints"
	"github.com/juju/juju/environs"
	"github.com/juju/juju/instance"
	"github.com/juju/juju/tools"
	"github.com/juju/loggo"
	"github.com/packethost/packngo"
)

type packetDevice struct {
	server packngo.Device
}

// AllInstances returns all instances currently known to the broker.
func (env *environ) AllInstances() ([]instance.Instance, error) {
	// Please note that this must *not* return instances that have not been
	// allocated as part of this environment -- if it does, juju will see they
	// are not tracked in state, assume they're stale/rogue, and shut them down.

	logger.Tracef("environ.AllInstances...")

	servers, err := env.client.instances()
	if err != nil {
		logger.Tracef("environ.AllInstances failed: %v", err)
		return nil, err
	}

	instances := make([]instance.Instance, 0, len(servers))
	for _, server := range servers {
		instance := packetDevice{server: server}
		instances = append(instances, instance)
	}

	if logger.LogLevel() <= loggo.TRACE {
		logger.Tracef("All instances, len = %d:", len(instances))
		for _, instance := range instances {
			logger.Tracef("... id: %q, status: %q", instance.Id(), instance.Status())
		}
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

	m, err := env.client.instanceMap()
	if err != nil {
		return nil, errors.Annotate(err, "environ.Instances failed")
	}

	var found int
	r := make([]instance.Instance, len(ids))
	for i, id := range ids {
		if s, ok := m[string(id)]; ok {
			r[i] = packetDevice{server: s}
			found++
		}
	}

	if found == 0 {
		err = environs.ErrNoInstances
	} else if found != len(ids) {
		err = environs.ErrPartialInstances
	}

	return r, errors.Trace(err)
}

func (*environ) MaintainInstance(args environs.StartInstanceParams) error {
	return nil
}

func (env *environ) PrecheckInstance(series string, cons constraints.Value, placement string) error {
	return nil
}

func (env *environ) PrepareForBootstrap(ctx environs.BootstrapContext) error {
	logger.Infof("preparing model %q", env.name)
	return nil
}

func (env *environ) StartInstance(args environs.StartInstanceParams) (*environs.StartInstanceResult, error) {
	logger.Infof("Packet environ.StartInstance...")

	if args.InstanceConfig == nil {
		return nil, errors.New("instance configuration is nil")
	}

	if len(args.Tools) == 0 {
		return nil, errors.New("tools not found")
	}

	img, err := findInstanceImage(args.ImageMetadata)
	if err != nil {
		return nil, err
	}

	tools, err := args.Tools.Match(tools.Filter{Arch: img.Arch})
	if err != nil {
		return nil, errors.Errorf("chosen architecture %v not present in %v", img.Arch, args.Tools.Arches())
	}

	if err := args.InstanceConfig.SetTools(tools); err != nil {
		return nil, errors.Trace(err)
	}
	if err := instancecfg.FinishInstanceConfig(args.InstanceConfig, env.Config()); err != nil {
		return nil, err
	}
	userData, err := providerinit.ComposeUserData(args.InstanceConfig, nil, PacketRenderer{})
	if err != nil {
		return nil, errors.Annotate(err, "cannot make user data")
	}

	logger.Debugf("packet user data; %d bytes", len(userData))

	client := env.client
	cfg := env.Config()
	server, rootdrive, arch, err := client.newInstance(args, img, userData, cfg.AuthorizedKeys())
	if err != nil {
		return nil, errors.Errorf("failed start instance: %v", err)
	}

	inst := &packetDevice{server: server}

	// prepare hardware characteristics
	hwch, err := inst.hardware(arch, rootdrive.Size())
	if err != nil {
		return nil, err
	}

	logger.Debugf("hardware: %v", hwch)
	return &environs.StartInstanceResult{
		Instance: inst,
		Hardware: hwch,
	}, nil
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
