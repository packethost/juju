package packet

import (
	"github.com/juju/errors"
	"github.com/juju/juju/instance"
	"github.com/juju/juju/network"
	"github.com/juju/juju/status"
	"github.com/packethost/packngo"
)

type packetDevice struct {
	server packngo.Device
}

var _ instance.Instance = (*packetDevice)(nil)

func (inst packetDevice) Status() instance.InstanceStatus {
	instStatus := inst.server.State
	jujuStatus := status.Pending
	switch instStatus {
	case "queued", "provisioning", "powering_on":
		jujuStatus = status.Allocating
	case "active":
		jujuStatus = status.Running
	case "rebooting":
		jujuStatus = status.Rebooting
	case "powering_off", "inactive":
		jujuStatus = status.Empty
	case "failed":
		jujuStatus = status.ProvisioningError
	default:
		jujuStatus = status.Empty
	}
	return instance.InstanceStatus{
		Status:  jujuStatus,
		Message: instStatus,
	}
}

func (inst packetDevice) Id() instance.Id {
	return inst.server.ID
}

func (inst packetDevice) Addresses() ([]network.Address, error) {
	addresses := make([]network.Address, 0, len(inst.server.Network))
	for _, a := range inst.server.Network {
		address := network.NewAddress(a.Address)
		addresses = append(addresses, address)
	}

	return addresses, nil
}

func (inst packetDevice) OpenPorts(machineID string, ports []network.IngressRule) error {
	return errors.NotImplementedf("OpenPorts")
}

func (inst packetDevice) ClosePorts(machineID string, ports []network.IngressRule) error {
	return errors.NotImplementedf("ClosePorts")
}

func (inst packetDevice) IngressRules(machineID string) ([]network.IngressRule, error) {
	return nil, errors.NotImplementedf("InstanceRules")
}
