package main

import (
	"image"
	"image/draw"
	"image/jpeg"
	"os"
	"fmt"
)

func createImagesHorizontal(imagePaths []string) (*image.RGBA, error) {
	if len(imagePaths) == 0 {
		return nil, fmt.Errorf("no hay imágenes para procesar")
	}

	var decodedImages []image.Image
	for _, imgPath := range imagePaths {
		imgFile, err := os.Open(imgPath)
		if err != nil {
			return nil, fmt.Errorf("error abriendo %s: %v", imgPath, err)
		}

		img, _, err := image.Decode(imgFile)
		imgFile.Close() // Cerrar inmediatamente después de usar
		if err != nil {
			return nil, fmt.Errorf("error decodificando %s: %v", imgPath, err)
		}
		decodedImages = append(decodedImages, img)
	}

	// Calcular dimensiones del canvas final
	totalWidth := 0
	maxHeight := 0
	for _, img := range decodedImages {
		bounds := img.Bounds()
		totalWidth += bounds.Dx()
		if bounds.Dy() > maxHeight {
			maxHeight = bounds.Dy()
		}
	}

	if totalWidth == 0 || maxHeight == 0 {
		return nil, fmt.Errorf("dimensiones inválidas")
	}

	// Crear canvas RGBA
	rgba := image.NewRGBA(image.Rect(0, 0, totalWidth, maxHeight))

	// Dibujar cada imagen en posición horizontal
	currentX := 0
	for _, img := range decodedImages {
		bounds := img.Bounds()
		drawRect := image.Rect(currentX, 0, currentX+bounds.Dx(), bounds.Dy())
		draw.Draw(rgba, drawRect, img, image.Point{0, 0}, draw.Src)
		currentX += bounds.Dx()
	}

	return rgba, nil
}

func compileToImg(c Compiler, outputPath string) error {
	var rowImages []*image.RGBA

	invert := true
	for rowIndex, row := range c.Rows {
		var imagePaths []string
		for stitchIndex, stitch := range row.Stitches {
			path := getImagePath(stitch)
			switch path {
			case "":
				return fmt.Errorf("punto desconocido en fila %d, posición %d: %T\n", 
					rowIndex, stitchIndex, stitch)
			case "ignore": 
					continue
			default:
				imagePaths = append(imagePaths, path)
			}
		}
		
		if len(imagePaths) == 0 {
			continue
		}

		if invert {
			imagePaths = reverse(imagePaths)
		}

		invert = !invert
		rowImage, err := createImagesHorizontal(imagePaths)
		if err != nil {
			return fmt.Errorf("error creando fila %d: %v", rowIndex, err)
		}
		rowImages = append(rowImages, rowImage)
	}

	invertRowImages := reverse(rowImages)
	finalImage, err := stackImagesVertically(invertRowImages)
	if err != nil {
		return fmt.Errorf("error apilando imágenes: %v", err)
	}

	// Guardar la imagen final
	out, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("error creando archivo de salida: %v", err)
	}
	defer out.Close()

	err = jpeg.Encode(out, finalImage, &jpeg.Options{Quality: 90})
	if err != nil {
		return fmt.Errorf("error codificando JPEG: %v", err)
	}

	return nil
}

func stackImagesVertically(images []*image.RGBA) (*image.RGBA, error) {
	if len(images) == 0 {
		return nil, fmt.Errorf("no hay imágenes para apilar")
	}

	// Calcular dimensiones totales
	maxWidth := 0
	totalHeight := 0
	for _, img := range images {
		if img == nil {
			return nil, fmt.Errorf("imagen nil encontrada")
		}
		bounds := img.Bounds()
		if bounds.Dx() > maxWidth {
			maxWidth = bounds.Dx()
		}
		totalHeight += bounds.Dy()
	}

	if maxWidth == 0 || totalHeight == 0 {
		return nil, fmt.Errorf("dimensiones inválidas para apilar")
	}

	// Crear canvas final
	finalImage := image.NewRGBA(image.Rect(0, 0, maxWidth, totalHeight))

	// Dibujar imágenes verticalmente
	currentY := 0
	for _, img := range images {
		bounds := img.Bounds()
		drawRect := image.Rect(0, currentY, bounds.Dx(), currentY+bounds.Dy())
		draw.Draw(finalImage, drawRect, img, image.Point{0, 0}, draw.Src)
		currentY += bounds.Dy()
	}

	return finalImage, nil
}

func getImagePath(stitch any) string {
	switch stitch.(type) {
	case *CableRC:
		return "rec/22rc.jpg"
	case *Knit:
		return "rec/k.jpg"
	case *Purl:
		return "rec/p.jpg"
	case *Ssk:
		return "rec/ssk.jpg"
	case *Ktog:
		return "rec/k2tog.jpg"
	case *Ptog:
		return "rec/p2tog.jpg"
	case *Yo:
		return "rec/yo.jpg"
	case *Co, *Bo:
		return "ignore"
	default:
		return ""
	}
}

func reverse[T any](list []T) []T {
    for i, j := 0, len(list)-1; i < j; {
        list[i], list[j] = list[j], list[i]
        i++
        j--
    }
    return list
}

