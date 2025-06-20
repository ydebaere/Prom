package userSchoolResource

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"backend/internal/database"
)

func FetchUserSchoolResource(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("DEBUG: FetchUserResourceSchool without params\n")
	query := `
	SELECT * 
	FROM user_school_resource`

	rows, err := database.GetConn().Query(query)
	if err != nil {
		http.Error(w, "Failed to execute query: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		http.Error(w, "Failed to get columns: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var userResources []map[string]interface{}
	userCache := make(map[string]string)
	addedUserIDs := make(map[string]bool)

	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			http.Error(w, "Failed to scan row: "+err.Error(), http.StatusInternalServerError)
			return
		}

		userResource := make(map[string]interface{})
		var user_id string

		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				userResource[col] = string(b)
			} else {
				userResource[col] = val
			}
			if col == "user_id" {
				user_id = fmt.Sprintf("%v", val)
				if _, found := addedUserIDs[user_id]; found {
					break
				}
				if userName, found := userCache[user_id]; found {
					userResource["user_name"] = userName
				} else {
					var given_name, family_name string
					err := database.GetConn().QueryRow(`
					SELECT given_name, family_name 
					FROM usr 
					WHERE id = $1`, user_id).Scan(&given_name, &family_name)
					if err != nil {
						http.Error(w, "Failed to get user: "+err.Error(), http.StatusInternalServerError)
						return
					}
					userName := given_name + " " + family_name
					userCache[user_id] = userName
					userResource["user_name"] = userName
				}
			}
			if col == "resource_id" {
				var resource_name, resource_description string
				err := database.GetConn().QueryRow(`
				SELECT name, description 
				FROM resource 
				WHERE id = $1`, val).Scan(&resource_name, &resource_description)
				if err != nil {
					http.Error(w, "Failed to get resource: "+err.Error(), http.StatusInternalServerError)
					return
				}
				userResource["resource_name"] = resource_name
				userResource["resource_description"] = resource_description
			}
			if col == "school_id" {
				var school_name string
				err := database.GetConn().QueryRow(`
				SELECT name 
				FROM school 
				WHERE id = $1`, val).Scan(&school_name)
				if err != nil {
					http.Error(w, "Failed to get school: "+err.Error(), http.StatusInternalServerError)
					return
				}
				userResource["school_name"] = school_name
			}
		}
		userResources = append(userResources, userResource)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userResources)
}

