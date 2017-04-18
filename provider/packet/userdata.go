package packet

import (
	"github.com/juju/errors"
	"github.com/juju/utils"
	jujuos "github.com/juju/utils/os"

	"github.com/juju/juju/cloudconfig/cloudinit"
	"github.com/juju/juju/cloudconfig/providerinit/renderers"
)

type PacketRenderer struct{}

func (PacketRenderer) Render(cfg cloudinit.CloudConfig, os jujuos.OSType) ([]byte, error) {
	switch os {
	case jujuos.Ubuntu:
		return renderers.RenderYAML(cfg, utils.Gzip, renderers.ToBase64)
	case jujuos.CentOS:
		return renderers.RenderScript(cfg, renderers.ToBase64)
	case jujuos.Windows:
		return renderers.RenderYAML(cfg, renderers.WinEmbedInScript, renderers.ToBase64)
	default:
		return nil, errors.Errorf("Cannot encode userdata for OS: %s", os)
	}
}
