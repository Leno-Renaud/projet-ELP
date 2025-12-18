package main

import (
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg" // Indispensable pour décoder le JPEG (init function)
	"image/png"
	_ "image/png" // Indispensable pour décoder le PNG (init function)
	"log"
	"os"
	"runtime"
	"sync"
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

// extractPixels convertit une image en matrice de pixels RGB
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

func blackWhite(rgbMatrix [][]Pixel, width, height int) [][]Pixel {
	numGoroutines := runtime.NumCPU()
	rowsPerGoroutine := (height + numGoroutines - 1) / numGoroutines

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for g := 0; g < numGoroutines; g++ {
		go func(goroutineID int) {
			defer wg.Done()

			startRow := goroutineID * rowsPerGoroutine
			endRow := (goroutineID + 1) * rowsPerGoroutine
			if endRow > height {
				endRow = height
			}

			for y := startRow; y < endRow; y++ {
				for x := 0; x < width; x++ {
					p := rgbMatrix[y][x]
					gray := uint16(0.299*float64(p.R) + 0.587*float64(p.G) + 0.114*float64(p.B))
					rgbMatrix[y][x] = Pixel{R: gray, G: gray, B: gray}
				}
			}
		}(g)
	}

	wg.Wait()
	return rgbMatrix
}

// pixelsToImage convertit une matrice de pixels en image RGBA
func pixelsToImage(rgbMatrix [][]Pixel) *image.RGBA {
	if len(rgbMatrix) == 0 || len(rgbMatrix[0]) == 0 {
		// Return an empty image if the matrix is empty to avoid panics
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

func main() {
	// Charger l'image
	m := loadImage("image.jpg")
	// Récupérer les dimensions
	bounds := m.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y
	fmt.Printf("Dimensions : %dx%d\n", width, height)
	// Extraire les pixels
	rgbMatrix := extractPixels(m, width, height)
	//rgbMatrix = blackWhite(rgbMatrix, width, height)
	rgbMatrix = downscalePixels(rgbMatrix, width, height, 4)
	// Convertir en image RGBA
	out := pixelsToImage(rgbMatrix)
	// Sauvegarder
	saveImage(out, "out2.png")
}

// downscalePixels réduit la définition sans changer la taille :
// pour chaque bloc `factor×factor` on calcule la couleur moyenne et on remplit
// tout le bloc avec cette couleur (pixelisation)
func downscalePixels(rgbMatrix [][]Pixel, width, height, factor int) [][]Pixel {
	if factor <= 1 {
		return rgbMatrix
	}
	if len(rgbMatrix) == 0 || len(rgbMatrix[0]) == 0 {
		return rgbMatrix
	}

	// Résultat a la même taille que l'entrée
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
