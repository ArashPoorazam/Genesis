package desktop

import (
	"fmt"
	"image"
	"os"

	"github.com/ArashPoorazam/Genesis/internal/simulation"
	"github.com/hajimehoshi/ebiten/v2"
)

func (driver *workflowDriver) relationshipLockSteps() []func() error {
	view := driver.view
	var original simulation.Config
	clickCell := func(row, column int, button ebiten.MouseButton) {
		bounds := view.matrixBounds.Add(view.inspectorRect().Min)
		point := image.Pt(bounds.Min.X+(column-view.firstVisibleColumn)*view.cellWidth+view.cellWidth/2, bounds.Min.Y+(row-view.firstVisibleRow)*view.cellHeight+view.cellHeight/2)
		view.matrixPointerPress(point, button)
	}
	return []func() error{
		func() error {
			original = view.controller.Simulation.Config.Clone()
			if err := view.editConfig(func(config *simulation.Config) error { config.SetRelationship(2, 3, .65); return nil }); err != nil {
				return err
			}
			clickCell(2, 3, ebiten.MouseButtonRight)
			return nil
		},
		func() error {
			if !view.controller.Simulation.Config.RelationshipLocked(2, 3) || !view.textFields[0].GetWidget().Disabled {
				return fmt.Errorf("right-click did not lock the hit cell and disable editing")
			}
			return driver.click("Shuffle relationships")
		},
		func() error {
			if view.controller.Simulation.Config.Relationships[2][3] != .65 {
				return fmt.Errorf("shuffle changed the locked relationship")
			}
			clickCell(3, 2, ebiten.MouseButtonLeft)
			return nil
		},
		func() error { return driver.click("Link pair: off") },
		func() error { return driver.submit(0, "0.9") },
		func() error {
			config := view.controller.Simulation.Config
			if config.Relationships[2][3] != .65 || config.Relationships[3][2] != .9 {
				return fmt.Errorf("reciprocal edit failed to respect the reverse lock")
			}
			return driver.click("Zero all")
		},
		func() error {
			config := view.controller.Simulation.Config
			for row := range config.Relationships {
				for column, strength := range config.Relationships[row] {
					expected := 0.0
					if row == 2 && column == 3 {
						expected = .65
					}
					if strength != expected {
						return fmt.Errorf("zero all did not preserve only the locked cell")
					}
				}
			}
			clickCell(2, 3, ebiten.MouseButtonLeft)
			driver.capturePath = os.Getenv("GENESIS_MATRIX_SCREENSHOT")
			return nil
		},
		func() error { clickCell(2, 3, ebiten.MouseButtonRight); return nil },
		func() error {
			if view.controller.Simulation.Config.RelationshipLocked(2, 3) || view.textFields[0].GetWidget().Disabled {
				return fmt.Errorf("second right-click did not restore editing")
			}
			return driver.click("Undo")
		},
		func() error {
			if !view.controller.Simulation.Config.RelationshipLocked(2, 3) {
				return fmt.Errorf("undo did not restore the cell lock")
			}
			if err := view.controller.Load(original, ""); err != nil {
				return err
			}
			view.selectedRow, view.selectedColumn, view.linked = 0, 0, false
			view.requestRebuild()
			return nil
		},
	}
}
