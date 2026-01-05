package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"time"
)

func main() {
	// Charger l'image
	image := loadImage("asiats.jpg")
	// Récupérer les dimensions
	bounds := image.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y
	fmt.Printf("Dimensions : %dx%d\n", width, height)
	fmt.Printf("Nombre de cœurs : %d\n\n", runtime.NumCPU())

	// ============================================
	// Test 1 & 2 : extractPixels (SÉQUENTIEL vs PARALLÈLE)
	// ============================================
	CompareExtractPixels(image)
	// ============================================
	// Récupérer les matrices
	// ============================================
	rgbMatrix1 := extractPixels(image, width, height)
	rgbMatrix2 := extractPixelsParallel(image, width, height)

	// ============================================
	// Test blackWhite (SÉQUENTIEL vs PARALLÈLE)
	// ============================================
	CompareBlackWhite(rgbMatrix1, width, height)

	// ============================================
	// Remap de pixels (source -> cible) si une cible est disponible
	// ============================================
	targetPath := "target.jpg"
	if _, err := os.Stat(targetPath); err == nil {
		fmt.Printf("=== REMAP vers %s (sans changer les pixels, seulement leur position) ===\n", targetPath)
		timg := loadImage(targetPath)
		tb := timg.Bounds()
		if tb.Max.X != width || tb.Max.Y != height {
			fmt.Printf("Dimensions différentes (%dx%d vs %dx%d), remap ignoré.\n\n", tb.Max.X, tb.Max.Y, width, height)
		} else {
			srcMatrix := extractPixels(image, width, height)
			tgtMatrix := extractPixels(timg, width, height)
			startRemap := time.Now()
			remapped := remapPixels(srcMatrix, tgtMatrix, 16) // 16 niveaux par canal → 4096 bins
			fmt.Printf("Remap terminé en %v\n", time.Since(startRemap))
			outRemap := pixelsToImage(remapped)
			outName := "remap.png"
			saveImage(outRemap, outName)
			fmt.Printf("Image remappée : %s\n\n", outName)
		}
	} else {
		fmt.Println("Aucune cible target.jpg trouvée, remap ignoré.")
	}

	// Utiliser la version parallèle pour le résultat final
	rgbMatrix := rgbMatrix2
	_ = rgbMatrix1 // éviter l'avertissement "unused"
	rgbMatrix = blackWhiteParallel(rgbMatrix, width, height)
	// Re-extraction (conserver le flux exact de l'ancien main)
	rgbMatrix = extractPixels(image, width, height)
	// Pixelisation
	rgbMatrix = downscalePixels(rgbMatrix, width, height, 4)

	// Convertir en image RGBA & sauvegarder
	out := pixelsToImage(rgbMatrix)
	saveImage(out, "out.png")
	fmt.Println("Image sauvegardée : out.png")

	// ============================================
	// Effet "Halo" : transformer une image source pour qu'elle ressemble à une image cible
	// ============================================
	fmt.Println("\n=== EFFET HALO ===")
	srcImg := loadImage("test101.png")
	tgtImg := loadImage("test100.jpg")

	// Réutiliser les variables `width` et `height` déjà déclarées plus haut
	width = srcImg.Bounds().Dx()
	height = srcImg.Bounds().Dy()

	// ⚠️ Assure-toi que les deux images ont la même taille
	tgtWidth := tgtImg.Bounds().Dx()
	tgtHeight := tgtImg.Bounds().Dy()

	if width != tgtWidth || height != tgtHeight {
		log.Fatal("Les images source et cible doivent avoir la même taille")
	}

	srcPixels := extractPixels(srcImg, width, height)
	tgtPixels := extractPixels(tgtImg, width, height)

	factor := 8

	result := transformToTarget(
		srcPixels,
		tgtPixels,
		width,
		height,
		factor,
	)

	out = pixelsToImage(result)
	saveImage(out, "result.png")
}
