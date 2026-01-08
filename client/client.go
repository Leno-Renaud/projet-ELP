package main

import (
	"io"
	"net"
	"os"
)

func main() {
	serverIP := "10.87.174.195:9000" // Remplace par l'IP du serveur, ex: "192.168.1.10:9000"
	conn, _ := net.Dial("tcp", serverIP)
	defer conn.Close()

	in, _ := os.Open("demo-project/images_sources/asiats_500x500.jpg")
	defer in.Close()

	io.Copy(conn, in)                // envoie l'image
	conn.(*net.TCPConn).CloseWrite() // signale la fin de l'envoi

	out, _ := os.Create("out.jpg")
	defer out.Close()

	io.Copy(out, conn) // reçoit l'image renvoyée
}
