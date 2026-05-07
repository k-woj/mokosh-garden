package game

import (
	"fmt"
	"image"
	_ "image/png"
	"io/fs"

	"github.com/hajimehoshi/ebiten/v2"
)

func loadImage(fsys fs.FS, path string) (*ebiten.Image, error) {
	img, err := decodeImage(fsys, path)
	if err != nil {
		return nil, err
	}
	return ebiten.NewImageFromImage(img), nil
}

func loadSpriteSheetFrames(fsys fs.FS, path string, frameWidth, frameHeight, frameCount int) ([]*ebiten.Image, error) {
	img, err := decodeImage(fsys, path)
	if err != nil {
		return nil, err
	}

	ebitenImg := ebiten.NewImageFromImage(img)
	framesPerRow := img.Bounds().Dx() / frameWidth

	frames := make([]*ebiten.Image, 0, frameCount)
	for i := 0; i < frameCount; i++ {
		col := i % framesPerRow
		row := i / framesPerRow
		x := col * frameWidth
		y := row * frameHeight
		frame := ebitenImg.SubImage(image.Rect(x, y, x+frameWidth, y+frameHeight)).(*ebiten.Image)
		frames = append(frames, frame)
	}
	return frames, nil
}

// loadPlantVariants loads up to 4 numbered sprite sheets (e.g. nettle_01.png … nettle_04.png).
// Missing numbers are silently skipped; returns an error only if none are found.
func loadPlantVariants(fsys fs.FS, dir, baseName string, frameW, frameH, frameCount int) ([][]*ebiten.Image, error) {
	var variants [][]*ebiten.Image
	for i := 1; i <= 4; i++ {
		path := fmt.Sprintf("%s/%s_%02d.png", dir, baseName, i)
		frames, err := loadSpriteSheetFrames(fsys, path, frameW, frameH, frameCount)
		if err != nil {
			continue
		}
		variants = append(variants, frames)
	}
	if len(variants) == 0 {
		return nil, fmt.Errorf("no sprite variants found for %s/%s", dir, baseName)
	}
	return variants, nil
}

func decodeImage(fsys fs.FS, path string) (image.Image, error) {
	file, err := fsys.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	return img, err
}
