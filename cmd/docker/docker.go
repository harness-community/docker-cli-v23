package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/harness-community/docker-cli-v23/cli"
	pluginmanager "github.com/harness-community/docker-cli-v23/cli-plugins/manager"
	"github.com/harness-community/docker-cli-v23/cli/command"
	"github.com/harness-community/docker-cli-v23/cli/command/commands"
	cliflags "github.com/harness-community/docker-cli-v23/cli/flags"
	"github.com/harness-community/docker-cli-v23/cli/version"
	"github.com/harness-community/docker-v23/api/types/versions"
	"github.com/moby/buildkit/util/appcontext"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func newDockerCommand(dockerCli *command.DockerCli) *cli.TopLevelCommand {
	var (
		opts    *cliflags.ClientOptions
		flags   *pflag.FlagSet
		helpCmd *cobra.Command
	)

	cmd := &cobra.Command{
		Use:              "docker [OPTIONS] COMMAND [ARG...]",
		Short:            "A self-sufficient runtime for containers",
		SilenceUsage:     true,
		SilenceErrors:    true,
		TraverseChildren: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return command.ShowHelp(dockerCli.Err())(cmd, args)
			}
			return fmt.Errorf("docker: '%s' is not a docker command.\nSee 'docker --help'", args[0])
		},
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return isSupported(cmd, dockerCli)
		},
		Version:               fmt.Sprintf("%s, build %s", version.Version, version.GitCommit),
		DisableFlagsInUseLine: true,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd:   false,
			HiddenDefaultCmd:    true,
			DisableDescriptions: true,
		},
	}
	cmd.SetIn(dockerCli.In())
	cmd.SetOut(dockerCli.Out())
	cmd.SetErr(dockerCli.Err())

	opts, flags, helpCmd = cli.SetupRootCommand(cmd)
	registerCompletionFuncForGlobalFlags(dockerCli, cmd)
	flags.BoolP("version", "v", false, "Print version information and quit")
	setFlagErrorFunc(dockerCli, cmd)

	setupHelpCommand(dockerCli, cmd, helpCmd)
	setHelpFunc(dockerCli, cmd)

	cmd.SetOut(dockerCli.Out())
	commands.AddCommands(cmd, dockerCli)

	cli.DisableFlagsInUseLine(cmd)
	setValidateArgs(dockerCli, cmd)

	// flags must be the top-level command flags, not cmd.Flags()
	return cli.NewTopLevelCommand(cmd, dockerCli, opts, flags)
}

func setFlagErrorFunc(dockerCli command.Cli, cmd *cobra.Command) {
	// When invoking `docker stack --nonsense`, we need to make sure FlagErrorFunc return appropriate
	// output if the feature is not supported.
	// As above cli.SetupRootCommand(cmd) have already setup the FlagErrorFunc, we will add a pre-check before the FlagErrorFunc
	// is called.
	flagErrorFunc := cmd.FlagErrorFunc()
	cmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		if err := pluginmanager.AddPluginCommandStubs(dockerCli, cmd.Root()); err != nil {
			return err
		}
		if err := isSupported(cmd, dockerCli); err != nil {
			return err
		}
		if err := hideUnsupportedFeatures(cmd, dockerCli); err != nil {
			return err
		}
		return flagErrorFunc(cmd, err)
	})
}

func setupHelpCommand(dockerCli command.Cli, rootCmd, helpCmd *cobra.Command) {
	origRun := helpCmd.Run
	origRunE := helpCmd.RunE

	helpCmd.Run = nil
	helpCmd.RunE = func(c *cobra.Command, args []string) error {
		if len(args) > 0 {
			helpcmd, err := pluginmanager.PluginRunCommand(dockerCli, args[0], rootCmd)
			if err == nil {
				return helpcmd.Run()
			}
			if !pluginmanager.IsNotFound(err) {
				return errors.Errorf("unknown help topic: %v", strings.Join(args, " "))
			}
		}
		if origRunE != nil {
			return origRunE(c, args)
		}
		origRun(c, args)
		return nil
	}
}

