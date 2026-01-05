package main

import (
	"image"
	"image/color"
	_ "image/jpeg" // Indispensable pour décoder le JPEG (init function)
	"image/png"
	_ "image/png" // Indispensable pour décoder le PNG (init function)
	"log"
	"os"
)

// Pixel représente une valeur RGB
type Pixel struct {
	R, G, B uint16
}

// Block représente un bloc `factor x factor` extrait d'une image
// X, Y sont les coordonnées du coin supérieur gauche du bloc dans l'image
// Pixels contient les lignes de pixels du bloc et Avg est la couleur moyenne
type Block struct {
	X, Y   int
	Pixels [][]Pixel
	Avg    Pixel
}

// loadImage charge une image depuis un fichier
func loadImage(filename string) image.Image {
	reader, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()

	m, _, err := image.Decode(reader)
	if err != nil {
		log.Fatal(err)
	}
	return m
}

// extractPixels convertit une image en matrice de pixels RGB (séquentiel)
func extractPixels(m image.Image, width, height int) [][]Pixel {
	rgbMatrix := make([][]Pixel, height)

	for y := 0; y < height; y++ {
		rgbMatrix[y] = make([]Pixel, width)

		for x := 0; x < width; x++ {
			r, g, b, _ := m.At(x, y).RGBA()

			rgbMatrix[y][x] = Pixel{
				R: uint16(r),
				G: uint16(g),
				B: uint16(b),
			}
		}
	}
	return rgbMatrix
}

// blackWhiteSeq convertit la matrice en niveaux de gris (séquentiel, in-place)
func blackWhite(rgbMatrix [][]Pixel, width, height int) [][]Pixel {
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			p := rgbMatrix[y][x]
			gray := uint16(0.299*float64(p.R) + 0.587*float64(p.G) + 0.114*float64(p.B))
			rgbMatrix[y][x] = Pixel{R: gray, G: gray, B: gray}
		}
	}
	return rgbMatrix
}

// pixelsToImage convertit une matrice de pixels en image RGBA
func pixelsToImage(rgbMatrix [][]Pixel) *image.RGBA {
	if len(rgbMatrix) == 0 || len(rgbMatrix[0]) == 0 {
		return image.NewRGBA(image.Rect(0, 0, 0, 0))
	}
	height := len(rgbMatrix)
	width := len(rgbMatrix[0])
	out := image.NewRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			p := rgbMatrix[y][x]
			out.Set(x, y, color.RGBA{
				R: uint8(p.R >> 8),
				G: uint8(p.G >> 8),
				B: uint8(p.B >> 8),
				A: 255,
			})
		}
	}
	return out
}

// saveImage sauvegarde une image en PNG
func saveImage(img *image.RGBA, filename string) {
	file, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	err = png.Encode(file, img)
	if err != nil {
		log.Fatal(err)
	}
}

// downscalePixels réduit la définition sans changer la taille (pixelisation)
func downscalePixels(rgbMatrix [][]Pixel, width, height, factor int) [][]Pixel {
	if factor <= 1 {
		return rgbMatrix
	}
	if len(rgbMatrix) == 0 || len(rgbMatrix[0]) == 0 {
		return rgbMatrix
	}

	result := make([][]Pixel, height)
	for y := 0; y < height; y++ {
		result[y] = make([]Pixel, width)
	}

	for by := 0; by < height; by += factor {
		for bx := 0; bx < width; bx += factor {
			var sumR, sumG, sumB uint64
			count := 0
			maxY := by + factor
			if maxY > height {
				maxY = height
			}
			maxX := bx + factor
			if maxX > width {
				maxX = width
			}

			for y := by; y < maxY; y++ {
				for x := bx; x < maxX; x++ {
					p := rgbMatrix[y][x]
					sumR += uint64(p.R)
					sumG += uint64(p.G)
					sumB += uint64(p.B)
					count++
				}
			}

			avg := Pixel{
				R: uint16(sumR / uint64(count)),
				G: uint16(sumG / uint64(count)),
				B: uint16(sumB / uint64(count)),
			}

			for y := by; y < maxY; y++ {
				for x := bx; x < maxX; x++ {
					result[y][x] = avg
				}
			}
		}
	}

	return result
}

func colorDistance(a, b Pixel) float64 {
	dr := float64(a.R) - float64(b.R)
	dg := float64(a.G) - float64(b.G)
	db := float64(a.B) - float64(b.B)
	return dr*dr + dg*dg + db*db
}
func averageBlock(pixels [][]Pixel) Pixel {
	var r, g, b uint64
	count := 0

	for _, row := range pixels {
		for _, p := range row {
			r += uint64(p.R)
			g += uint64(p.G)
			b += uint64(p.B)
			count++
		}
	}

	return Pixel{
		R: uint16(r / uint64(count)),
		G: uint16(g / uint64(count)),
		B: uint16(b / uint64(count)),
	}
}
func splitIntoBlocks(img [][]Pixel, width, height, factor int) []Block {
	var blocks []Block

	for by := 0; by < height; by += factor {
		for bx := 0; bx < width; bx += factor {

			maxY := by + factor
			if maxY > height {
				maxY = height
			}
			maxX := bx + factor
			if maxX > width {
				maxX = width
			}

			blockPixels := make([][]Pixel, maxY-by)
			for y := by; y < maxY; y++ {
				blockPixels[y-by] = img[y][bx:maxX]
			}

			blocks = append(blocks, Block{
				X:      bx,
				Y:      by,
				Pixels: blockPixels,
				Avg:    averageBlock(blockPixels),
			})
		}
	}
	return blocks
}
func matchBlocks(source, target []Block) []Block {
	result := make([]Block, len(target))

	for i, tb := range target {
		best := source[0]
		bestDist := colorDistance(tb.Avg, best.Avg)

		for _, sb := range source[1:] {
			d := colorDistance(tb.Avg, sb.Avg)
			if d < bestDist {
				bestDist = d
				best = sb
			}
		}

		// placer le bloc source à la position du bloc cible
		best.X = tb.X
		best.Y = tb.Y
		result[i] = best
	}

	return result
}
func reconstructImage(blocks []Block, width, height int) [][]Pixel {
	result := make([][]Pixel, height)
	for y := 0; y < height; y++ {
		result[y] = make([]Pixel, width)
	}

	// When a source block (placed at a target position) extends beyond image
	// bounds (happens when dimensions are not divisible by factor), clip writes
	// to avoid panics.
	for _, b := range blocks {
		for by := range b.Pixels {
			ay := b.Y + by
			if ay < 0 || ay >= height {
				continue
			}
			for bx := range b.Pixels[by] {
				ax := b.X + bx
				if ax < 0 || ax >= width {
					continue
				}
				result[ay][ax] = b.Pixels[by][bx]
			}
		}
	}

	return result
}
func transformToTarget(
	source [][]Pixel,
	target [][]Pixel,
	width, height, factor int,
) [][]Pixel {

	srcBlocks := splitIntoBlocks(source, width, height, factor)
	tgtBlocks := splitIntoBlocks(target, width, height, factor)

	mapped := matchBlocks(srcBlocks, tgtBlocks)

	return reconstructImage(mapped, width, height)
}
