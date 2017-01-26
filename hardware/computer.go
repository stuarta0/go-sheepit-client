package hardware

type Computer interface {
    // String core_script = "import bpy\n" + "bpy.context.user_preferences.system.compute_device_type = \"%s\"\n" + "bpy.context.scene.cycles.device = \"%s\"\n" + "bpy.context.user_preferences.system.compute_device = \"%s\"\n";
    // if using GPU and has GPU: core_script % ("CUDA", "GPU", gpu.CudaName())
    // else: core_script % ("NONE", "CPU", "CPU")

	GetDeviceName() string
	GetComputeDeviceType() string
	GetComputeDeviceName() string
	GetOptimalTileSize() int
}