package main

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"github.com/lxc/lxd/client"
	"github.com/lxc/lxd/lxc/config"
	"github.com/lxc/lxd/shared"
	"github.com/lxc/lxd/shared/api"
	cli "github.com/lxc/lxd/shared/cmd"
	"github.com/lxc/lxd/shared/i18n"
)

type cmdInfo struct {
	global *cmdGlobal

	flagShowLog   bool
	flagResources bool
	flagTarget    string
}

func (c *cmdInfo) Command() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Use = i18n.G("info [<remote>:][<container>]")
	cmd.Short = i18n.G("Show container or server information")
	cmd.Long = cli.FormatSection(i18n.G("Description"), i18n.G(
		`Show container or server information`))
	cmd.Example = cli.FormatSection("", i18n.G(
		`lxc info [<remote>:]<container> [--show-log]
    For container information.

lxc info [<remote>:] [--resources]
    For LXD server information.`))

	cmd.RunE = c.Run
	cmd.Flags().BoolVar(&c.flagShowLog, "show-log", false, i18n.G("Show the container's last 100 log lines?"))
	cmd.Flags().BoolVar(&c.flagResources, "resources", false, i18n.G("Show the resources available to the server"))
	cmd.Flags().StringVar(&c.flagTarget, "target", "", i18n.G("Cluster member name")+"``")

	return cmd
}

func (c *cmdInfo) Run(cmd *cobra.Command, args []string) error {
	conf := c.global.conf

	// Sanity checks
	exit, err := c.global.CheckArgs(cmd, args, 0, 1)
	if exit {
		return err
	}

	var remote string
	var cName string
	if len(args) == 1 {
		remote, cName, err = conf.ParseRemote(args[0])
		if err != nil {
			return err
		}
	} else {
		remote, cName, err = conf.ParseRemote("")
		if err != nil {
			return err
		}
	}

	d, err := conf.GetContainerServer(remote)
	if err != nil {
		return err
	}

	if cName == "" {
		return c.remoteInfo(d)
	}

	return c.containerInfo(d, conf.Remotes[remote], cName, c.flagShowLog)
}

func (c *cmdInfo) remoteInfo(d lxd.ContainerServer) error {
	// Targeting
	if c.flagTarget != "" {
		if !d.IsClustered() {
			return fmt.Errorf(i18n.G("To use --target, the destination remote must be a cluster"))
		}

		d = d.UseTarget(c.flagTarget)
	}

	if c.flagResources {
		resources, err := d.GetServerResources()
		if err != nil {
			return err
		}

		renderCPU := func(cpu api.ResourcesCPUSocket, prefix string) {
			if cpu.Vendor != "" {
				fmt.Printf(prefix+i18n.G("Vendor: %v")+"\n", cpu.Vendor)
			}

			if cpu.Name != "" {
				fmt.Printf(prefix+i18n.G("Name: %v")+"\n", cpu.Name)
			}

			fmt.Printf(prefix+i18n.G("Cores: %v")+"\n", cpu.Cores)
			fmt.Printf(prefix+i18n.G("Threads: %v")+"\n", cpu.Threads)

			if cpu.Frequency > 0 {
				if cpu.FrequencyTurbo > 0 {
					fmt.Printf(prefix+i18n.G("Frequency: %vMhz (max: %vMhz)")+"\n", cpu.Frequency, cpu.FrequencyTurbo)
				} else {
					fmt.Printf(prefix+i18n.G("Frequency: %vMhz")+"\n", cpu.Frequency)
				}
			}

			fmt.Printf(prefix+i18n.G("NUMA node: %v")+"\n", cpu.NUMANode)
		}

		if len(resources.CPU.Sockets) == 1 {
			fmt.Printf(i18n.G("CPU:") + "\n")
			renderCPU(resources.CPU.Sockets[0], "  ")
		} else if len(resources.CPU.Sockets) > 1 {
			fmt.Printf(i18n.G("CPUs:") + "\n")
			for _, cpu := range resources.CPU.Sockets {
				fmt.Printf("  "+i18n.G("Socket %d:")+"\n", cpu.Socket)
				renderCPU(cpu, "    ")
			}
		}

		fmt.Printf("\n" + i18n.G("Memory:") + "\n")
		fmt.Printf("  "+i18n.G("Free: %v")+"\n", shared.GetByteSizeString(int64(resources.Memory.Total-resources.Memory.Used), 2))
		fmt.Printf("  "+i18n.G("Used: %v")+"\n", shared.GetByteSizeString(int64(resources.Memory.Used), 2))
		fmt.Printf("  "+i18n.G("Total: %v")+"\n", shared.GetByteSizeString(int64(resources.Memory.Total), 2))

		renderGPU := func(gpu api.ResourcesGPUCard, prefix string) {
			if gpu.Vendor != "" {
				fmt.Printf(prefix+i18n.G("Vendor: %v (%v)")+"\n", gpu.Vendor, gpu.VendorID)
			}

			if gpu.Product != "" {
				fmt.Printf(prefix+i18n.G("Product: %v (%v)")+"\n", gpu.Product, gpu.ProductID)
			}

			fmt.Printf(prefix+i18n.G("PCI address: %v")+"\n", gpu.PCIAddress)
			fmt.Printf(prefix+i18n.G("Driver: %v (%v)")+"\n", gpu.Driver, gpu.DriverVersion)
			fmt.Printf(prefix+i18n.G("NUMA node: %v")+"\n", gpu.NUMANode)

			if gpu.Nvidia != nil {
				fmt.Printf(prefix + i18n.G("NVIDIA information:") + "\n")
				fmt.Printf(prefix+"  "+i18n.G("Architecture: %v")+"\n", gpu.Nvidia.Architecture)
				fmt.Printf(prefix+"  "+i18n.G("Brand: %v")+"\n", gpu.Nvidia.Brand)
				fmt.Printf(prefix+"  "+i18n.G("Model: %v")+"\n", gpu.Nvidia.Model)
				fmt.Printf(prefix+"  "+i18n.G("CUDA Version: %v")+"\n", gpu.Nvidia.CUDAVersion)
				fmt.Printf(prefix+"  "+i18n.G("NVRM Version: %v")+"\n", gpu.Nvidia.NVRMVersion)
				fmt.Printf(prefix+"  "+i18n.G("UUID: %v")+"\n", gpu.Nvidia.UUID)
			}
		}

		if len(resources.GPU.Cards) == 1 {
			fmt.Printf("\n" + i18n.G("GPU:") + "\n")
			renderGPU(resources.GPU.Cards[0], "  ")
		} else if len(resources.GPU.Cards) > 1 {
			fmt.Printf("\n" + i18n.G("GPUs:") + "\n")
			for _, gpu := range resources.GPU.Cards {
				fmt.Printf("  "+i18n.G("Card %d:")+"\n", gpu.ID)
				renderGPU(gpu, "    ")
			}
		}

		return nil
	}

	serverStatus, _, err := d.GetServer()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(&serverStatus)
	if err != nil {
		return err
	}

	fmt.Printf("%s", data)

	return nil
}

