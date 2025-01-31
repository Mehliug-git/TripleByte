package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

const serviceName = "pd"

type myService struct{}

// Boucle du service Windows
func (m *myService) Execute(args []string, r <-chan svc.ChangeRequest, s chan<- svc.Status) (svcSpecificEC bool, exitCode uint32) {
	s <- svc.Status{State: svc.StartPending}
	s <- svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown}

	for {
		select {
		case c := <-r:
			switch c.Cmd {
			case svc.Stop, svc.Shutdown:
				return false, 0
			}
		default:

			// Attendre 30 secondes avant de recommencer
			time.Sleep(30 * time.Second)
		}
	}
}

// Vérifier si le service existe déjà
func serviceExists(serviceName string) bool {
	m, err := mgr.Connect()
	if err != nil {
		log.Println("Impossible de se connecter au gestionnaire de services:", err)
		return false
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err != nil {
		return false // Le service n'existe pas
	}
	s.Close()
	return true // Le service existe
}

// Installer et démarrer le service
func installAndStartService(serviceName, exePath string) error {
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("échec de la connexion au gestionnaire de services: %v", err)
	}
	defer m.Disconnect()

	s, err := m.CreateService(serviceName, exePath, mgr.Config{
		DisplayName: serviceName,
		StartType:   mgr.StartAutomatic, // Démarrage automatique
	})
	if err != nil {
		return fmt.Errorf("échec de la création du service: %v", err)
	}
	defer s.Close()

	err = s.Start()
	if err != nil {
		return fmt.Errorf("échec du démarrage du service: %v", err)
	}

	fmt.Println("Service installé et démarré avec succès")
	return nil
}

func main() {
	// Vérifie si le programme est un service Windows
	isService, err := svc.IsWindowsService()
	if err != nil {
		log.Fatalf("Erreur lors de la détection du service: %v", err)
	}

	if isService {
		// Lancer le service Windows
		err = svc.Run(serviceName, &myService{})
		if err != nil {
			log.Fatalf("Échec de l'exécution du service: %v", err)
		}
		return
	}

	// Si le programme est exécuté normalement, il installe automatiquement le service
	exePath, err := os.Executable()
	if err != nil {
		log.Fatalf("Impossible de récupérer le chemin de l'exécutable: %v", err)
	}

	if !serviceExists(serviceName) {
		log.Println("Installation automatique du service en cours...")
		err = installAndStartService(serviceName, exePath)
		if err != nil {
			log.Fatalf("Échec de l'installation du service: %v", err)
		}
		log.Println("Le service est maintenant installé et en cours d'exécution.")
	} else {
		log.Println("Le service est déjà installé et en cours d'exécution.")
	}
}
