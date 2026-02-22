package game

import (
	"image"
	_ "image/png"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
)

func loadBackgroundImage(path string) (*ebiten.Image, int, int, error) {
	img, err := decodeImage(path)
	if err != nil {
		return nil, 0, 0, err
	}

	ebitenImg := ebiten.NewImageFromImage(img)
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	return ebitenImg, width, height, nil
}

func loadSpriteSheetFrames(path string, frameWidth, frameHeight, frameCount int) ([]*ebiten.Image, error) {
	img, err := decodeImage(path)
	if err != nil {
		return nil, err
	}

	ebitenImg := ebiten.NewImageFromImage(img)
	bounds := img.Bounds()
	sheetWidth := bounds.Dx()

	frames := make([]*ebiten.Image, 0, frameCount)
	for i := 0; i < frameCount; i++ {
		col := i % (sheetWidth / frameWidth)
		row := i / (sheetWidth / frameWidth)
		x := col * frameWidth
		y := row * frameHeight

		frame := ebitenImg.SubImage(image.Rect(x, y, x+frameWidth, y+frameHeight)).(*ebiten.Image)
		frames = append(frames, frame)
	}

	return frames, nil
}

func decodeImage(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	return img, err
}
