package packet

import (
	"github.com/juju/juju/environs"
	"github.com/packethost/packngo"
)

type environClient struct {
	conn *packngo.Client
	uuid string
}

var newClient = func(cloud environs.CloudSpec, uuid string) (client *environClient, err error) {
	logger.Debugf("creating Packet client: id=%q", uuid)

	credAttrs := cloud.Credential.Attributes()
	packetApiToken := credAttrs[credApiToken]

	// create connection to CloudSigma
	conn := packngo.NewClient("", packetApiToken, nil)

	client = &environClient{
		conn: conn,
		uuid: uuid,
	}

	return client, nil
}
