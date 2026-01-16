package main

import (
	"image"
	"image/color"
	_ "image/jpeg" // Indispensable pour décoder le JPEG (init function)
	"image/png"
	_ "image/png" // Indispensable pour décoder le PNG (init function)
	"log"
	"math/rand"
	"os"
)

// Pixel représente une valeur RGB
type Pixel struct {
	R, G, B uint16
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

// copyMatrix crée une copie profonde de la matrice
func copyMatrix(src [][]Pixel) [][]Pixel {
	dst := make([][]Pixel, len(src))
	for i := range src {
		dst[i] = append([]Pixel{}, src[i]...)
	}
	return dst
}

// blackWhite convertit la matrice en niveaux de gris (séquentiel, in-place)
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

// Remap pixels from a source image to match the color distribution of a target image
// without changing pixel values—only their positions. Images must share dimensions.

// quantizePixel maps a Pixel to a bin index using `levels` discrete values per channel.
// Example: levels=16 -> 4096 bins.
func quantizePixel(p Pixel, levels int) int {
	step := 65536 / levels
	qR := int(p.R) / step
	qG := int(p.G) / step
	qB := int(p.B) / step
	if qR >= levels {
		qR = levels - 1
	}
	if qG >= levels {
		qG = levels - 1
	}
	if qB >= levels {
		qB = levels - 1
	}
	return (qR*levels+qG)*levels + qB
}

// binCenter returns the approximate center value (0..65535) of a bin on each channel.
func binCenter(binIdx, levels int) (int, int, int) {
	step := 65536 / levels
	plane := levels * levels
	qR := binIdx / plane
	rem := binIdx % plane
	qG := rem / levels
	qB := rem % levels
	cR := qR*step + step/2
	cG := qG*step + step/2
	cB := qB*step + step/2
	return cR, cG, cB
}

func sqDist(aR, aG, aB, bR, bG, bB int) int {
	dR := aR - bR
	dG := aG - bG
	dB := aB - bB
	return dR*dR + dG*dG + dB*dB
}

// buildSourceBins groups source pixels into bins and keeps their exact values.
func buildSourceBins(src [][]Pixel, levels int) [][]Pixel {
	bins := make([][]Pixel, levels*levels*levels)
	for y := 0; y < len(src); y++ {
		row := src[y]
		for x := 0; x < len(row); x++ {
			p := row[x]
			bin := quantizePixel(p, levels)
			bins[bin] = append(bins[bin], p)
		}
	}
	return bins
}

// buildTargetHistogram counts how many pixels of the target fall into each bin.
func buildTargetHistogram(target [][]Pixel, levels int) []int {
	hist := make([]int, levels*levels*levels)
	for y := 0; y < len(target); y++ {
		row := target[y]
		for x := 0; x < len(row); x++ {
			bin := quantizePixel(row[x], levels)
			hist[bin]++
		}
	}
	return hist
}

// popPixel removes and returns one pixel from the requested bin; if empty, it finds
// the nearest bin (in quantized color space) that still has supply. Returns false if
// no pixel is available (should not happen when images have identical pixel counts).
func popPixel(bin int, bins [][]Pixel, levels int) (Pixel, bool) {
	if len(bins[bin]) > 0 {
		last := bins[bin][len(bins[bin])-1]
		bins[bin] = bins[bin][:len(bins[bin])-1]
		return last, true
	}

	targetCR, targetCG, targetCB := binCenter(bin, levels)
	bestIdx := -1
	bestDist := int(^uint(0) >> 1) // max int

	for idx, supply := range bins {
		if len(supply) == 0 {
			continue
		}
		cR, cG, cB := binCenter(idx, levels)
		dist := sqDist(targetCR, targetCG, targetCB, cR, cG, cB)
		if dist < bestDist {
			bestDist = dist
			bestIdx = idx
		}
	}

	if bestIdx == -1 {
		return Pixel{}, false
	}

	last := bins[bestIdx][len(bins[bestIdx])-1]
	bins[bestIdx] = bins[bestIdx][:len(bins[bestIdx])-1]
	return last, true
}

// remapPixels rearranges source pixels to match the target color distribution.
// Assumptions: src and target have identical dimensions. No pixel value is changed.
// levels controls the number of bins per channel (e.g., 16 -> 4096 bins).
// Pixels are placed in a randomized order to distribute source pixels uniformly.
func remapPixels(src [][]Pixel, target [][]Pixel, levels int) [][]Pixel {
	if len(src) == 0 || len(target) == 0 || len(src) != len(target) || len(src[0]) != len(target[0]) {
		return nil
	}

	height := len(target)
	width := len(target[0])

	bins := buildSourceBins(src, levels)

	out := make([][]Pixel, height)
	for y := 0; y < height; y++ {
		out[y] = make([]Pixel, width)
	}

	positions := make([][2]int, height*width)
	idx := 0
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			positions[idx] = [2]int{x, y}
			idx++
		}
	}
	rand.Shuffle(len(positions), func(i, j int) {
		positions[i], positions[j] = positions[j], positions[i]
	})

	for _, pos := range positions {
		x, y := pos[0], pos[1]
		bin := quantizePixel(target[y][x], levels)
		p, ok := popPixel(bin, bins, levels)
		if !ok {
			out[y][x] = target[y][x]
			continue
		}
		out[y][x] = p
	}

	return out
}
