package commands

import (
	"io/ioutil"

	"github.com/docker/app/internal/store"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/config"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// NewRootCmd returns the base root command.
func NewRootCmd(use string, dockerCli command.Cli) *cobra.Command {
	cmd := &cobra.Command{
		Short:       "Docker Application",
		Long:        `A tool to build and manage Docker Applications.`,
		Use:         use,
		Annotations: map[string]string{"experimentalCLI": "true"},
	}
	addCommands(cmd, dockerCli)
	return cmd
}

func addCommands(cmd *cobra.Command, dockerCli command.Cli) {
	cmd.AddCommand(
		installCmd(dockerCli),
		upgradeCmd(dockerCli),
		uninstallCmd(dockerCli),
		listCmd(dockerCli),
		statusCmd(dockerCli),
		initCmd(dockerCli),
		inspectCmd(dockerCli),
		mergeCmd(dockerCli),
		renderCmd(dockerCli),
		splitCmd(),
		validateCmd(),
		versionCmd(dockerCli),
		completionCmd(dockerCli, cmd),
		bundleCmd(dockerCli),
		pushCmd(dockerCli),
		pullCmd(dockerCli),
	)
}

func firstOrEmpty(list []string) string {
	if len(list) != 0 {
		return list[0]
	}
	return ""
}

func muteDockerCli(dockerCli command.Cli) func() {
	stdout := dockerCli.Out()
	stderr := dockerCli.Err()
	dockerCli.Apply(command.WithCombinedStreams(ioutil.Discard)) //nolint:errcheck // WithCombinedStreams cannot error
	return func() {
		dockerCli.Apply(command.WithOutputStream(stdout), command.WithErrorStream(stderr)) //nolint:errcheck // as above
	}
}

func prepareStores(targetContext string) (store.BundleStore, store.InstallationStore, store.CredentialStore, error) {
	appstore, err := store.NewApplicationStore(config.Dir())
	if err != nil {
		return nil, nil, nil, err
	}
	installationStore, err := appstore.InstallationStore(targetContext)
	if err != nil {
		return nil, nil, nil, err
	}
	bundleStore, err := appstore.BundleStore()
	if err != nil {
		return nil, nil, nil, err
	}
	credentialStore, err := appstore.CredentialStore(targetContext)
	if err != nil {
		return nil, nil, nil, err
	}
	return bundleStore, installationStore, credentialStore, nil
}

func prepareBundleStore() (store.BundleStore, error) {
	appstore, err := store.NewApplicationStore(config.Dir())
	if err != nil {
		return nil, err
	}
	bundleStore, err := appstore.BundleStore()
	if err != nil {
		return nil, err
	}
	return bundleStore, nil
}

type parametersOptions struct {
	parametersFiles []string
	overrides       []string
}

func (o *parametersOptions) addFlags(flags *pflag.FlagSet) {
	flags.StringArrayVar(&o.parametersFiles, "parameters-file", []string{}, "Override parameters file")
	flags.StringArrayVarP(&o.overrides, "set", "s", []string{}, "Override parameter value")
}

type credentialOptions struct {
	targetContext    string
	credentialsets   []string
	credentials      []string
	sendRegistryAuth bool
}

func (o *credentialOptions) addFlags(flags *pflag.FlagSet) {
	flags.StringVar(&o.targetContext, "target-context", "", "Context on which the application is installed (default: <current-context>)")
	flags.StringArrayVar(&o.credentialsets, "credential-set", []string{}, "Use a YAML file containing a credential set or a credential set present in the credential store")
	flags.StringArrayVar(&o.credentials, "credential", nil, "Add a single credential, additive ontop of any --credential-set used")
	flags.BoolVar(&o.sendRegistryAuth, "with-registry-auth", false, "Sends registry auth")
}

func (o *credentialOptions) SetDefaultTargetContext(dockerCli command.Cli) {
	o.targetContext = getTargetContext(o.targetContext, dockerCli.CurrentContext())
}

func (o *credentialOptions) CredentialSetOpts(dockerCli command.Cli, credentialStore store.CredentialStore) []credentialSetOpt {
	return []credentialSetOpt{
		addNamedCredentialSets(credentialStore, o.credentialsets),
		addCredentials(o.credentials),
		addDockerCredentials(o.targetContext, dockerCli.ContextStore()),
		addRegistryCredentials(o.sendRegistryAuth, dockerCli),
	}
}

type registryOptions struct {
	insecureRegistries []string
}

func (o *registryOptions) addFlags(flags *pflag.FlagSet) {
	flags.StringSliceVar(&o.insecureRegistries, "insecure-registries", nil, "Use HTTP instead of HTTPS when pulling from/pushing to those registries")
}

type pullOptions struct {
	pull bool
}

func (o *pullOptions) addFlags(flags *pflag.FlagSet) {
	flags.BoolVar(&o.pull, "pull", false, "Pull the bundle")
}
