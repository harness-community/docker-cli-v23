package system

import (
	"context"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"

	"github.com/harness-community/docker-cli-v23/cli"
	pluginmanager "github.com/harness-community/docker-cli-v23/cli-plugins/manager"
	"github.com/harness-community/docker-cli-v23/cli/command"
	"github.com/harness-community/docker-cli-v23/cli/command/completion"
	"github.com/harness-community/docker-cli-v23/cli/command/formatter"
	"github.com/harness-community/docker-cli-v23/cli/debug"
	flagsHelper "github.com/harness-community/docker-cli-v23/cli/flags"
	"github.com/harness-community/docker-cli-v23/templates"
	"github.com/harness-community/docker-v23/api/types"
	"github.com/harness-community/docker-v23/api/types/swarm"
	"github.com/harness-community/docker-v23/api/types/versions"
	"github.com/docker/go-units"
	"github.com/spf13/cobra"
)

type infoOptions struct {
	format string
}

type clientInfo struct {
	Debug    bool
	Context  string
	Plugins  []pluginmanager.Plugin
	Warnings []string
}

type info struct {
	// This field should/could be ServerInfo but is anonymous to
	// preserve backwards compatibility in the JSON rendering
	// which has ServerInfo immediately within the top-level
	// object.
	*types.Info  `json:",omitempty"`
	ServerErrors []string `json:",omitempty"`

	ClientInfo   *clientInfo `json:",omitempty"`
	ClientErrors []string    `json:",omitempty"`
}

// NewInfoCommand creates a new cobra.Command for `docker info`
func NewInfoCommand(dockerCli command.Cli) *cobra.Command {
	var opts infoOptions

	cmd := &cobra.Command{
		Use:   "info [OPTIONS]",
		Short: "Display system-wide information",
		Args:  cli.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInfo(cmd, dockerCli, &opts)
		},
		Annotations: map[string]string{
			"category-top": "12",
			"aliases":      "docker system info, docker info",
		},
		ValidArgsFunction: completion.NoComplete,
	}

	cmd.Flags().StringVarP(&opts.format, "format", "f", "", flagsHelper.InspectFormatHelp)
	return cmd
}

func runInfo(cmd *cobra.Command, dockerCli command.Cli, opts *infoOptions) error {
	info := info{
		ClientInfo: &clientInfo{
			Context: dockerCli.CurrentContext(),
			Debug:   debug.IsEnabled(),
		},
		Info: &types.Info{},
	}
	if plugins, err := pluginmanager.ListPlugins(dockerCli, cmd.Root()); err == nil {
		info.ClientInfo.Plugins = plugins
	} else {
		info.ClientErrors = append(info.ClientErrors, err.Error())
	}

	if needsServerInfo(opts.format, info) {
		ctx := context.Background()
		if dinfo, err := dockerCli.Client().Info(ctx); err == nil {
			info.Info = &dinfo
		} else {
			info.ServerErrors = append(info.ServerErrors, err.Error())
			if opts.format == "" {
				// reset the server info to prevent printing "empty" Server info
				// and warnings, but don't reset it if a custom format was specified
				// to prevent errors from Go's template parsing during format.
				info.Info = nil
			} else {
				// if a format is provided, print the error, as it may be hidden
				// otherwise if the template doesn't include the ServerErrors field.
				fmt.Fprintln(dockerCli.Err(), err)
			}
		}
	}

	if opts.format == "" {
		return prettyPrintInfo(dockerCli, info)
	}
	return formatInfo(dockerCli, info, opts.format)
}

// placeHolders does a rudimentary match for possible placeholders in a
// template, matching a '.', followed by an letter (a-z/A-Z).
var placeHolders = regexp.MustCompile(`\.[a-zA-Z]`)

