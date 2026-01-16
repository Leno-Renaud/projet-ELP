package main

import (
	"fmt"
	"image"
	"time"
)

// CompareExtractPixels compare les versions séquentielle et parallèle de extractPixels
func CompareExtractPixels(img image.Image) {
	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y

	fmt.Println("=== TEST extractPixels (SÉQUENTIEL) ===")
	start1 := time.Now()
	rgbMatrix1 := extractPixels(img, width, height)
	duration1 := time.Since(start1)
	fmt.Printf("Temps : %v\n\n", duration1)

	fmt.Println("=== TEST extractPixelsParallel (PARALLÈLE) ===")
	start2 := time.Now()
	rgbMatrix2 := extractPixelsParallel(img, width, height)
	duration2 := time.Since(start2)
	fmt.Printf("Temps : %v\n\n", duration2)

	fmt.Println("=== COMPARAISON ===")
	speedup := float64(duration1) / float64(duration2)
	savings := (1 - float64(duration2)/float64(duration1)) * 100
	fmt.Printf("Speedup : %.2fx\n", speedup)
	fmt.Printf("Gain de temps : %.2f%%\n", savings)
	fmt.Printf("Différence : %v\n\n", duration1-duration2)

	_ = rgbMatrix1 // éviter les avertissements "unused"
	_ = rgbMatrix2
}

// CompareBlackWhite compare les versions séquentielle et parallèle de blackWhite
func CompareBlackWhite(rgbMatrix [][]Pixel, width, height int) {
	fmt.Println("=== TEST blackWhite (SÉQUENTIEL) ===")
	start1 := time.Now()
	rgbMatrix1 := blackWhite(rgbMatrix, width, height)
	duration1 := time.Since(start1)
	fmt.Printf("Temps : %v\n\n", duration1)

	fmt.Println("=== TEST blackWhiteParallel (PARALLÈLE) ===")
	start2 := time.Now()
	rgbMatrix2 := blackWhiteParallel(rgbMatrix, width, height)
	duration2 := time.Since(start2)
	fmt.Printf("Temps : %v\n\n", duration2)

	fmt.Println("=== COMPARAISON ===")
	speedup := float64(duration1) / float64(duration2)
	savings := (1 - float64(duration2)/float64(duration1)) * 100
	fmt.Printf("Speedup : %.2fx\n", speedup)
	fmt.Printf("Gain de temps : %.2f%%\n", savings)
	fmt.Printf("Différence : %v\n\n", duration1-duration2)

	_ = rgbMatrix1
	_ = rgbMatrix2
}

func CompareDownscalePixels(rgbMatrix [][]Pixel, width, height int) {
	fmt.Println("=== TEST downscalePixels (SÉQUENTIEL) ===")
	start1 := time.Now()
	rgbMatrix1 := downscalePixels(rgbMatrix, width, height, 2)
	duration1 := time.Since(start1)
	fmt.Printf("Temps : %v\n\n", duration1)

	fmt.Println("=== TEST downscalePixelsParallel (PARALLÈLE) ===")
	start2 := time.Now()
	rgbMatrix2 := downscalePixelsParallel(rgbMatrix, width, height, 2)
	duration2 := time.Since(start2)
	fmt.Printf("Temps : %v\n\n", duration2)

	fmt.Println("=== COMPARAISON ===")
	speedup := float64(duration1) / float64(duration2)
	savings := (1 - float64(duration2)/float64(duration1)) * 100
	fmt.Printf("Speedup : %.2fx\n", speedup)
	fmt.Printf("Gain de temps : %.2f%%\n", savings)
	fmt.Printf("Différence : %v\n\n", duration1-duration2)

	_ = rgbMatrix1
	_ = rgbMatrix2
}

func CompareRemapPixels(sourceMatrix [][]Pixel, destinationMatrix [][]Pixel, levels int) {
	fmt.Println("=== TEST remapPixels (SÉQUENTIEL) ===")
	start1 := time.Now()
	rgbMatrix1 := remapPixels(sourceMatrix, destinationMatrix, levels)
	duration1 := time.Since(start1)
	fmt.Printf("Temps : %v\n\n", duration1)

	fmt.Println("=== TEST remapPixels (PARALLÈLE) ===")
	start2 := time.Now()
	rgbMatrix2 := remapPixelsParallel(sourceMatrix, destinationMatrix, levels)
	duration2 := time.Since(start2)
	fmt.Printf("Temps : %v\n\n", duration2)

	fmt.Println("=== COMPARAISON ===")
	speedup := float64(duration1) / float64(duration2)
	savings := (1 - float64(duration2)/float64(duration1)) * 100
	fmt.Printf("Speedup : %.2fx\n", speedup)
	fmt.Printf("Gain de temps : %.2f%%\n", savings)
	fmt.Printf("Différence : %v\n\n", duration1-duration2)

	_ = rgbMatrix1
	_ = rgbMatrix2
}
