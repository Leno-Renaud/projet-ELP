package main

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"io"
	"net"
)

func main() {
	ln, _ := net.Listen("tcp", ":9000")
	fmt.Println("Serveur démarré sur :9000")

	// Charger l'image cible pour le remap
	targetImg := loadImage("carosse_500x500.jpg")
	targetBounds := targetImg.Bounds()
	targetW, targetH := targetBounds.Max.X, targetBounds.Max.Y
	targetMatrix := extractPixels(targetImg, targetW, targetH)

	for {
		conn, _ := ln.Accept()
		go func(c net.Conn) {
			defer c.Close()

			// Lire toutes les données
			buf, err := io.ReadAll(c)
			if err != nil {
				fmt.Println("Erreur lecture:", err)
				return
			}

			// Le premier byte est le choix, le reste est l'image
			if len(buf) < 2 {
				fmt.Println("Données insuffisantes")
				return
			}
			choice := buf[0]
			imgData := buf[1:]

			// Décoder l'image
			img, _, err := image.Decode(bytes.NewReader(imgData))
			if err != nil {
				fmt.Println("Erreur décodage image:", err)
				return
			}
			b := img.Bounds()
			w, h := b.Max.X, b.Max.Y

			// 3. Appliquer le traitement selon le choix
			var result [][]Pixel
			switch choice {
			case 1: // BW
				fmt.Println("Traitement: Noir et blanc")
				result = blackWhiteParallel(extractPixelsParallel(img, w, h), w, h)
			case 2: // Downscale
				fmt.Println("Traitement: Downscale (facteur 4)")
				pixels := extractPixelsParallel(img, w, h)
				result = downscalePixelsParallel(pixels, w, h, 4)
			case 3: // Remap
				fmt.Println("Traitement: Remap vers carosse_500x500.jpg")
				if w != targetW || h != targetH {
					fmt.Printf("Dimensions incompatibles (%dx%d vs %dx%d)\n", w, h, targetW, targetH)
					return
				}
				srcMatrix := extractPixelsParallel(img, w, h)
				result = remapPixelsParallel(srcMatrix, targetMatrix, 16)
			default:
				fmt.Println("Choix invalide:", choice)
				return
			}

			// 4. Encoder et envoyer le résultat
			jpeg.Encode(c, pixelsToImage(result), nil)
			fmt.Println("Requête complétée")
		}(conn)
	}
}
