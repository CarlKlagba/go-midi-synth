package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
)

func main() {
	a := app.New()
	w := a.NewWindow("Sexy Synth")
	w.Resize(fyne.NewSize(400, 300))

	waveLabel := widget.NewLabel("wave?")
	sinButton := widget.NewButton("Sinusoidal", func() {
		waveLabel.SetText("Sinusoidal")
	})
	sawButton := widget.NewButton("Sawtooth", func() {
		waveLabel.SetText("Sawtooth")
	})
	squareButton := widget.NewButton("Square", func() {
		waveLabel.SetText("Square")
	})
	triangleButton := widget.NewButton("Triangle", func() {
		waveLabel.SetText("Triangle")
	})

	waveBox := container.NewVBox(
		waveLabel,
		sinButton,
		sawButton,
		squareButton,
		triangleButton,
	)

	volume := 50.0
	volumeBinding := binding.BindFloat(&volume)
	volumeSlider := widget.NewSliderWithData(0, 100, volumeBinding)

	labelSlider := widget.NewLabelWithData(binding.FloatToString(volumeBinding))

	volumeBox := container.NewVBox(
		widget.NewLabel("Volume"),
		volumeSlider,
		labelSlider,
	)

	mainBox := container.NewVBox(
		waveBox,
		volumeBox,
	)
	w.SetContent(mainBox)

	w.ShowAndRun()
}