func FetchUserSchoolResourceWithSchoolID(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("DEBUG: FetchUserSchoolResource schoolID: %s\n", r.URL.Query().Get("schoolID"))
	s := r.URL.Query().Get("schoolID")
	schoolID, err := strconv.Atoi(s)
	if err != nil {
		http.Error(w, "Invalid schoolID: "+err.Error(), http.StatusBadRequest)
		return
	}

	query := `
    SELECT usr.id AS user_id, usr.given_name, usr.family_name, usr.unique_name,
    r.id AS resource_id, r.name AS resource_name, r.description AS resource_description, r.duration AS resource_duration
    FROM user_school_resource usr_sch_res
    JOIN usr ON usr_sch_res.user_id = usr.id
    JOIN resource r ON usr_sch_res.resource_id = r.id
    WHERE usr_sch_res.school_id = $1`

	rows, err := database.GetConn().Query(query, schoolID)
	if err != nil {
		http.Error(w, "Failed to execute query: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Structure pour stocker les ressources et leurs utilisateurs
	resources := make(map[int]map[string]interface{})

	for rows.Next() {
		var userID, resourceID, resourceDuration int
		var given_name, family_name, resourceName, resourceDescription, unique_name string

		if err := rows.Scan(
			&userID, &given_name, &family_name, &unique_name,
			&resourceID, &resourceName, &resourceDescription,
			&resourceDuration); err != nil {
			http.Error(w, "Failed to scan row: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Si la ressource n'existe pas encore dans la map, on l'ajoute
		if _, exists := resources[resourceID]; !exists {
			resources[resourceID] = map[string]interface{}{
				"resource_id":          resourceID,                 // ID de la ressource
				"resource_name":        resourceName,               // Nom de la ressource
				"resource_description": resourceDescription,        // Description de la ressource
				"resource_duration":    resourceDuration,           // Durée de la ressource
				"users":                []map[string]interface{}{}, // Liste des utilisateurs
				"userSet":              make(map[int]bool),         // Set pour éviter les doublons
			}
		}

		// Récupérer la map de la ressource actuelle
		resource := resources[resourceID]

		// Ajouter l'utilisateur à la liste des utilisateurs de la ressource si non déjà ajouté
		userSet := resource["userSet"].(map[int]bool)
		if !userSet[userID] {
			user := map[string]interface{}{
				"user_id":     userID,      // ID de l'utilisateur
				"given_name":  given_name,  // Prénom de l'utilisateur
				"family_name": family_name, // Nom de famille de l'utilisateur
				"unique_name": unique_name, // Nom unique de l'utilisateur
			}
			resource["users"] = append(resource["users"].([]map[string]interface{}), user)
			userSet[userID] = true // Marquer cet utilisateur comme ajouté pour cette ressource
		}
	}

	// Supprimer les maps `userSet` avant de convertir en slice
	var result []map[string]interface{}
	for _, resource := range resources {
		delete(resource, "userSet") // Supprimer la map temporaire
		result = append(result, resource)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
func FetchUserSchoolResourceWithUserID(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("DEBUG: FetchUserResourceSchool userID: %s\n", r.URL.Query().Get("userID"))
	userID := r.URL.Query().Get("userID")
	userQuery := `
	SELECT id
	FROM usr
	WHERE id = $1`

	var id string
	err := database.GetConn().QueryRow(userQuery, userID).Scan(&id)
	if err != nil {
		http.Error(w, "Failed to get user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	query := `
	SELECT * 
	FROM user_school_resource
	WHERE user_id = $1`

	rows, err := database.GetConn().Query(query, id)
	if err != nil {
		http.Error(w, "Failed to execute query: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		http.Error(w, "Failed to get columns: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var userResources []map[string]interface{}
	userCache := make(map[string]string)
	addedUserIDs := make(map[string]bool)

	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			http.Error(w, "Failed to scan row: "+err.Error(), http.StatusInternalServerError)
			return
		}

		userResource := make(map[string]interface{})
		var user_id string

		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				userResource[col] = string(b)
			} else {
				userResource[col] = val
			}
			if col == "user_id" {
				user_id = fmt.Sprintf("%v", val)
				if _, found := addedUserIDs[user_id]; found {
					break
				}
				if userName, found := userCache[user_id]; found {
					userResource["user_name"] = userName
				} else {
					var given_name, family_name string
					err := database.GetConn().QueryRow(`
					SELECT given_name, family_name 
					FROM usr 
					WHERE id = $1`, user_id).Scan(&given_name, &family_name)
					if err != nil {
						http.Error(w, "Failed to get user: "+err.Error(), http.StatusInternalServerError)
						return
					}
					userName := given_name + " " + family_name
					userCache[user_id] = userName
					userResource["user_name"] = userName
				}
			}
			if col == "resource_id" {
				var resource_name, resource_description string
				err := database.GetConn().QueryRow(`
				SELECT name, description 
				FROM resource 
				WHERE id = $1`, val).Scan(&resource_name, &resource_description)
				if err != nil {
					http.Error(w, "Failed to get resource: "+err.Error(), http.StatusInternalServerError)
					return
				}
				userResource["resource_name"] = resource_name
				userResource["resource_description"] = resource_description
			}
			if col == "school_id" {
				var school_name string
				err := database.GetConn().QueryRow(`
				SELECT name 
				FROM school 
				WHERE id = $1`, val).Scan(&school_name)
				if err != nil {
					http.Error(w, "Failed to get school: "+err.Error(), http.StatusInternalServerError)
					return
				}
				userResource["school_name"] = school_name
			}
		}
		userResources = append(userResources, userResource)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userResources)
}
func FetchUserSchoolResourceWithResourceID(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("DEBUG: FetchUserResourceSchool resourceID: %s\n", r.URL.Query().Get("resourceID"))
	resourceID := r.URL.Query().Get("resourceID")

	query := `
	SELECT * 
	FROM user_school_resource
	WHERE resource_id = $1`

	rows, err := database.GetConn().Query(query, resourceID)
	if err != nil {
		http.Error(w, "Failed to execute query: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		http.Error(w, "Failed to get columns: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var userResources []map[string]interface{}
	userCache := make(map[string]string)
	addedUserIDs := make(map[string]bool)

	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			http.Error(w, "Failed to scan row: "+err.Error(), http.StatusInternalServerError)
			return
		}

		userResource := make(map[string]interface{})
		var user_id string

		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				userResource[col] = string(b)
			} else {
				userResource[col] = val
			}
			if col == "user_id" {
				user_id = fmt.Sprintf("%v", val)
				if _, found := addedUserIDs[user_id]; found {
					break
				}
				if userName, found := userCache[user_id]; found {
					userResource["user_name"] = userName
				} else {
					var given_name, family_name string
					err := database.GetConn().QueryRow(`
					SELECT given_name, family_name 
					FROM usr 
					WHERE id = $1`, user_id).Scan(&given_name, &family_name)
					if err != nil {
						http.Error(w, "Failed to get user: "+err.Error(), http.StatusInternalServerError)
						return
					}
					userName := given_name + " " + family_name
					userCache[user_id] = userName
					userResource["user_name"] = userName
				}
			}
			if col == "resource_id" {
				var resource_name, resource_description string
				err := database.GetConn().QueryRow(`
				SELECT name, description 
				FROM resource 
				WHERE id = $1`, val).Scan(&resource_name, &resource_description)
				if err != nil {
					http.Error(w, "Failed to get resource: "+err.Error(), http.StatusInternalServerError)
					return
				}
				userResource["resource_name"] = resource_name
				userResource["resource_description"] = resource_description
			}
			if col == "school_id" {
				var school_name string
				err := database.GetConn().QueryRow(`
				SELECT name 
				FROM school 
				WHERE id = $1`, val).Scan(&school_name)
				if err != nil {
					http.Error(w, "Failed to get school: "+err.Error(), http.StatusInternalServerError)
					return
				}
				userResource["school_name"] = school_name
			}
		}
		userResources = append(userResources, userResource)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userResources)
}
func FetchUserFromUSR(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("DEBUG: FetchUserFromUSR\n")
	// Récupérer les paramètres
	schoolStr := r.URL.Query().Get("schoolID")
	resourceStr := r.URL.Query().Get("resourceID")

	schoolID, err := strconv.Atoi(schoolStr)
	if err != nil {
		http.Error(w, "Invalid schoolID: "+err.Error(), http.StatusBadRequest)
		return
	}

	resourceID, err := strconv.Atoi(resourceStr)
	if err != nil {
		http.Error(w, "Invalid resourceID: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Requête pour obtenir les utilisateurs uniques
	query := `
	SELECT DISTINCT ON (u.id)
    u.id,
    u.unique_name,
    u.given_name,
    u.family_name
	FROM user_school_resource u_s_r
			JOIN usr u ON u_s_r.user_id = u.id
	WHERE u_s_r.school_id = $1 AND u_s_r.resource_id = $2
	`

	rows, err := database.GetConn().Query(query, schoolID, resourceID)
	if err != nil {
		http.Error(w, "Failed to execute query: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []map[string]interface{}

	for rows.Next() {
		var id int
		var uniqueName, givenName, familyName string

		if err := rows.Scan(&id, &uniqueName, &givenName, &familyName); err != nil {
			http.Error(w, "Failed to scan row: "+err.Error(), http.StatusInternalServerError)
			return
		}

		user := map[string]interface{}{
			"user_id":     id,         // ID de l'utilisateur
			"unique_name": uniqueName, // Nom unique de l'utilisateur
			"given_name":  givenName,  // Prénom de l'utilisateur
			"family_name": familyName, // Nom de famille de l'utilisateur
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Error iterating rows: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}
func FetchAllRessourcesFromUSR(w http.ResponseWriter, r *http.Request) {
	s := r.URL.Query().Get("schoolID")

	schoolID, err := strconv.Atoi(s)
	if err != nil {
		http.Error(w, "Invalid schoolID: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Récupérer le nom de l’école
	var schoolName string
	err = database.GetConn().QueryRow(`
		SELECT name FROM school WHERE id = $1`, schoolID).Scan(&schoolName)
	if err != nil {
		http.Error(w, "Failed to get school: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Requête pour récupérer les ressources sans doublons
	query := `
	SELECT *
	FROM user_school_resource
	WHERE usr.school_id = $1`

	rows, err := database.GetConn().Query(query, schoolID)
	if err != nil {
		http.Error(w, "Failed to execute query: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var resources []map[string]interface{}

	for rows.Next() {
		var resource_id, user_id, duration int
		var resourceName, resourceDescription, agentGivenName, agentFamilyName string

		if err := rows.Scan(&resource_id, &resourceName, &resourceDescription, &duration); err != nil {
			http.Error(w, "Failed to scan row: "+err.Error(), http.StatusInternalServerError)
			return
		}

		resource := map[string]interface{}{
			"resource_id":          resource_id,
			"resource_name":        resourceName,
			"resource_description": resourceDescription,
			"resource_duration":    duration,
			"school_id":            schoolID,
			"school_name":          schoolName,
			"agent_id":             user_id,
			"agent_given_name":     agentGivenName,
			"agent_family_name":    agentFamilyName,
		}

		resources = append(resources, resource)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Error iterating rows: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resources)
}
func FetchResourceFromUSR(w http.ResponseWriter, r *http.Request) {
	s := r.URL.Query().Get("schoolID")

	schoolID, err := strconv.Atoi(s)
	if err != nil {
		http.Error(w, "Invalid schoolID: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Récupérer le nom de l’école
	var schoolName string
	err = database.GetConn().QueryRow(`
		SELECT name FROM school WHERE id = $1`, schoolID).Scan(&schoolName)
	if err != nil {
		http.Error(w, "Failed to get school: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Requête pour récupérer les ressources sans doublons
	query := `
	SELECT DISTINCT ON (r.id) 
		r.id AS resource_id, 
		r.name AS resource_name, 
		r.description AS resource_description,
		r.duration as resource_duration
	FROM user_school_resource usr
	JOIN resource r ON usr.resource_id = r.id
	WHERE usr.school_id = $1`

	rows, err := database.GetConn().Query(query, schoolID)
	if err != nil {
		http.Error(w, "Failed to execute query: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var resources []map[string]interface{}

	for rows.Next() {
		var id, duration int
		var name, description string

		if err := rows.Scan(&id, &name, &description, &duration); err != nil {
			http.Error(w, "Failed to scan row: "+err.Error(), http.StatusInternalServerError)
			return
		}

		resource := map[string]interface{}{
			"resource_id":          id,
			"resource_name":        name,
			"resource_description": description,
			"school_id":            schoolID,
			"school_name":          schoolName,
			"resource_duration":    duration,
		}

		resources = append(resources, resource)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Error iterating rows: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resources)
}
func FetchSchoolFromUSR(w http.ResponseWriter, r *http.Request) {
	s := r.URL.Query().Get("schoolID")

	schoolID, err := strconv.Atoi(s)
	if err != nil {
		http.Error(w, "Invalid schoolID: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Récupérer le nom de l’école
	var schoolName string
	err = database.GetConn().QueryRow(`
		SELECT name FROM school WHERE id = $1`, s).Scan(&schoolName)
	if err != nil {
		http.Error(w, "Failed to get school: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Requête pour récupérer les ressources sans doublons
	query := `
	SELECT DISTINCT ON (r.id) 
		r.id AS resource_id, 
		r.name AS resource_name, 
		r.description AS resource_description
	FROM user_school_resource usr
	JOIN resource r ON usr.resource_id = r.id
	WHERE usr.school_id = $1`

	rows, err := database.GetConn().Query(query, schoolID)
	if err != nil {
		http.Error(w, "Failed to execute query: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var resources []map[string]interface{}

	for rows.Next() {
		var id int
		var name, description string

		if err := rows.Scan(&id, &name, &description); err != nil {
			http.Error(w, "Failed to scan row: "+err.Error(), http.StatusInternalServerError)
			return
		}

		resource := map[string]interface{}{
			"resource_id":          id,
			"resource_name":        name,
			"resource_description": description,
			"school_id":            schoolID,
			"school_name":          schoolName,
		}

		resources = append(resources, resource)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Error iterating rows: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resources)
}
func FetchUserBySchoolID(w http.ResponseWriter, r *http.Request) {
	s := r.URL.Query().Get("schoolID")

	schoolID, err := strconv.Atoi(s)
	if err != nil {
		http.Error(w, "Invalid schoolID: "+err.Error(), http.StatusBadRequest)
		return
	}

	query := `
	SELECT * 
	FROM user_school_resource
	WHERE school_id = $1`

	rows, err := database.GetConn().Query(query, schoolID)
	if err != nil {
		http.Error(w, "Failed to execute query: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		http.Error(w, "Failed to get columns: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var userResources []map[string]interface{}
	userCache := make(map[string]string)
	addedUserIDs := make(map[string]bool)

	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			fmt.Printf("DEBUG: Error scanning row: %v\n", err)
			http.Error(w, "Failed to scan row: "+err.Error(), http.StatusInternalServerError)
			return
		}

		userResource := make(map[string]interface{})
		var userID string

		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				userResource[col] = string(b)
			} else {
				userResource[col] = val
			}
			if col == "user_id" {
				userID = fmt.Sprintf("%v", val)
				if _, found := addedUserIDs[userID]; found {
					continue
				}
				if userName, found := userCache[userID]; found {
					userResource["user_name"] = userName
				} else {
					var given_name, family_name, unique_name string
					err := database.GetConn().QueryRow(`
					SELECT given_name, family_name, unique_name 
					FROM usr 
					WHERE id = $1`, userID).Scan(&given_name, &family_name, &unique_name)
					if err != nil {
						http.Error(w, "Failed to get user: "+err.Error(), http.StatusInternalServerError)
						return
					}
					userName := given_name + " " + family_name
					userCache[userID] = userName
					userResource["user_name"] = userName
					userResource["user_given_name"] = given_name
					userResource["user_family_name"] = family_name
					userResource["user_unique_name"] = unique_name
				}
			}
			if col == "resource_id" {
				var resourceName, resourceDescription string
				err := database.GetConn().QueryRow(`
				SELECT name, description 
				FROM resource 
				WHERE id = $1`, val).Scan(&resourceName, &resourceDescription)
				if err != nil {
					http.Error(w, "Failed to get resource: "+err.Error(), http.StatusInternalServerError)
					return
				}
				userResource["resource_name"] = resourceName
				userResource["resource_description"] = resourceDescription
			}
			if col == "school_id" {
				var schoolName string
				err := database.GetConn().QueryRow(`
				SELECT name 
				FROM school 
				WHERE id = $1`, val).Scan(&schoolName)
				if err != nil {
					http.Error(w, "Failed to get school: "+err.Error(), http.StatusInternalServerError)
					return
				}
				userResource["school_name"] = schoolName
			}
		}
		userResources = append(userResources, userResource)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Error iterating rows: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userResources)

}
func FetchUserResource(w http.ResponseWriter, r *http.Request) {
	schoolID := r.URL.Query().Get("schoolID")
	fmt.Printf("DEBUG: FetchUserResource userResourceID: %s\n", schoolID)

	query := `
	SELECT * 
	FROM user_school_resource
	WHERE school_id = $1`

	rows, err := database.GetConn().Query(query, schoolID)
	if err != nil {
		http.Error(w, "Failed to execute query: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		http.Error(w, "Failed to get columns: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var userResources []map[string]interface{}
	userCache := make(map[string]string)
	addedUserIDs := make(map[string]bool)

	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			http.Error(w, "Failed to scan row: "+err.Error(), http.StatusInternalServerError)
			return
		}

		userResource := make(map[string]interface{})
		var user_id string

		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				userResource[col] = string(b)
			} else {
				userResource[col] = val
			}
			if col == "user_id" {
				user_id = fmt.Sprintf("%v", val)
				if _, found := addedUserIDs[user_id]; found {
					break
				}
				if userName, found := userCache[user_id]; found {
					userResource["user_name"] = userName
				} else {
					var given_name, family_name string
					err := database.GetConn().QueryRow(`
					SELECT given_name, family_name 
					FROM usr 
					WHERE id = $1`, user_id).Scan(&given_name, &family_name)
					if err != nil {
						http.Error(w, "Failed to get user: "+err.Error(), http.StatusInternalServerError)
						return
					}
					userName := given_name + " " + family_name
					userCache[user_id] = userName
					userResource["user_name"] = userName
				}
			}
			if col == "resource_id" {
				var resource_name, resource_description string
				err := database.GetConn().QueryRow(`
				SELECT name, description 
				FROM resource 
				WHERE id = $1`, val).Scan(&resource_name, &resource_description)
				if err != nil {
					http.Error(w, "Failed to get resource: "+err.Error(), http.StatusInternalServerError)
					return
				}
				userResource["resource_name"] = resource_name
				userResource["resource_description"] = resource_description
			}
			if col == "school_id" {
				var school_name string
				err := database.GetConn().QueryRow(`
				SELECT name 
				FROM school 
				WHERE id = $1`, val).Scan(&school_name)
				if err != nil {
					http.Error(w, "Failed to get school: "+err.Error(), http.StatusInternalServerError)
					return
				}
				userResource["school_name"] = school_name
			}
		}
		userResources = append(userResources, userResource)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Error iterating rows: "+err.Error(), http.StatusInternalServerError)
		return
	}

	for _, userResource := range userResources {
		fmt.Printf("DEBUG: %v\n", userResource)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, "Error iterating rows: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userResources)
}
func AddUserSchoolResource(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("DEBUG: UpdateUserSchoolResource\n")
	userID := r.URL.Query().Get("userID")
	schoolID := r.URL.Query().Get("schoolID")
	resourceID := r.URL.Query().Get("resourceID")

	if userID == "" || schoolID == "" || resourceID == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}
	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		http.Error(w, "Invalid userID: "+err.Error(), http.StatusBadRequest)
		return
	}
	schoolIDInt, err := strconv.Atoi(schoolID)
	if err != nil {
		http.Error(w, "Invalid schoolID: "+err.Error(), http.StatusBadRequest)
		return
	}
	resourceIDInt, err := strconv.Atoi(resourceID)
	if err != nil {
		http.Error(w, "Invalid resourceID: "+err.Error(), http.StatusBadRequest)
		return
	}

	query := `
	INSERT INTO 
	user_school_resource (user_id, school_id, resource_id) 
	VALUES ($1, $2, $3)`

	result, err := database.GetConn().Exec(query, userIDInt, schoolIDInt, resourceIDInt)
	if err != nil {
		http.Error(w, "Failed to execute query: "+err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, "Failed to get rows affected: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"rowsAffected": rowsAffected})
}
func UpdateUserSchoolResource(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("DEBUG: UpdateUserSchoolResource\n")
	userID := r.URL.Query().Get("userID")
	schoolID := r.URL.Query().Get("schoolID")
	resourceID := r.URL.Query().Get("resourceID")

	if userID == "" || schoolID == "" || resourceID == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}
	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		http.Error(w, "Invalid userID: "+err.Error(), http.StatusBadRequest)
		return
	}
	schoolIDInt, err := strconv.Atoi(schoolID)
	if err != nil {
		http.Error(w, "Invalid schoolID: "+err.Error(), http.StatusBadRequest)
		return
	}
	resourceIDInt, err := strconv.Atoi(resourceID)
	if err != nil {
		http.Error(w, "Invalid resourceID: "+err.Error(), http.StatusBadRequest)
		return
	}

	query := `
	UPDATE user_school_resource 
	SET resource_id = $1 
	WHERE user_id = $2 
	AND school_id = $3`

	result, err := database.GetConn().Exec(query, resourceIDInt, userIDInt, schoolIDInt)
	if err != nil {
		http.Error(w, "Failed to execute query: "+err.Error(), http.StatusInternalServerError)
		return
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, "Failed to get rows affected: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"rowsAffected": rowsAffected})
}
func DeleteUserSchoolResource(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("DEBUG: DeleteUserSchoolResource\n")
	userID := r.URL.Query().Get("userID")
	schoolID := r.URL.Query().Get("schoolID")
	resourceID := r.URL.Query().Get("resourceID")

	if userID == "" || schoolID == "" || resourceID == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	uid, err := strconv.Atoi(userID)
	if err != nil {
		http.Error(w, "Invalid userID: "+err.Error(), http.StatusBadRequest)
		return
	}
	sid, err := strconv.Atoi(schoolID)
	if err != nil {
		http.Error(w, "Invalid schoolID: "+err.Error(), http.StatusBadRequest)
		return
	}

	rid, err := strconv.Atoi(resourceID)
	if err != nil {
		http.Error(w, "Invalid resourceID: "+err.Error(), http.StatusBadRequest)
		return
	}

	query := `
	DELETE FROM 
	user_school_resource 
	WHERE user_id = $1 
	AND school_id = $2
	AND resource_id = $3`

	result, err := database.GetConn().Exec(query, uid, sid, rid)
	if err != nil {
		http.Error(w, "Failed to execute query: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Vérifie si l'utilisateur a encore des entrées restantes
	checkQuery := `
	SELECT COUNT(*) 
	FROM user_school_resource 
	WHERE user_id = $1 AND school_id = $2`

	var count int
	err = database.GetConn().QueryRow(checkQuery, uid, sid).Scan(&count)
	if err != nil {
		http.Error(w, "Failed to check remaining entries: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Si l'utilisateur n'a plus d'entrées, insérer une entrée par défaut en fonction de l'école
	if count == 0 {
		// Récupérer l'ID de la ressource par défaut
		var resourceID int
		rIDQuery := `
		SELECT id
		FROM resource
		WHERE school = $1
		AND name ILIKE '%guest%'
		or name ILIKE '%invite%'
		`
		err := database.GetConn().QueryRow(rIDQuery, sid).Scan(&resourceID)
		if err != nil {
			http.Error(w, "Failed to get resource ID: "+err.Error(), http.StatusInternalServerError)
			return
		}
		// Insérer l'entrée par défaut
		insertQuery := `
		INSERT INTO user_school_resource (user_id, school_id, resource_id)
		VALUES ($1, $2, $3)`
		_, err = database.GetConn().Exec(insertQuery, uid, sid, resourceID)
		if err != nil {
			http.Error(w, "Failed to insert default entry: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, "Failed to get rows affected: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"rowsAffected": rowsAffected})
}
func AddUserSchoolResourceTrigger(w http.ResponseWriter, r *http.Request, id int) {
	fmt.Printf("DEBUG: AddUserSchoolResourceTrigger\n")
	schoolID := r.URL.Query().Get("schoolID")

	if schoolID == "" || id == 0 {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	var resourceID int

	rIDQuery := `
	SELECT id
	FROM resource
	WHERE school = $1
	AND name ILIKE '%guest%'
	or name ILIKE '%invite%'
	`

	err := database.GetConn().QueryRow(rIDQuery, schoolID).Scan(&resourceID)
	if err != nil {
		http.Error(w, "Failed to get resource ID: "+err.Error(), http.StatusInternalServerError)
		return
	}

	query := `
	INSERT INTO 
	user_school_resource (user_id, school_id, resource_id) 
	VALUES ($1, $2, $3)`

	_, err = database.GetConn().Exec(query, id, schoolID, resourceID)
	if err != nil {
		http.Error(w, "Failed to execute query: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
