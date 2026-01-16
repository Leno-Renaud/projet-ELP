package main

import (
	"fmt"
	"io"
	"net"
	"os"
)

func main() {
	// 1) Demander l'IP du serveur puis construire l'adresse avec le port 9000
	fmt.Println("=== CLIENT - Connexion au serveur ===")
	fmt.Print("Entrez l'adresse IP du serveur: ")
	var ip string
	_, ipErr := fmt.Scanln(&ip)
	if ipErr != nil || len(ip) == 0 {
		fmt.Println("IP invalide!")
		return
	}

	addr := fmt.Sprintf("%s:9000", ip)

	// 2) Se connecter au serveur avec l'adresse construite
	conn, dialErr := net.Dial("tcp", addr)
	if dialErr != nil {
		fmt.Println("Connexion échouée:", dialErr)
		return
	}
	defer conn.Close()

	// 3) Demander le chemin de l'image à traiter
	fmt.Print("Entrez le chemin de l'image: ")
	var imagePath string
	fmt.Scanln(&imagePath)
	if len(imagePath) == 0 {
		fmt.Println("Chemin invalide!")
		return
	}

	// 4) Demander à l'utilisateur quel traitement il veut
	fmt.Println("=== Choix du traitement ===")
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

	// Envoyer le choix (1 byte)
	conn.Write([]byte{byte(choice)})

	// Envoyer l'image
	in, err := os.Open(imagePath)
	if err != nil {
		fmt.Println("Erreur ouverture image:", err)
		return
	}
	defer in.Close()

	io.Copy(conn, in)                // envoie l'image
	conn.(*net.TCPConn).CloseWrite() // signale la fin de l'envoi

	out, _ := os.Create("output/out.jpg")
	defer out.Close()

	io.Copy(out, conn) // reçoit l'image renvoyée

	fmt.Println("Traitement terminé! Résultat sauvegardé dans out.jpg")
}