// needsServerInfo detects if the given template uses any server information.
// If only client-side information is used in the template, we can skip
// connecting to the daemon. This allows (e.g.) to only get cli-plugin
// information, without also making a (potentially expensive) API call.
func needsServerInfo(template string, info info) bool {
	if len(template) == 0 || placeHolders.FindString(template) == "" {
		// The template is empty, or does not contain formatting fields
		// (e.g. `table` or `raw` or `{{ json .}}`). Assume we need server-side
		// information to render it.
		return true
	}

	// A template is provided and has at least one field set.
	tmpl, err := templates.NewParse("", template)
	if err != nil {
		// ignore parsing errors here, and let regular code handle them
		return true
	}

	type sparseInfo struct {
		ClientInfo   *clientInfo `json:",omitempty"`
		ClientErrors []string    `json:",omitempty"`
	}

	// This constructs an "info" object that only has the client-side fields.
	err = tmpl.Execute(io.Discard, sparseInfo{
		ClientInfo:   info.ClientInfo,
		ClientErrors: info.ClientErrors,
	})
	// If executing the template failed, it means the template needs
	// server-side information as well. If it succeeded without server-side
	// information, we don't need to make API calls to collect that information.
	return err != nil
}

func prettyPrintInfo(dockerCli command.Cli, info info) error {
	fmt.Fprintln(dockerCli.Out(), "Client:")
	if info.ClientInfo != nil {
		prettyPrintClientInfo(dockerCli, *info.ClientInfo)
	}
	for _, err := range info.ClientErrors {
		fmt.Fprintln(dockerCli.Err(), "ERROR:", err)
	}

	fmt.Fprintln(dockerCli.Out())
	fmt.Fprintln(dockerCli.Out(), "Server:")
	if info.Info != nil {
		for _, err := range prettyPrintServerInfo(dockerCli, *info.Info) {
			info.ServerErrors = append(info.ServerErrors, err.Error())
		}
	}
	for _, err := range info.ServerErrors {
		fmt.Fprintln(dockerCli.Err(), "ERROR:", err)
	}

	if len(info.ServerErrors) > 0 || len(info.ClientErrors) > 0 {
		return fmt.Errorf("errors pretty printing info")
	}
	return nil
}

func prettyPrintClientInfo(dockerCli command.Cli, info clientInfo) {
	fmt.Fprintln(dockerCli.Out(), " Context:   ", info.Context)
	fmt.Fprintln(dockerCli.Out(), " Debug Mode:", info.Debug)

	if len(info.Plugins) > 0 {
		fmt.Fprintln(dockerCli.Out(), " Plugins:")
		for _, p := range info.Plugins {
			if p.Err == nil {
				fmt.Fprintf(dockerCli.Out(), "  %s: %s (%s)\n", p.Name, p.ShortDescription, p.Vendor)
				fprintlnNonEmpty(dockerCli.Out(), "    Version: ", p.Version)
				fprintlnNonEmpty(dockerCli.Out(), "    Path:    ", p.Path)
			} else {
				info.Warnings = append(info.Warnings, fmt.Sprintf("WARNING: Plugin %q is not valid: %s", p.Path, p.Err))
			}
		}
	}

	if len(info.Warnings) > 0 {
		fmt.Fprintln(dockerCli.Err(), strings.Join(info.Warnings, "\n"))
	}
}

