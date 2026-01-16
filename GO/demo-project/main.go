package main

import (
	"fmt"
	"runtime"
)

func main() {
	// Charger l'image
	image := loadImage("images_sources/asiats_500x500.jpg")
	image2 := loadImage("images_sources/carosse_500x500.jpg")
	// Récupérer les dimensions
	bounds := image.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y

	bounds2 := image2.Bounds()
	width2, height2 := bounds2.Max.X, bounds2.Max.Y

	rgbMatrix := extractPixels(image, width, height)
	rgbMatrix2 := extractPixels(image2, width2, height2)

	fmt.Printf("Dimensions : %dx%d\n", width, height)
	fmt.Printf("Nombre de cœurs : %d\n\n", runtime.NumCPU())

	// ============================================
	// Test Fonctions (SÉQUENTIEL vs PARALLÈLE)
	// ============================================
	CompareExtractPixels(image)
	CompareBlackWhite(copyMatrix(rgbMatrix), width, height)
	CompareDownscalePixels(copyMatrix(rgbMatrix), width, height)
	CompareRemapPixels(copyMatrix(rgbMatrix), rgbMatrix2, 16)
	// ============================================
	// Traitements
	// ============================================
	traitementBW := blackWhite(copyMatrix(rgbMatrix), width, height)
	traitementDownscale := downscalePixelsParallel(copyMatrix(rgbMatrix), width, height, 2)
	traitementRemap := remapPixelsParallel(copyMatrix(rgbMatrix), rgbMatrix2, 16)

	imageTraitementBW := pixelsToImage(traitementBW)
	imageTraitementDownscale := pixelsToImage(traitementDownscale)
	imageTraitementRemap := pixelsToImage(traitementRemap)

	saveImage(imageTraitementBW, "output/blackWhite.jpg")
	saveImage(imageTraitementDownscale, "output/downscale.jpg")
	saveImage(imageTraitementRemap, "output/remap.jpg")
	fmt.Println("=== Traitements effectués avec succès ===")
}
