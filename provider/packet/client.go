package packet

import (
	"github.com/juju/errors"
	"github.com/juju/juju/environs"
	"github.com/juju/juju/instance"
	"github.com/juju/juju/providerinit"
	"github.com/packethost/packngo"
)

const (
	consumerToken = "p9ACZWqkb8EJeyea2TjvSgCwTwUc91s2iniHfbW4fKLEvV3AuDkG13FWepd1PLxX"
	controllerTag = "controller"
)

type environClient struct {
	conn      *packngo.Client
	projectID string
	uuid      string
}

var newClient = func(cloud environs.CloudSpec, uuid string) (client *environClient, err error) {
	logger.Debugf("creating Packet client: id=%q", uuid)

	credAttrs := cloud.Credential.Attributes()
	packetAPIToken := credAttrs[credAPIToken]
	projectID := credAttrs[credProjectID]

	conn := packngo.NewClient(consumerToken, packetAPIToken, nil)

	client = &environClient{
		conn:      conn,
		uuid:      uuid,
		projectID: projectID,
	}

	return client, nil
}

func (c *environClient) getInstanceList() ([]instance.Instance, error) {

	devices, _, err := c.conn.Devices.List(c.projectID)
	if err != nil {
		return nil, errors.Trace(err)
	}
	instanceList := make([]instance.Instance, len(devices))
	for i, d := range devices {
		instanceList[i] = packetDevice{server: d}
		//instanceList = append(instanceList, packetDevice{server: d}.(instance.Instance))
	}
	return instanceList, nil
}

type selector func(instance.Instance) bool

func all(flist []selector, i instance.Instance) bool {
	for _, f := range flist {
		if !f(i) {
			return false
		}
	}
	return true
}

func selectInstances(instances []instance.Instance, flist []selector) []instance.Instance {
	retList := make([]instance.Instance, len(instances))
	for _, i := range instances {
		if all(flist, i) {
			retList = append(retList, i)
		}
	}
	return retList
}

func getIdSliceMembershipTest(idSlice []instance.Id) selector {
	return func(i instance.Instance) bool {
		for _, id := range idSlice {
			if id == i.Id() {
				return true
			}
		}
		return false
	}
}

func (c *environClient) envInstances(ids ...instance.Id) ([]instance.Instance, error) {
	instanceList, err := c.getInstanceList()
	if err != nil {
		return nil, err
	}
	inThisEnv := getTagTest(c.uuid)
	filters := []selector{inThisEnv}
	if len(ids) > 0 {
		inListedIds := getIdSliceMembershipTest(ids)
		filters = append(filters, inListedIds)
	}

	return selectInstances(instanceList, filters), nil
}

func getTagTest(tag string) selector {
	return func(i instance.Instance) bool {
		tags := i.(packetDevice).server.Tags
		for _, t := range tags {
			if tag == t {
				return true
			}
		}
		return false
	}
}

func getIds(instanceList []instance.Instance) []instance.Id {
	retList := make([]instance.Id, len(instanceList))
	for _, i := range instanceList {
		retList = append(retList, i.Id())
	}
	return retList
}

func (c *environClient) stopInstance(id instance.Id) error {
	uuid := string(id)
	if uuid == "" {
		return errors.New("invalid instance id")
	}
	_, err := c.conn.Devices.Delete(uuid)

	if err != nil {
		return errors.Trace(err)
	}

	return nil
}

func (c *environClient) getControllerIds() ([]instance.Id, error) {
	instanceList, err := c.getInstanceList()
	if err != nil {
		return nil, errors.Trace(err)
	}
	inThisEnv := getTagTest(c.uuid)
	isController := getTagTest(controllerTag)

	controllers := selectInstances(instanceList, []selector{inThisEnv, isController})

	return getIds(controllers), nil

}

//        var data = `{
//        "hostname": "aaa",
//        "plan": "baremetal_0",
//        "facility": "sjc1",
//        "operating_system": "ubuntu_16_04",
//        "billing_cycle": "hourly",
//        "project_id": "89b497ee-5afc-420a-8fb5-56984898f4df",
//        "tags": ["t1", "t2"]
//        }`

func (c *environClient) newInstance(args environs.StartInstanceParams) (*environs.StartInstanceResult, error) {
	// see ../{joyent,cloudsigma}/client.go for ideas
	cr := packngo.DeviceCreateRequest{}
	cr.BillingCycle = "hourly"
	cr.Tags = []string{c.uuid}
	cr.ProjectID = c.projectID
	cr.UserData = providerinit.ComposeUserData(args.InstanceConfig, nil, PakcetRenderer{})
	return nil, nil
}
