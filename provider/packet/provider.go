package packet

import (
	"github.com/juju/errors"
	"github.com/juju/jsonschema"
	"github.com/juju/juju/cloud"
	"github.com/juju/juju/environs"
	"github.com/juju/juju/environs/config"
	"github.com/juju/loggo"
)

var logger = loggo.GetLogger("juju.provider.packet")

const (
	providerType = "packet"
)

type environProvider struct {
	environProviderCredentials
}

var providerInstance = environProvider{}

var _ environs.EnvironProvider = (*environProvider)(nil)

func init() {
	environs.RegisterProvider(providerType, providerInstance)
}

func (environProvider) Open(args environs.OpenParams) (environs.Environ, error) {
	logger.Infof("opening model %q", args.Config.Name())
	if err := validateCloudSpec(args.Cloud); err != nil {
		return nil, errors.Annotate(err, "validating cloud spec")
	}

	client, err := newClient(args.Cloud, args.Config.UUID())
	if err != nil {
		return nil, errors.Trace(err)
	}
	env := &environ{
		name:   args.Config.Name(),
		cloud:  args.Cloud,
		client: client,
	}
	if err := env.SetConfig(args.Config); err != nil {
		return nil, err
	}

	return env, nil
}

func (p environProvider) CloudSchema() *jsonschema.Schema {
	return nil
}

func (p environProvider) Ping(endpoint string) error {
	return errors.NotImplementedf("Ping")
}

// PrepareConfig is defined by EnvironProvider.
func (environProvider) PrepareConfig(args environs.PrepareConfigParams) (*config.Config, error) {
	if err := validateCloudSpec(args.Cloud); err != nil {
		return nil, errors.Annotate(err, "validating cloud spec")
	}
	return args.Config, nil
}

func (environProvider) Validate(cfg, old *config.Config) (*config.Config, error) {
	logger.Infof("validating model %q", cfg.Name())

	newEcfg, err := validateConfig(cfg, nil)
	if err != nil {
		return nil, errors.Errorf("invalid config: %v", err)
	}
	if old != nil {
		oldEcfg, err := validateConfig(old, nil)
		if err != nil {
			return nil, errors.Errorf("invalid base config: %v", err)
		}
		if newEcfg, err = validateConfig(cfg, oldEcfg); err != nil {
			return nil, errors.Errorf("invalid config change: %v", err)
		}
	}

	return newEcfg.Config, nil
}

func validateCloudSpec(spec environs.CloudSpec) error {
	if err := spec.Validate(); err != nil {
		return errors.Trace(err)
	}
	if spec.Credential == nil {
		return errors.NotValidf("missing credential")
	}
	if authType := spec.Credential.AuthType(); authType != cloud.ApiTokenAuthType {
		return errors.NotSupportedf("%q auth-type", authType)
	}
	return nil
}
