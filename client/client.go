package main

import (
	"fmt"
	"io"
	"net"
	"os"
)

func main() {
	// Demander à l'utilisateur quel traitement il veut
	fmt.Println("=== CLIENT - Choix du traitement ===")
	fmt.Println("1. Noir et blanc (BW)")
	fmt.Println("2. Downscale (facteur 4)")
	fmt.Println("3. Remap vers carosse_500x500.jpg")
	fmt.Print("Votre choix (1-3): ")

	var choice int
	_, err := fmt.Scanf("%d", &choice)
	if err != nil || choice < 1 || choice > 3 {
		fmt.Println("Choix invalide!")
		return
	}

	serverIP := "localhost:9000" // Remplace par l'IP du serveur, ex: "192.168.1.10:9000"
	conn, _ := net.Dial("tcp", serverIP)
	defer conn.Close()

	// Envoyer le choix (1 byte)
	conn.Write([]byte{byte(choice)})

	// Envoyer l'image
	in, _ := os.Open("images_sources/asiats_500x500.jpg")
	defer in.Close()

	io.Copy(conn, in)                // envoie l'image
	conn.(*net.TCPConn).CloseWrite() // signale la fin de l'envoi

	out, _ := os.Create("output/out.jpg")
	defer out.Close()

	io.Copy(out, conn) // reçoit l'image renvoyée

	fmt.Println("Traitement terminé! Résultat sauvegardé dans out.jpg")
}