//nolint:gocyclo
func prettyPrintServerInfo(dockerCli command.Cli, info types.Info) []error {
	var errs []error

	fmt.Fprintln(dockerCli.Out(), " Containers:", info.Containers)
	fmt.Fprintln(dockerCli.Out(), "  Running:", info.ContainersRunning)
	fmt.Fprintln(dockerCli.Out(), "  Paused:", info.ContainersPaused)
	fmt.Fprintln(dockerCli.Out(), "  Stopped:", info.ContainersStopped)
	fmt.Fprintln(dockerCli.Out(), " Images:", info.Images)
	fprintlnNonEmpty(dockerCli.Out(), " Server Version:", info.ServerVersion)
	fprintlnNonEmpty(dockerCli.Out(), " Storage Driver:", info.Driver)
	if info.DriverStatus != nil {
		for _, pair := range info.DriverStatus {
			fmt.Fprintf(dockerCli.Out(), "  %s: %s\n", pair[0], pair[1])
		}
	}
	if info.SystemStatus != nil {
		for _, pair := range info.SystemStatus {
			fmt.Fprintf(dockerCli.Out(), " %s: %s\n", pair[0], pair[1])
		}
	}
	fprintlnNonEmpty(dockerCli.Out(), " Logging Driver:", info.LoggingDriver)
	fprintlnNonEmpty(dockerCli.Out(), " Cgroup Driver:", info.CgroupDriver)
	fprintlnNonEmpty(dockerCli.Out(), " Cgroup Version:", info.CgroupVersion)

	fmt.Fprintln(dockerCli.Out(), " Plugins:")
	fmt.Fprintln(dockerCli.Out(), "  Volume:", strings.Join(info.Plugins.Volume, " "))
	fmt.Fprintln(dockerCli.Out(), "  Network:", strings.Join(info.Plugins.Network, " "))

	if len(info.Plugins.Authorization) != 0 {
		fmt.Fprintln(dockerCli.Out(), "  Authorization:", strings.Join(info.Plugins.Authorization, " "))
	}

	fmt.Fprintln(dockerCli.Out(), "  Log:", strings.Join(info.Plugins.Log, " "))

	fmt.Fprintln(dockerCli.Out(), " Swarm:", info.Swarm.LocalNodeState)
	printSwarmInfo(dockerCli, info)

	if len(info.Runtimes) > 0 {
		fmt.Fprint(dockerCli.Out(), " Runtimes:")
		for name := range info.Runtimes {
			fmt.Fprintf(dockerCli.Out(), " %s", name)
		}
		fmt.Fprint(dockerCli.Out(), "\n")
		fmt.Fprintln(dockerCli.Out(), " Default Runtime:", info.DefaultRuntime)
	}

	if info.OSType == "linux" {
		fmt.Fprintln(dockerCli.Out(), " Init Binary:", info.InitBinary)

		for _, ci := range []struct {
			Name   string
			Commit types.Commit
		}{
			{"containerd", info.ContainerdCommit},
			{"runc", info.RuncCommit},
			{"init", info.InitCommit},
		} {
			fmt.Fprintf(dockerCli.Out(), " %s version: %s", ci.Name, ci.Commit.ID)
			if ci.Commit.ID != ci.Commit.Expected {
				fmt.Fprintf(dockerCli.Out(), " (expected: %s)", ci.Commit.Expected)
			}
			fmt.Fprint(dockerCli.Out(), "\n")
		}
		if len(info.SecurityOptions) != 0 {
			if kvs, err := types.DecodeSecurityOptions(info.SecurityOptions); err != nil {
				errs = append(errs, err)
			} else {
				fmt.Fprintln(dockerCli.Out(), " Security Options:")
				for _, so := range kvs {
					fmt.Fprintln(dockerCli.Out(), "  "+so.Name)
					for _, o := range so.Options {
						switch o.Key {
						case "profile":
							fmt.Fprintln(dockerCli.Out(), "   Profile:", o.Value)
						}
					}
				}
			}
		}
	}

	// Isolation only has meaning on a Windows daemon.
	if info.OSType == "windows" {
		fmt.Fprintln(dockerCli.Out(), " Default Isolation:", info.Isolation)
	}

	fprintlnNonEmpty(dockerCli.Out(), " Kernel Version:", info.KernelVersion)
	fprintlnNonEmpty(dockerCli.Out(), " Operating System:", info.OperatingSystem)
	fprintlnNonEmpty(dockerCli.Out(), " OSType:", info.OSType)
	fprintlnNonEmpty(dockerCli.Out(), " Architecture:", info.Architecture)
	fmt.Fprintln(dockerCli.Out(), " CPUs:", info.NCPU)
	fmt.Fprintln(dockerCli.Out(), " Total Memory:", units.BytesSize(float64(info.MemTotal)))
	fprintlnNonEmpty(dockerCli.Out(), " Name:", info.Name)
	fprintlnNonEmpty(dockerCli.Out(), " ID:", info.ID)
	fmt.Fprintln(dockerCli.Out(), " Docker Root Dir:", info.DockerRootDir)
	fmt.Fprintln(dockerCli.Out(), " Debug Mode:", info.Debug)

	if info.Debug {
		fmt.Fprintln(dockerCli.Out(), "  File Descriptors:", info.NFd)
		fmt.Fprintln(dockerCli.Out(), "  Goroutines:", info.NGoroutines)
		fmt.Fprintln(dockerCli.Out(), "  System Time:", info.SystemTime)
		fmt.Fprintln(dockerCli.Out(), "  EventsListeners:", info.NEventsListener)
	}

	fprintlnNonEmpty(dockerCli.Out(), " HTTP Proxy:", info.HTTPProxy)
	fprintlnNonEmpty(dockerCli.Out(), " HTTPS Proxy:", info.HTTPSProxy)
	fprintlnNonEmpty(dockerCli.Out(), " No Proxy:", info.NoProxy)

	if info.IndexServerAddress != "" {
		u := dockerCli.ConfigFile().AuthConfigs[info.IndexServerAddress].Username
		if len(u) > 0 {
			fmt.Fprintln(dockerCli.Out(), " Username:", u)
		}
		fmt.Fprintln(dockerCli.Out(), " Registry:", info.IndexServerAddress)
	}

	if len(info.Labels) > 0 {
		fmt.Fprintln(dockerCli.Out(), " Labels:")
		for _, lbl := range info.Labels {
			fmt.Fprintln(dockerCli.Out(), "  "+lbl)
		}
	}

	fmt.Fprintln(dockerCli.Out(), " Experimental:", info.ExperimentalBuild)

	if info.RegistryConfig != nil && (len(info.RegistryConfig.InsecureRegistryCIDRs) > 0 || len(info.RegistryConfig.IndexConfigs) > 0) {
		fmt.Fprintln(dockerCli.Out(), " Insecure Registries:")
		for _, registry := range info.RegistryConfig.IndexConfigs {
			if !registry.Secure {
				fmt.Fprintln(dockerCli.Out(), "  "+registry.Name)
			}
		}

		for _, registry := range info.RegistryConfig.InsecureRegistryCIDRs {
			mask, _ := registry.Mask.Size()
			fmt.Fprintf(dockerCli.Out(), "  %s/%d\n", registry.IP.String(), mask)
		}
	}

	if info.RegistryConfig != nil && len(info.RegistryConfig.Mirrors) > 0 {
		fmt.Fprintln(dockerCli.Out(), " Registry Mirrors:")
		for _, mirror := range info.RegistryConfig.Mirrors {
			fmt.Fprintln(dockerCli.Out(), "  "+mirror)
		}
	}

	fmt.Fprintln(dockerCli.Out(), " Live Restore Enabled:", info.LiveRestoreEnabled)
	if info.ProductLicense != "" {
		fmt.Fprintln(dockerCli.Out(), " Product License:", info.ProductLicense)
	}

	if info.DefaultAddressPools != nil && len(info.DefaultAddressPools) > 0 {
		fmt.Fprintln(dockerCli.Out(), " Default Address Pools:")
		for _, pool := range info.DefaultAddressPools {
			fmt.Fprintf(dockerCli.Out(), "   Base: %s, Size: %d\n", pool.Base, pool.Size)
		}
	}

	fmt.Fprint(dockerCli.Out(), "\n")

	printServerWarnings(dockerCli, info)
	return errs
}