func (c *cmdInfo) containerInfo(d lxd.ContainerServer, remote config.Remote, name string, showLog bool) error {
	// Sanity checks
	if c.flagTarget != "" {
		return fmt.Errorf(i18n.G("--target cannot be used with containers"))
	}

	ct, _, err := d.GetContainer(name)
	if err != nil {
		return err
	}

	cs, _, err := d.GetContainerState(name)
	if err != nil {
		return err
	}

	const layout = "2006/01/02 15:04 UTC"

	fmt.Printf(i18n.G("Name: %s")+"\n", ct.Name)
	if ct.Location != "" {
		fmt.Printf(i18n.G("Location: %s")+"\n", ct.Location)
	}
	if remote.Addr != "" {
		fmt.Printf(i18n.G("Remote: %s")+"\n", remote.Addr)
	}

	fmt.Printf(i18n.G("Architecture: %s")+"\n", ct.Architecture)
	if shared.TimeIsSet(ct.CreatedAt) {
		fmt.Printf(i18n.G("Created: %s")+"\n", ct.CreatedAt.UTC().Format(layout))
	}

	fmt.Printf(i18n.G("Status: %s")+"\n", ct.Status)
	if ct.Ephemeral {
		fmt.Printf(i18n.G("Type: ephemeral") + "\n")
	} else {
		fmt.Printf(i18n.G("Type: persistent") + "\n")
	}
	fmt.Printf(i18n.G("Profiles: %s")+"\n", strings.Join(ct.Profiles, ", "))
	if cs.Pid != 0 {
		fmt.Printf(i18n.G("Pid: %d")+"\n", cs.Pid)

		// IP addresses
		ipInfo := ""
		if cs.Network != nil {
			for netName, net := range cs.Network {
				vethStr := ""
				if net.HostName != "" {
					vethStr = fmt.Sprintf("\t%s", net.HostName)
				}

				for _, addr := range net.Addresses {
					ipInfo += fmt.Sprintf("  %s:\t%s\t%s%s\n", netName, addr.Family, addr.Address, vethStr)
				}
			}
		}

		if ipInfo != "" {
			fmt.Println(i18n.G("Ips:"))
			fmt.Printf(ipInfo)
		}
		fmt.Println(i18n.G("Resources:"))

		// Processes
		fmt.Printf("  "+i18n.G("Processes: %d")+"\n", cs.Processes)

		// Disk usage
		diskInfo := ""
		if cs.Disk != nil {
			for entry, disk := range cs.Disk {
				if disk.Usage != 0 {
					diskInfo += fmt.Sprintf("    %s: %s\n", entry, shared.GetByteSizeString(disk.Usage, 2))
				}
			}
		}

		if diskInfo != "" {
			fmt.Println(fmt.Sprintf("  %s", i18n.G("Disk usage:")))
			fmt.Printf(diskInfo)
		}

		// CPU usage
		cpuInfo := ""
		if cs.CPU.Usage != 0 {
			cpuInfo += fmt.Sprintf("    %s: %v\n", i18n.G("CPU usage (in seconds)"), cs.CPU.Usage/1000000000)
		}

		if cpuInfo != "" {
			fmt.Println(fmt.Sprintf("  %s", i18n.G("CPU usage:")))
			fmt.Printf(cpuInfo)
		}

		// Memory usage
		memoryInfo := ""
		if cs.Memory.Usage != 0 {
			memoryInfo += fmt.Sprintf("    %s: %s\n", i18n.G("Memory (current)"), shared.GetByteSizeString(cs.Memory.Usage, 2))
		}

		if cs.Memory.UsagePeak != 0 {
			memoryInfo += fmt.Sprintf("    %s: %s\n", i18n.G("Memory (peak)"), shared.GetByteSizeString(cs.Memory.UsagePeak, 2))
		}

		if cs.Memory.SwapUsage != 0 {
			memoryInfo += fmt.Sprintf("    %s: %s\n", i18n.G("Swap (current)"), shared.GetByteSizeString(cs.Memory.SwapUsage, 2))
		}

		if cs.Memory.SwapUsagePeak != 0 {
			memoryInfo += fmt.Sprintf("    %s: %s\n", i18n.G("Swap (peak)"), shared.GetByteSizeString(cs.Memory.SwapUsagePeak, 2))
		}

		if memoryInfo != "" {
			fmt.Println(fmt.Sprintf("  %s", i18n.G("Memory usage:")))
			fmt.Printf(memoryInfo)
		}

		// Network usage
		networkInfo := ""
		if cs.Network != nil {
			for netName, net := range cs.Network {
				networkInfo += fmt.Sprintf("    %s:\n", netName)
				networkInfo += fmt.Sprintf("      %s: %s\n", i18n.G("Bytes received"), shared.GetByteSizeString(net.Counters.BytesReceived, 2))
				networkInfo += fmt.Sprintf("      %s: %s\n", i18n.G("Bytes sent"), shared.GetByteSizeString(net.Counters.BytesSent, 2))
				networkInfo += fmt.Sprintf("      %s: %d\n", i18n.G("Packets received"), net.Counters.PacketsReceived)
				networkInfo += fmt.Sprintf("      %s: %d\n", i18n.G("Packets sent"), net.Counters.PacketsSent)
			}
		}

		if networkInfo != "" {
			fmt.Println(fmt.Sprintf("  %s", i18n.G("Network usage:")))
			fmt.Printf(networkInfo)
		}
	}

	// List snapshots
	firstSnapshot := true
	snaps, err := d.GetContainerSnapshots(name)
	if err != nil {
		return nil
	}

	for _, snap := range snaps {
		if firstSnapshot {
			fmt.Println(i18n.G("Snapshots:"))
		}

		fields := strings.Split(snap.Name, shared.SnapshotDelimiter)
		fmt.Printf("  %s", fields[len(fields)-1])

		if shared.TimeIsSet(snap.CreatedAt) {
			fmt.Printf(" ("+i18n.G("taken at %s")+")", snap.CreatedAt.UTC().Format(layout))
		}

		if shared.TimeIsSet(snap.ExpiresAt) {
			fmt.Printf(" ("+i18n.G("expires at %s")+")", snap.ExpiresAt.UTC().Format(layout))
		}

		if snap.Stateful {
			fmt.Printf(" (" + i18n.G("stateful") + ")")
		} else {
			fmt.Printf(" (" + i18n.G("stateless") + ")")
		}
		fmt.Printf("\n")

		firstSnapshot = false
	}

	if showLog {
		log, err := d.GetContainerLogfile(name, "lxc.log")
		if err != nil {
			return err
		}

		stuff, err := ioutil.ReadAll(log)
		if err != nil {
			return err
		}

		fmt.Printf("\n"+i18n.G("Log:")+"\n\n%s\n", string(stuff))
	}

	return nil
}
