package database

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"backend/internal/config"

	_ "github.com/lib/pq" // Import nécessaire pour le driver PostgreSQL
)

var conn *sql.DB

// GetConn retourne la connexion active à la base de données.
func GetConn() *sql.DB {
	if conn == nil {
		log.Fatal("La connexion à la base de données n'est pas initialisée.")
	}
	return conn
}

// InitDatabase initialise la connexion à la base de données
// en utilisant la configuration fournie en paramètre.
func InitDatabase(conf *config.Config) {
	dsn, err := buildDSN(conf)
	if err != nil {
		log.Fatalf("Erreur lors de la génération du DSN : %v", err)
	}

	// Connexion à la base de données
	db, err := sql.Open(strings.ToLower(conf.DBType), dsn)
	if err != nil {
		log.Fatalf("Erreur lors de l'ouverture de la base de données : %v", err)
	}

	// Gestion de l'erreur éventuelle
	if err = db.Ping(); err != nil {
		log.Fatalf("Impossible de se connecter à la base de données : %v", err)
	}

	conn = db
}

// buildDSN génère la chaîne de connexion selon le type de base de données.
// contenu dans le fichier .env
func buildDSN(conf *config.Config) (string, error) {
	switch strings.ToUpper(conf.DBType) {
	case "POSTGRES":
		return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			conf.DBUser, conf.DBPassword, conf.DBHost, conf.DBPort, conf.DBName), nil

	case "MYSQL":
		return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
			conf.DBUser, conf.DBPassword, conf.DBHost, conf.DBPort, conf.DBName), nil

	default:
		return "", fmt.Errorf("type de base de données non supporté : %s", conf.DBType)
	}
}
