package desktop

import "testing"

func TestWorkspaceGeometry(t *testing.T) {
	for _, size := range [][2]int{{1024, 768}, {1440, 900}, {1920, 1080}, {3440, 1440}} {
		view := &Window{width: size[0], height: size[1]}
		screen := rectangle(0, 0, view.width, view.height)
		for _, expanded := range []bool{false, true} {
			view.matrixExpanded = expanded
			for _, visible := range []bool{false, true} {
				view.inspectorVisible = visible
				if !view.drawerHandleRect().In(screen) {
					t.Errorf("%v expanded=%t visible=%t: handle extends offscreen", size, expanded, visible)
				}
				if !view.inspectorRect().In(screen) {
					t.Errorf("%v: inspector extends offscreen", size)
				}
				if view.playbackRect().Overlaps(view.inspectorRect()) {
					t.Errorf("%v: playback overlaps inspector", size)
				}
			}
		}
	}
}
