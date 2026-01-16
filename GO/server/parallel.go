package main

import (
	"image"
	"math/rand"
	"runtime"
	"sync"
)

// extractPixelsParallel convertit une image en matrice de pixels RGB (parallèle)
func extractPixelsParallel(m image.Image, width, height int) [][]Pixel {
	rgbMatrix := make([][]Pixel, height)
	for y := 0; y < height; y++ {
		rgbMatrix[y] = make([]Pixel, width)
	}

	// Guard against empty images
	if height == 0 || width == 0 {
		return rgbMatrix
	}

	// Don't spawn more workers than rows: clamp to height
	numWorkers := runtime.NumCPU()
	if numWorkers > height {
		numWorkers = height
	}

	chunkSize := height / numWorkers
	// Ensure at least one row per worker
	if chunkSize == 0 {
		chunkSize = 1
	}

	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		startY := i * chunkSize
		endY := startY + chunkSize
		// Clamp endY to avoid out of range when chunk distribution doesn't cover all rows
		if endY > height {
			endY = height
		}
		if i == numWorkers-1 {
			endY = height
		}

		wg.Add(1)
		go func(start, end int) {
			defer wg.Done()
			for y := start; y < end; y++ {
				for x := 0; x < width; x++ {
					r, g, b, _ := m.At(x, y).RGBA()
					rgbMatrix[y][x] = Pixel{R: uint16(r), G: uint16(g), B: uint16(b)}
				}
			}
		}(startY, endY)
	}

	wg.Wait()
	return rgbMatrix
}

// blackWhiteParallel convertit la matrice en niveaux de gris (parallèle)
func blackWhiteParallel(rgbMatrix [][]Pixel, width, height int) [][]Pixel {
	// Guard against empty input
	if height == 0 || width == 0 {
		return rgbMatrix
	}

	numGoroutines := runtime.NumCPU()
	if numGoroutines > height {
		numGoroutines = height
	}
	if numGoroutines == 0 {
		return rgbMatrix
	}

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

// downscalePixels réduit la définition sans changer la taille (parallèle)
func downscalePixelsParallel(rgbMatrix [][]Pixel, width, height, factor int) [][]Pixel {
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

	// Diviser le travail par lignes
	if height == 0 || width == 0 {
		return result
	}

	numGoroutines := runtime.NumCPU()
	if numGoroutines > height {
		numGoroutines = height
	}
	if numGoroutines == 0 {
		return result
	}

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

			// Traiter les blocs dans cette tranche de lignes
			for by := startRow; by < endRow; by += factor {
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

					// Calculer la moyenne du bloc
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

					// Remplir le bloc avec la moyenne
					for y := by; y < maxY; y++ {
						for x := bx; x < maxX; x++ {
							result[y][x] = avg
						}
					}
				}
			}
		}(g)
	}

	wg.Wait()
	return result
}

// remapPixels part d'une matrice de pixel source et reconstitue une image target
func remapPixelsParallel(src [][]Pixel, target [][]Pixel, levels int) [][]Pixel {
	if len(src) == 0 || len(target) == 0 || len(src) != len(target) || len(src[0]) != len(target[0]) {
		return nil
	}

	height := len(target)
	width := len(target[0])

	// Supply: group all source pixels by bin.
	bins := buildSourceBins(src, levels)

	// Output image (same dimensions).
	out := make([][]Pixel, height)
	for y := 0; y < height; y++ {
		out[y] = make([]Pixel, width)
	}

	// Create a randomized order of target positions to avoid bias.
	positions := make([][2]int, height*width)
	idx := 0
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			positions[idx] = [2]int{x, y}
			idx++
		}
	}
	rand.Shuffle(len(positions), func(i, j int) { positions[i], positions[j] = positions[j], positions[i] })

	// Feed positions into a buffered channel consumed by workers.
	posCh := make(chan [2]int, len(positions))
	for _, p := range positions {
		posCh <- p
	}
	close(posCh)

	var mu sync.Mutex // protects access to bins and popPixel
	workers := runtime.NumCPU()
	// no need to start more workers than positions
	if workers > len(positions) {
		workers = len(positions)
	}
	if workers == 0 {
		return out
	}
	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for pos := range posCh {
				x, y := pos[0], pos[1]
				bin := quantizePixel(target[y][x], levels)

				mu.Lock()
				p, ok := popPixel(bin, bins, levels)
				mu.Unlock()

				if !ok {
					out[y][x] = target[y][x]
					continue
				}
				out[y][x] = p
			}
		}()
	}

	wg.Wait()
	return out
}
