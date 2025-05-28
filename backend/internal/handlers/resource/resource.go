package resource

import (
	"encoding/json"
	"net/http"

	"backend/internal/database"
)

// Définition de la structure représentant une ressource
type Resource struct {
	ID          int    `json:"id"`          // Identifiant unique de la ressource
	Name        string `json:"name"`        // Nom de la ressource
	Description string `json:"description"` // Description de la ressource
	School      int    `json:"school"`      // Identifiant de l'école associée
	Duration    int    `json:"duration"`    // Durée par défaut pour l'utilisation de cette ressource (en minutes)
	Visible     bool   `json:"visible"`     // Indique si la ressource est visible pour les utilisateurs
}

// Récupère toutes les ressources présentes en base de données
func FetchResources(w http.ResponseWriter, r *http.Request) {
	query := `
	SELECT id, name, description, school, duration, visible
	FROM resource`

	// Exécution de la requête
	rows, err := database.GetConn().Query(query)
	if err != nil {
		http.Error(w, "Échec de l'exécution de la requête : "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var resources []Resource
	// Parcours des résultats
	for rows.Next() {
		var res Resource
		// Mapping des colonnes sur la structure Resource
		if err := rows.Scan(&res.ID, &res.Name, &res.Description, &res.School, &res.Duration, &res.Visible); err != nil {
			http.Error(w, "Échec de l'analyse de la ligne : "+err.Error(), http.StatusInternalServerError)
			return
		}
		resources = append(resources, res)
	}

	// Vérification des erreurs après l'itération
	if err := rows.Err(); err != nil {
		http.Error(w, "Erreur lors de l'itération des lignes : "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Réponse JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resources)
}

// Récupère les ressources associées à une école spécifique (via schoolID)
func FetchResourcesBySchool(w http.ResponseWriter, r *http.Request) {
	schoolID := r.URL.Query().Get("schoolID")

	if schoolID == "" {
		http.Error(w, "schoolID est requis", http.StatusBadRequest)
		return
	}

	query := `
		SELECT id, name, description, school, duration, visible
		FROM resource 
		WHERE school = $1`

	rows, err := database.GetConn().Query(query, schoolID)
	if err != nil {
		http.Error(w, "Échec de l'exécution de la requête : "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var resources []Resource
	// Remplissage de la liste à partir des lignes retournées
	for rows.Next() {
		var res Resource
		if err := rows.Scan(&res.ID, &res.Name, &res.Description, &res.School, &res.Duration, &res.Visible); err != nil {
			http.Error(w, "Échec de l'analyse de la ligne : "+err.Error(), http.StatusInternalServerError)
			return
		}
		resources = append(resources, res)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Erreur lors de l'itération des lignes : "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Réponse JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resources)
}

// Crée une nouvelle ressource dans la base de données
func CreateResource(w http.ResponseWriter, r *http.Request) {
	var resource Resource
	// Décodage du corps de la requête JSON dans la structure Resource
	if err := json.NewDecoder(r.Body).Decode(&resource); err != nil {
		http.Error(w, "Échec du décodage du corps de la requête : "+err.Error(), http.StatusBadRequest)
		return
	}

	queryInsert := `
		INSERT INTO resource (name, description, school, duration, visible)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`

	// Insertion dans la base et récupération de l'ID généré
	err := database.GetConn().QueryRow(
		queryInsert,
		resource.Name,
		resource.Description,
		resource.School,
		resource.Duration,
		resource.Visible,
	).Scan(&resource.ID)
	if err != nil {
		http.Error(w, "Échec de l'exécution de la requête : "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Réponse JSON avec l'objet créé
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resource)
}

// Met à jour une ressource existante en base
func UpdateResource(w http.ResponseWriter, r *http.Request) {
	var resource Resource
	// Décodage du corps JSON
	if err := json.NewDecoder(r.Body).Decode(&resource); err != nil {
		http.Error(w, "Échec du décodage du corps de la requête : "+err.Error(), http.StatusBadRequest)
		return
	}

	query := `
	UPDATE 
	resource 
	SET name = $1, description = $2, school = $3, duration = $4, visible = $5
	WHERE id = $6`

	// Exécution de la requête de mise à jour
	result, err := database.GetConn().Exec(query, resource.Name, resource.Description, resource.School, resource.Duration, resource.Visible, resource.ID)
	if err != nil {
		http.Error(w, "Échec de l'exécution de la requête : "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Récupération du nombre de lignes modifiées
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, "Échec de la récupération du nombre de lignes affectées : "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Réponse JSON indiquant combien de lignes ont été modifiées
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"lignesAffectees": rowsAffected})
}

// Supprime une ressource en fonction de son ID
func DeleteResource(w http.ResponseWriter, r *http.Request) {
	resourceID := r.URL.Query().Get("resourceID")
	if resourceID == "" {
		http.Error(w, "resourceID est requis", http.StatusBadRequest)
		return
	}

	query := `
	DELETE 
	FROM resource 
	WHERE id = $1`

	// Exécution de la suppression
	result, err := database.GetConn().Exec(query, resourceID)
	if err != nil {
		http.Error(w, "Échec de l'exécution de la requête : "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Récupération du nombre de suppressions
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, "Échec de la récupération du nombre de lignes affectées : "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Réponse JSON avec le statut et le nombre de lignes supprimées
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"lignesAffectees": rowsAffected})
}
