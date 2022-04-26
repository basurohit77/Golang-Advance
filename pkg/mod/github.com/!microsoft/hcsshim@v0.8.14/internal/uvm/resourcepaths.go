package uvm

const (
	gpuResourcePath                  string = "VirtualMachine/ComputeTopology/Gpu"
	memoryResourcePath               string = "VirtualMachine/ComputeTopology/Memory/SizeInMB"
	cpuGroupResourcePath             string = "VirtualMachine/ComputeTopology/Processor/CpuGroup"
	idledResourcePath                string = "VirtualMachine/ComputeTopology/Processor/IdledProcessors"
	cpuFrequencyPowerCapResourcePath string = "VirtualMachine/ComputeTopology/Processor/CpuFrequencyPowerCap"
	cpuLimitsResourcePath            string = "VirtualMachine/ComputeTopology/Processor/Limits"
	serialResourceFormat             string = "VirtualMachine/Devices/ComPorts/%d"
	flexibleIovResourceFormat        string = "VirtualMachine/Devices/FlexibleIov/%s"
	licensingResourcePath            string = "VirtualMachine/Devices/Licensing"
	mappedPipeResourceFormat         string = "VirtualMachine/Devices/MappedPipes/%s"
	networkResourceFormat            string = "VirtualMachine/Devices/NetworkAdapters/%s"
	plan9ShareResourcePath           string = "VirtualMachine/Devices/Plan9/Shares"
	scsiResourceFormat               string = "VirtualMachine/Devices/Scsi/%s/Attachments/%d"
	sharedMemoryRegionResourcePath   string = "VirtualMachine/Devices/SharedMemory/Regions"
	virtualPciResourceFormat         string = "VirtualMachine/Devices/VirtualPci/%s"
	vPMemControllerResourceFormat    string = "VirtualMachine/Devices/VirtualPMem/Devices/%d"
	vPMemDeviceResourceFormat        string = "VirtualMachine/Devices/VirtualPMem/Devices/%d/Mappings/%d"
	vSmbShareResourcePath            string = "VirtualMachine/Devices/VirtualSmb/Shares"
	hvsocketConfigResourceFormat     string = "VirtualMachine/Devices/HvSocket/HvSocketConfig/ServiceTable/%s"
)
