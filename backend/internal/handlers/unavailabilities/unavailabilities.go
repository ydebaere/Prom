package unavailabilities

import (
	"encoding/json"
	"net/http"

	"backend/internal/database"
)

// Fonction pour récupérer les indisponibilités
func GetUnavailabilities(w http.ResponseWriter, r *http.Request) {
	query := `
	SELECT *
	FROM 
	unavailabilities`

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

	var unavailabilities []map[string]interface{}
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

		unavailability := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				unavailability[col] = string(b)
			} else {
				unavailability[col] = val
			}
			if col == "user" {
				guestID := val
				var firstName string
				var lastName string
				err := database.GetConn().QueryRow(`
				SELECT firstname, lastname 
				FROM user WHERE id = $1`, guestID).Scan(&firstName, &lastName)
				if err != nil {
					http.Error(w, "Failed to find guest name: "+err.Error(), http.StatusInternalServerError)
					return
				}
				unavailability["user_name"] = lastName + " " + firstName
			}
		}
		unavailabilities = append(unavailabilities, unavailability)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Error iterating rows: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(unavailabilities)
}

// Fonction pour récupérer les indisponibilités d'un utilisateur
func GetUnavailabilitiesByUser(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("userID")

	userIDQuery := `
	SELECT id
	FROM usr
	WHERE unique_name = $1`
	var userID int
	err := database.GetConn().QueryRow(userIDQuery, userIDStr).Scan(&userID)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to execute query: "+err.Error(), http.StatusInternalServerError)
		return
	}

	query := `
	SELECT *
	FROM unavailabilities
	where user_id = $1
	`
	rows, err := database.GetConn().Query(query, userID)
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

	var unavailabilities []map[string]interface{}
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

		unavailability := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				unavailability[col] = string(b)
			} else {
				unavailability[col] = val
			}
			if col == "user" {
				guestID := val
				var firstName string
				var lastName string
				err := database.GetConn().QueryRow(`
				SELECT firstname, lastname 
				FROM user WHERE id = $1`, guestID).Scan(&firstName, &lastName)
				if err != nil {
					http.Error(w, "Failed to find guest name: "+err.Error(), http.StatusInternalServerError)
					return
				}
				unavailability["user_name"] = lastName + " " + firstName
			}
		}
		unavailabilities = append(unavailabilities, unavailability)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Error iterating rows: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(unavailabilities)
}

// Fonction pour créer une indisponibilité
func CreateUnavailability(w http.ResponseWriter, r *http.Request) {
	var unavailability map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&unavailability); err != nil {
		http.Error(w, "Failed to decode request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	userIDQuery := `SELECT id FROM usr WHERE unique_name = $1`
	var userID int
	err := database.GetConn().QueryRow(userIDQuery, unavailability["user_id"]).Scan(&userID)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to execute query: "+err.Error(), http.StatusInternalServerError)
		return
	}
	query := `
	INSERT 
	INTO unavailabilities (user_id, start_time, end_time, date, reason) 
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id`

	var id int
	err = database.GetConn().QueryRow(query, userID, unavailability["start_time"], unavailability["end_time"], unavailability["day"], unavailability["reason"]).Scan(&id)
	if err != nil {
		http.Error(w, "Failed to execute query: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"id": id})
}

// Fonction pour supprimer une indisponibilité
func DeleteUnavailability(w http.ResponseWriter, r *http.Request) {
	unavailabilityID := r.URL.Query().Get("unavailabilityID")

	query := `
	DELETE 
	FROM unavailabilities 
	WHERE id = $1`

	result, err := database.GetConn().Exec(query, unavailabilityID)
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

// Fonction pour mettre à jour une indisponibilité
func UpdateUnavailability(w http.ResponseWriter, r *http.Request) {
	var unavailability map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&unavailability); err != nil {
		http.Error(w, "Failed to decode request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	query := `
	UPDATE unavailabilities 
	SET date = $1, start = $2, end = $3, reason = $4 
	WHERE id = $5`

	result, err := database.GetConn().Exec(query, unavailability["date"], unavailability["start"], unavailability["end"], unavailability["reason"], unavailability["id"])
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