//nolint:gocyclo
func printSwarmInfo(dockerCli command.Cli, info types.Info) {
	if info.Swarm.LocalNodeState == swarm.LocalNodeStateInactive || info.Swarm.LocalNodeState == swarm.LocalNodeStateLocked {
		return
	}
	fmt.Fprintln(dockerCli.Out(), "  NodeID:", info.Swarm.NodeID)
	if info.Swarm.Error != "" {
		fmt.Fprintln(dockerCli.Out(), "  Error:", info.Swarm.Error)
	}
	fmt.Fprintln(dockerCli.Out(), "  Is Manager:", info.Swarm.ControlAvailable)
	if info.Swarm.Cluster != nil && info.Swarm.ControlAvailable && info.Swarm.Error == "" && info.Swarm.LocalNodeState != swarm.LocalNodeStateError {
		fmt.Fprintln(dockerCli.Out(), "  ClusterID:", info.Swarm.Cluster.ID)
		fmt.Fprintln(dockerCli.Out(), "  Managers:", info.Swarm.Managers)
		fmt.Fprintln(dockerCli.Out(), "  Nodes:", info.Swarm.Nodes)
		var strAddrPool strings.Builder
		if info.Swarm.Cluster.DefaultAddrPool != nil {
			for _, p := range info.Swarm.Cluster.DefaultAddrPool {
				strAddrPool.WriteString(p + "  ")
			}
			fmt.Fprintln(dockerCli.Out(), "  Default Address Pool:", strAddrPool.String())
			fmt.Fprintln(dockerCli.Out(), "  SubnetSize:", info.Swarm.Cluster.SubnetSize)
		}
		if info.Swarm.Cluster.DataPathPort > 0 {
			fmt.Fprintln(dockerCli.Out(), "  Data Path Port:", info.Swarm.Cluster.DataPathPort)
		}
		fmt.Fprintln(dockerCli.Out(), "  Orchestration:")

		taskHistoryRetentionLimit := int64(0)
		if info.Swarm.Cluster.Spec.Orchestration.TaskHistoryRetentionLimit != nil {
			taskHistoryRetentionLimit = *info.Swarm.Cluster.Spec.Orchestration.TaskHistoryRetentionLimit
		}
		fmt.Fprintln(dockerCli.Out(), "   Task History Retention Limit:", taskHistoryRetentionLimit)
		fmt.Fprintln(dockerCli.Out(), "  Raft:")
		fmt.Fprintln(dockerCli.Out(), "   Snapshot Interval:", info.Swarm.Cluster.Spec.Raft.SnapshotInterval)
		if info.Swarm.Cluster.Spec.Raft.KeepOldSnapshots != nil {
			fmt.Fprintf(dockerCli.Out(), "   Number of Old Snapshots to Retain: %d\n", *info.Swarm.Cluster.Spec.Raft.KeepOldSnapshots)
		}
		fmt.Fprintln(dockerCli.Out(), "   Heartbeat Tick:", info.Swarm.Cluster.Spec.Raft.HeartbeatTick)
		fmt.Fprintln(dockerCli.Out(), "   Election Tick:", info.Swarm.Cluster.Spec.Raft.ElectionTick)
		fmt.Fprintln(dockerCli.Out(), "  Dispatcher:")
		fmt.Fprintln(dockerCli.Out(), "   Heartbeat Period:", units.HumanDuration(info.Swarm.Cluster.Spec.Dispatcher.HeartbeatPeriod))
		fmt.Fprintln(dockerCli.Out(), "  CA Configuration:")
		fmt.Fprintln(dockerCli.Out(), "   Expiry Duration:", units.HumanDuration(info.Swarm.Cluster.Spec.CAConfig.NodeCertExpiry))
		fmt.Fprintln(dockerCli.Out(), "   Force Rotate:", info.Swarm.Cluster.Spec.CAConfig.ForceRotate)
		if caCert := strings.TrimSpace(info.Swarm.Cluster.Spec.CAConfig.SigningCACert); caCert != "" {
			fmt.Fprintf(dockerCli.Out(), "   Signing CA Certificate: \n%s\n\n", caCert)
		}
		if len(info.Swarm.Cluster.Spec.CAConfig.ExternalCAs) > 0 {
			fmt.Fprintln(dockerCli.Out(), "   External CAs:")
			for _, entry := range info.Swarm.Cluster.Spec.CAConfig.ExternalCAs {
				fmt.Fprintf(dockerCli.Out(), "     %s: %s\n", entry.Protocol, entry.URL)
			}
		}
		fmt.Fprintln(dockerCli.Out(), "  Autolock Managers:", info.Swarm.Cluster.Spec.EncryptionConfig.AutoLockManagers)
		fmt.Fprintln(dockerCli.Out(), "  Root Rotation In Progress:", info.Swarm.Cluster.RootRotationInProgress)
	}
	fmt.Fprintln(dockerCli.Out(), "  Node Address:", info.Swarm.NodeAddr)
	if len(info.Swarm.RemoteManagers) > 0 {
		managers := []string{}
		for _, entry := range info.Swarm.RemoteManagers {
			managers = append(managers, entry.Addr)
		}
		sort.Strings(managers)
		fmt.Fprintln(dockerCli.Out(), "  Manager Addresses:")
		for _, entry := range managers {
			fmt.Fprintf(dockerCli.Out(), "   %s\n", entry)
		}
	}
}

