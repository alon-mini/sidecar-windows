package monitor

// formModalDimensions returns the content width/height for the form modal.
func (m Model) formModalDimensions() (int, int) {
	modalWidth := m.Width * 80 / 100
	if modalWidth > 90 {
		modalWidth = 90
	}
	if modalWidth < 50 {
		modalWidth = 50
	}

	modalHeight := m.Height * 85 / 100
	if modalHeight > 35 {
		modalHeight = 35
	}
	if modalHeight < 20 {
		modalHeight = 20
	}

	return modalWidth, modalHeight
}