func tryRunPluginHelp(dockerCli command.Cli, ccmd *cobra.Command, cargs []string) error {
	root := ccmd.Root()

	cmd, _, err := root.Traverse(cargs)
	if err != nil {
		return err
	}
	helpcmd, err := pluginmanager.PluginRunCommand(dockerCli, cmd.Name(), root)
	if err != nil {
		return err
	}
	return helpcmd.Run()
}

func setHelpFunc(dockerCli command.Cli, cmd *cobra.Command) {
	defaultHelpFunc := cmd.HelpFunc()
	cmd.SetHelpFunc(func(ccmd *cobra.Command, args []string) {
		if err := pluginmanager.AddPluginCommandStubs(dockerCli, ccmd.Root()); err != nil {
			ccmd.Println(err)
			return
		}

		if len(args) >= 1 {
			err := tryRunPluginHelp(dockerCli, ccmd, args)
			if err == nil {
				return
			}
			if !pluginmanager.IsNotFound(err) {
				ccmd.Println(err)
				return
			}
		}

		if err := isSupported(ccmd, dockerCli); err != nil {
			ccmd.Println(err)
			return
		}
		if err := hideUnsupportedFeatures(ccmd, dockerCli); err != nil {
			ccmd.Println(err)
			return
		}

		defaultHelpFunc(ccmd, args)
	})
}

func setValidateArgs(dockerCli command.Cli, cmd *cobra.Command) {
	// The Args is handled by ValidateArgs in cobra, which does not allows a pre-hook.
	// As a result, here we replace the existing Args validation func to a wrapper,
	// where the wrapper will check to see if the feature is supported or not.
	// The Args validation error will only be returned if the feature is supported.
	cli.VisitAll(cmd, func(ccmd *cobra.Command) {
		// if there is no tags for a command or any of its parent,
		// there is no need to wrap the Args validation.
		if !hasTags(ccmd) {
			return
		}

		if ccmd.Args == nil {
			return
		}

		cmdArgs := ccmd.Args
		ccmd.Args = func(cmd *cobra.Command, args []string) error {
			if err := isSupported(cmd, dockerCli); err != nil {
				return err
			}
			return cmdArgs(cmd, args)
		}
	})
}

func tryPluginRun(dockerCli command.Cli, cmd *cobra.Command, subcommand string, envs []string) error {
	plugincmd, err := pluginmanager.PluginRunCommand(dockerCli, subcommand, cmd)
	if err != nil {
		return err
	}
	plugincmd.Env = append(envs, plugincmd.Env...)

	go func() {
		// override SIGTERM handler so we let the plugin shut down first
		<-appcontext.Context().Done()
	}()

	if err := plugincmd.Run(); err != nil {
		statusCode := 1
		exitErr, ok := err.(*exec.ExitError)
		if !ok {
			return err
		}
		if ws, ok := exitErr.Sys().(syscall.WaitStatus); ok {
			statusCode = ws.ExitStatus()
		}
		return cli.StatusError{
			StatusCode: statusCode,
		}
	}
	return nil
}

func runDocker(dockerCli *command.DockerCli) error {
	tcmd := newDockerCommand(dockerCli)

	cmd, args, err := tcmd.HandleGlobalFlags()
	if err != nil {
		return err
	}

	if err := tcmd.Initialize(); err != nil {
		return err
	}

	var envs []string
	args, os.Args, envs, err = processAliases(dockerCli, cmd, args, os.Args)
	if err != nil {
		return err
	}

	if cli.HasCompletionArg(args) {
		// We add plugin command stubs early only for completion. We don't
		// want to add them for normal command execution as it would cause
		// a significant performance hit.
		err = pluginmanager.AddPluginCommandStubs(dockerCli, cmd)
		if err != nil {
			return err
		}
	}

	if len(args) > 0 {
		ccmd, _, err := cmd.Find(args)
		if err != nil || pluginmanager.IsPluginCommand(ccmd) {
			err := tryPluginRun(dockerCli, cmd, args[0], envs)
			if !pluginmanager.IsNotFound(err) {
				return err
			}
			// For plugin not found we fall through to
			// cmd.Execute() which deals with reporting
			// "command not found" in a consistent way.
		}
	}

	// We've parsed global args already, so reset args to those
	// which remain.
	cmd.SetArgs(args)
	return cmd.Execute()
}

