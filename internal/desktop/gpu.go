package desktop

import "github.com/ArashPoorazam/Genesis/internal/gpuphysics"

func (view *Window) setGPUBoost(enabled bool) {
	if !enabled {
		view.closeGPU()
		return
	}
	if view.gpuEngine == nil {
		engine, err := gpuphysics.New()
		if err != nil {
			view.status = "GPU boost unavailable; using CPU: " + err.Error()
			return
		}
		view.gpuEngine = engine
	}
	view.controller.Accelerator = view.gpuEngine
	view.controller.Message = ""
	view.status = ""
}

func (view *Window) closeGPU() {
	if view.gpuEngine != nil {
		view.gpuEngine.Close()
		view.gpuEngine = nil
	}
	view.controller.Accelerator = nil
}
