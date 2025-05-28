package main

import (
	"fmt"
	"log"
	"net/http"

	"backend/internal/config"
	"backend/internal/database"
	"backend/internal/middleware"
	"backend/internal/routes"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

func main() {
	// Chargement des paramètres de configuration
	cfg := config.LoadConfig()

	// Initialisation de la base de données
	database.InitDatabase(cfg)

	// Enregistrer les routes
	routes.Routes()

	// Configuration des CORS
	handlerWithCors := middleware.SetupCORS(http.DefaultServeMux)

	// Démarrage du serveur
	fmt.Println("Serveur démarré sur le port 3004")
	fmt.Println("http://localhost:3004")
	log.Fatal(http.ListenAndServe(":3004", handlerWithCors))
}