func main() {
	dockerCli, err := command.NewDockerCli()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	logrus.SetOutput(dockerCli.Err())

	if err := runDocker(dockerCli); err != nil {
		if sterr, ok := err.(cli.StatusError); ok {
			if sterr.Status != "" {
				fmt.Fprintln(dockerCli.Err(), sterr.Status)
			}
			// StatusError should only be used for errors, and all errors should
			// have a non-zero exit status, so never exit with 0
			if sterr.StatusCode == 0 {
				os.Exit(1)
			}
			os.Exit(sterr.StatusCode)
		}
		fmt.Fprintln(dockerCli.Err(), err)
		os.Exit(1)
	}
}

type versionDetails interface {
	CurrentVersion() string
	ServerInfo() command.ServerInfo
}

func hideFlagIf(f *pflag.Flag, condition func(string) bool, annotation string) {
	if f.Hidden {
		return
	}
	var val string
	if values, ok := f.Annotations[annotation]; ok {
		if len(values) > 0 {
			val = values[0]
		}
		if condition(val) {
			f.Hidden = true
		}
	}
}

func hideSubcommandIf(subcmd *cobra.Command, condition func(string) bool, annotation string) {
	if subcmd.Hidden {
		return
	}
	if v, ok := subcmd.Annotations[annotation]; ok {
		if condition(v) {
			subcmd.Hidden = true
		}
	}
}

func hideUnsupportedFeatures(cmd *cobra.Command, details versionDetails) error {
	var (
		notExperimental = func(_ string) bool { return !details.ServerInfo().HasExperimental }
		notOSType       = func(v string) bool { return details.ServerInfo().OSType != "" && v != details.ServerInfo().OSType }
		notSwarmStatus  = func(v string) bool {
			s := details.ServerInfo().SwarmStatus
			if s == nil {
				// engine did not return swarm status header
				return false
			}
			switch v {
			case "manager":
				// requires the node to be a manager
				return !s.ControlAvailable
			case "active":
				// requires swarm to be active on the node (e.g. for swarm leave)
				// only hide the command if we're sure the node is "inactive"
				// for any other status, assume the "leave" command can still
				// be used.
				return s.NodeState == "inactive"
			case "":
				// some swarm commands, such as "swarm init" and "swarm join"
				// are swarm-related, but do not require swarm to be active
				return false
			default:
				// ignore any other value for the "swarm" annotation
				return false
			}
		}
		versionOlderThan = func(v string) bool { return versions.LessThan(details.CurrentVersion(), v) }
	)

	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		// hide flags not supported by the server
		// root command shows all top-level flags
		if cmd.Parent() != nil {
			if cmds, ok := f.Annotations["top-level"]; ok {
				f.Hidden = !findCommand(cmd, cmds)
			}
			if f.Hidden {
				return
			}
		}

		hideFlagIf(f, notExperimental, "experimental")
		hideFlagIf(f, notOSType, "ostype")
		hideFlagIf(f, notSwarmStatus, "swarm")
		hideFlagIf(f, versionOlderThan, "version")
	})

	for _, subcmd := range cmd.Commands() {
		hideSubcommandIf(subcmd, notExperimental, "experimental")
		hideSubcommandIf(subcmd, notOSType, "ostype")
		hideSubcommandIf(subcmd, notSwarmStatus, "swarm")
		hideSubcommandIf(subcmd, versionOlderThan, "version")
	}
	return nil
}

