package unavailabilities

import (
	"encoding/json"
	"net/http"

	"backend/internal/database"
)

// get unavailabilities
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

func CreateUnavailability(w http.ResponseWriter, r *http.Request) {
	var unavailability map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&unavailability); err != nil {
		http.Error(w, "Failed to decode request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	query := `
	INSERT 
	INTO unavailabilities (user, date, start, end, reason) 
	VALUES ($1, $2, $3, $4, $5)`

	result, err := database.GetConn().Exec(query, unavailability["user"], unavailability["date"], unavailability["start"], unavailability["end"], unavailability["reason"])
	if err != nil {
		http.Error(w, "Failed to execute query: "+err.Error(), http.StatusInternalServerError)
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		http.Error(w, "Failed to get last insert ID: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"id": id})
}

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