func printServerWarnings(dockerCli command.Cli, info types.Info) {
	if versions.LessThan(dockerCli.Client().ClientVersion(), "1.42") {
		printSecurityOptionsWarnings(dockerCli, info)
	}
	if len(info.Warnings) > 0 {
		fmt.Fprintln(dockerCli.Err(), strings.Join(info.Warnings, "\n"))
		return
	}
	// daemon didn't return warnings. Fallback to old behavior
	printServerWarningsLegacy(dockerCli, info)
}

// printSecurityOptionsWarnings prints warnings based on the security options
// returned by the daemon.
// DEPRECATED: warnings are now generated by the daemon, and returned in
// info.Warnings. This function is used to provide backward compatibility with
// daemons that do not provide these warnings. No new warnings should be added
// here.
func printSecurityOptionsWarnings(dockerCli command.Cli, info types.Info) {
	if info.OSType == "windows" {
		return
	}
	kvs, _ := types.DecodeSecurityOptions(info.SecurityOptions)
	for _, so := range kvs {
		if so.Name != "seccomp" {
			continue
		}
		for _, o := range so.Options {
			if o.Key == "profile" && o.Value != "default" && o.Value != "builtin" {
				_, _ = fmt.Fprintln(dockerCli.Err(), "WARNING: You're not using the default seccomp profile")
			}
		}
	}
}

