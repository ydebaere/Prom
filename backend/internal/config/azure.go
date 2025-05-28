package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Définition d'une structure Azure qui contient les informations de configuration
// nécessaires pour l'authentification : TenantID, ClientID et Secret.
type Azure struct {
	TenantID string `json:"tenant_id"` // Identifiant du tenant Azure
	ClientID string `json:"client_id"` // Identifiant du client (application)
	Secret   string `json:"secret"`    // Secret du client (mot de passe)
}

// LoadAzureConfig charge la configuration Azure à partir des variables d'environnement.
// Ces variables peuvent être définies dans un fichier .env ou directement dans le système.
func LoadAzureConfig() *Azure {

	// Tente de charger les variables d'environnement depuis le fichier .env
	err := godotenv.Load(".env")
	if err != nil {
		// Si le fichier .env est introuvable, afficher un avertissement
		// Les variables d'environnement système seront utilisées à la place
		log.Println("Avertissement : Impossible de charger le fichier .env (utilisation des variables d'environnement système)")
	}

	// Retourne une instance de la structure Azure avec les valeurs des variables d'environnement
	return &Azure{
		TenantID: os.Getenv("CAMPUS_ICALE_AUTH_BACKEND_MT_TENANTID"), // Lecture de la variable TENANTID
		ClientID: os.Getenv("CAMPUS_ICALE_AUTH_BACKEND_MT_CLIENTID"), // Lecture de la variable CLIENTID
		Secret:   os.Getenv("CAMPUS_ICALE_AUTH_BACKEND_MT_SECRET"),   // Lecture de la variable SECRET
	}
}
