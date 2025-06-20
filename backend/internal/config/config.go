package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Définition de la structure Config contenant tous les paramètres nécessaires
// à l'application : informations pour le SMTP, la base de données, et le JWT.
type Config struct {
	SMTPUser     string // Nom d'utilisateur pour le serveur SMTP
	SMTPPassword string // Mot de passe pour le serveur SMTP
	SMTPHost     string // Hôte du serveur SMTP
	SMTPPort     string // Port du serveur SMTP
	SMTPAuth     string // Type d'authentification pour le serveur SMTP (ex: "true" ou "false")

	DBType     string // Type de base de données (ex: mysql, postgres)
	DBUser     string // Nom d'utilisateur pour la base de données
	DBPassword string // Mot de passe pour la base de données
	DBHost     string // Hôte de la base de données
	DBPort     string // Port de la base de données
	DBName     string // Nom de la base de données

	JWTKey []byte // Clé secrète utilisée pour signer les tokens JWT
}

// LoadConfig charge la configuration de l'application à partir
// des variables d'environnement, avec possibilité de lecture via un fichier .env
func LoadConfig() *Config {

	// Tente de charger les variables d'environnement depuis le fichier .env
	err := godotenv.Load(".env")
	if err != nil {
		// Si le fichier .env est introuvable, afficher un avertissement
		// Les variables d'environnement système seront utilisées à la place
		log.Println("Avertissement : Impossible de charger le fichier .env (utilisation des variables d'environnement système)")
	}

	// Lecture du type de base de données (ex: MYSQL, POSTGRES) pour adapter dynamiquement les noms des variables
	var DBType = os.Getenv("DB_TYPE")

	// Création et retour d'une instance de la configuration complète avec les valeurs des variables d'environnement
	return &Config{
		DBType:       DBType,
		SMTPUser:     os.Getenv("SMTP_USER"),     // Lecture de l'utilisateur SMTP
		SMTPPassword: os.Getenv("SMTP_PASSWORD"), // Lecture du mot de passe SMTP
		SMTPHost:     os.Getenv("SMTP_HOST"),     // Lecture de l'hôte SMTP
		SMTPPort:     os.Getenv("SMTP_PORT"),     // Lecture du port SMTP
		SMTPAuth:     os.Getenv("SMTP_AUTH"),     // Lecture de l'authentification SMTP

		// Les noms des variables d'environnement de la base de données sont construits dynamiquement
		// selon le type de base (ex: DB_MYSQL_USER, DB_POSTGRES_USER, etc.)
		DBUser:     os.Getenv("DB_" + strings.ToUpper(DBType) + "_USER"),
		DBPassword: os.Getenv("DB_" + strings.ToUpper(DBType) + "_PASSWORD"),
		DBHost:     os.Getenv("DB_" + strings.ToUpper(DBType) + "_HOST"),
		DBPort:     os.Getenv("DB_" + strings.ToUpper(DBType) + "_PORT"),
		DBName:     os.Getenv("DB_" + strings.ToUpper(DBType) + "_NAME"),

		// Lecture de la clé JWT, convertie en slice d'octets
		JWTKey: []byte(os.Getenv("JWT_KEY")),
	}
}