// printServerWarningsLegacy generates warnings based on information returned by the daemon.
// DEPRECATED: warnings are now generated by the daemon, and returned in
// info.Warnings. This function is used to provide backward compatibility with
// daemons that do not provide these warnings. No new warnings should be added
// here.
func printServerWarningsLegacy(dockerCli command.Cli, info types.Info) {
	if info.OSType == "windows" {
		return
	}
	if !info.MemoryLimit {
		fmt.Fprintln(dockerCli.Err(), "WARNING: No memory limit support")
	}
	if !info.SwapLimit {
		fmt.Fprintln(dockerCli.Err(), "WARNING: No swap limit support")
	}
	if !info.OomKillDisable && info.CgroupVersion != "2" {
		fmt.Fprintln(dockerCli.Err(), "WARNING: No oom kill disable support")
	}
	if !info.CPUCfsQuota {
		fmt.Fprintln(dockerCli.Err(), "WARNING: No cpu cfs quota support")
	}
	if !info.CPUCfsPeriod {
		fmt.Fprintln(dockerCli.Err(), "WARNING: No cpu cfs period support")
	}
	if !info.CPUShares {
		fmt.Fprintln(dockerCli.Err(), "WARNING: No cpu shares support")
	}
	if !info.CPUSet {
		fmt.Fprintln(dockerCli.Err(), "WARNING: No cpuset support")
	}
	if !info.IPv4Forwarding {
		fmt.Fprintln(dockerCli.Err(), "WARNING: IPv4 forwarding is disabled")
	}
	if !info.BridgeNfIptables {
		fmt.Fprintln(dockerCli.Err(), "WARNING: bridge-nf-call-iptables is disabled")
	}
	if !info.BridgeNfIP6tables {
		fmt.Fprintln(dockerCli.Err(), "WARNING: bridge-nf-call-ip6tables is disabled")
	}
}

func formatInfo(dockerCli command.Cli, info info, format string) error {
	if format == formatter.JSONFormatKey {
		format = formatter.JSONFormat
	}

	// Ensure slice/array fields render as `[]` not `null`
	if info.ClientInfo != nil && info.ClientInfo.Plugins == nil {
		info.ClientInfo.Plugins = make([]pluginmanager.Plugin, 0)
	}

	tmpl, err := templates.Parse(format)
	if err != nil {
		return cli.StatusError{
			StatusCode: 64,
			Status:     "template parsing error: " + err.Error(),
		}
	}
	err = tmpl.Execute(dockerCli.Out(), info)
	dockerCli.Out().Write([]byte{'\n'})
	return err
}

func fprintlnNonEmpty(w io.Writer, label, value string) {
	if value != "" {
		fmt.Fprintln(w, label, value)
	}
}