// Checks if a command or one of its ancestors is in the list
func findCommand(cmd *cobra.Command, commands []string) bool {
	if cmd == nil {
		return false
	}
	for _, c := range commands {
		if c == cmd.Name() {
			return true
		}
	}
	return findCommand(cmd.Parent(), commands)
}

func isSupported(cmd *cobra.Command, details versionDetails) error {
	if err := areSubcommandsSupported(cmd, details); err != nil {
		return err
	}
	return areFlagsSupported(cmd, details)
}

func areFlagsSupported(cmd *cobra.Command, details versionDetails) error {
	errs := []string{}

	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if !f.Changed {
			return
		}
		if !isVersionSupported(f, details.CurrentVersion()) {
			errs = append(errs, fmt.Sprintf(`"--%s" requires API version %s, but the Docker daemon API version is %s`, f.Name, getFlagAnnotation(f, "version"), details.CurrentVersion()))
			return
		}
		if !isOSTypeSupported(f, details.ServerInfo().OSType) {
			errs = append(errs, fmt.Sprintf(
				`"--%s" is only supported on a Docker daemon running on %s, but the Docker daemon is running on %s`,
				f.Name,
				getFlagAnnotation(f, "ostype"), details.ServerInfo().OSType),
			)
			return
		}
		if _, ok := f.Annotations["experimental"]; ok && !details.ServerInfo().HasExperimental {
			errs = append(errs, fmt.Sprintf(`"--%s" is only supported on a Docker daemon with experimental features enabled`, f.Name))
		}
		// buildkit-specific flags are noop when buildkit is not enabled, so we do not add an error in that case
	})
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "\n"))
	}
	return nil
}

// Check recursively so that, e.g., `docker stack ls` returns the same output as `docker stack`
func areSubcommandsSupported(cmd *cobra.Command, details versionDetails) error {
	// Check recursively so that, e.g., `docker stack ls` returns the same output as `docker stack`
	for curr := cmd; curr != nil; curr = curr.Parent() {
		// Important: in the code below, calls to "details.CurrentVersion()" and
		// "details.ServerInfo()" are deliberately executed inline to make them
		// be executed "lazily". This is to prevent making a connection with the
		// daemon to perform a "ping" (even for commands that do not require a
		// daemon connection).
		//
		// See commit b39739123b845f872549e91be184cc583f5b387c for details.

		if cmdVersion, ok := curr.Annotations["version"]; ok && versions.LessThan(details.CurrentVersion(), cmdVersion) {
			return fmt.Errorf("%s requires API version %s, but the Docker daemon API version is %s", cmd.CommandPath(), cmdVersion, details.CurrentVersion())
		}
		if ost, ok := curr.Annotations["ostype"]; ok && details.ServerInfo().OSType != "" && ost != details.ServerInfo().OSType {
			return fmt.Errorf("%s is only supported on a Docker daemon running on %s, but the Docker daemon is running on %s", cmd.CommandPath(), ost, details.ServerInfo().OSType)
		}
		if _, ok := curr.Annotations["experimental"]; ok && !details.ServerInfo().HasExperimental {
			return fmt.Errorf("%s is only supported on a Docker daemon with experimental features enabled", cmd.CommandPath())
		}
	}
	return nil
}

func getFlagAnnotation(f *pflag.Flag, annotation string) string {
	if value, ok := f.Annotations[annotation]; ok && len(value) == 1 {
		return value[0]
	}
	return ""
}

func isVersionSupported(f *pflag.Flag, clientVersion string) bool {
	if v := getFlagAnnotation(f, "version"); v != "" {
		return versions.GreaterThanOrEqualTo(clientVersion, v)
	}
	return true
}

func isOSTypeSupported(f *pflag.Flag, osType string) bool {
	if v := getFlagAnnotation(f, "ostype"); v != "" && osType != "" {
		return osType == v
	}
	return true
}

// hasTags return true if any of the command's parents has tags
func hasTags(cmd *cobra.Command) bool {
	for curr := cmd; curr != nil; curr = curr.Parent() {
		if len(curr.Annotations) > 0 {
			return true
		}
	}

	return false
}
