package workschedule

import (
	"backend/internal/database"
	"backend/internal/handlers/user"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type WorkSchedule struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	DayOfWeek string    `json:"day_of_week"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

// Fonction pour créer un planning de travail
func CreateWorkSchedule(w http.ResponseWriter, r *http.Request) {
	fmt.Println("DEBUG: CreateWorkSchedule called")
	var schedule map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&schedule); err != nil {
		http.Error(w, "Failed to decode request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	idQuery := `SELECT COALESCE(MAX(id), 0) FROM work_schedule`
	var lastID int
	err := database.GetConn().QueryRow(idQuery).Scan(&lastID)
	if err != nil {
		http.Error(w, "Échec de la récupération du dernier ID : "+err.Error(), http.StatusInternalServerError)
		return
	}
	newID := lastID + 1

	userID, err := user.GetUseridFromDB(schedule["user_id"].(string))
	if err != nil {
		http.Error(w, "Échec de la récupération de l'ID utilisateur : "+err.Error(), http.StatusInternalServerError)
		return
	}

	query := `
		INSERT INTO work_schedule (id, user_id, day_of_week, start_time, end_time)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err = database.GetConn().Exec(query, newID, userID, schedule["day_of_week"], schedule["start_time"], schedule["end_time"])
	if err != nil {
		http.Error(w, "Échec de l'insertion du planning de travail : "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Planning de travail créé",
		"id":      newID,
	})
}

// Fonction pour récupérer les plannings de travail par utilisateur
func GetWorkSchedulesByUser(w http.ResponseWriter, r *http.Request) {
	fmt.Println("DEBUG: GetWorkSchedulesByUser called")
	userIDString := r.URL.Query().Get("userID")
	if userIDString == "" {
		http.Error(w, "Identifiant utilisateur manquant", http.StatusBadRequest)
		return
	}

	userID, err := user.GetUseridFromDB(userIDString)
	if err != nil {
		http.Error(w, "Échec de la récupération de l'ID utilisateur : "+err.Error(), http.StatusInternalServerError)
		return
	}
	query := `
		SELECT id, user_id, day_of_week, start_time, end_time
		FROM work_schedule WHERE user_id = $1
		ORDER BY day_of_week, start_time
	`
	rows, err := database.GetConn().Query(query, userID)
	if err != nil {
		http.Error(w, "Erreur lors de la requête : "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	schedulesByDay := make(map[string][]map[string]interface{})
	for rows.Next() {
		var id int
		var user_id, day string
		var start, end time.Time
		if err := rows.Scan(&id, &user_id, &day, &start, &end); err != nil {
			http.Error(w, "Erreur lors de la lecture de la ligne : "+err.Error(), http.StatusInternalServerError)
			return
		}
		schedule := map[string]interface{}{
			"id":         id,
			"user_id":    user_id,
			"start_time": start,
			"end_time":   end,
		}
		schedulesByDay[day] = append(schedulesByDay[day], schedule)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(schedulesByDay)
}

// Fonction pour mettre à jour un planning de travail
func UpdateWorkSchedule(w http.ResponseWriter, r *http.Request) {
	fmt.Println("DEBUG: UpdateWorkSchedule called")
	var schedule map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&schedule); err != nil {
		http.Error(w, "Invalid input: "+err.Error(), http.StatusBadRequest)
		return
	}

	userID, err := user.GetUseridFromDB(schedule["user_id"].(string))
	if err != nil {
		http.Error(w, "Failed to fetch user ID: "+err.Error(), http.StatusInternalServerError)
		return
	}

	query := `
		UPDATE work_schedule 
		SET day_of_week = $1, start_time = $2, end_time = $3 
		WHERE id = $4 AND user_id = $5
	`
	result, err := database.GetConn().Exec(query, schedule["day_of_week"], schedule["start_time"], schedule["end_time"], schedule["id"], userID)
	if err != nil {
		http.Error(w, "Update failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	rowsAffected, _ := result.RowsAffected()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"updated": rowsAffected,
	})
}

// Fonction pour supprimer un planning de travail
func DeleteWorkSchedule(w http.ResponseWriter, r *http.Request) {
	fmt.Println("DEBUG: DeleteWorkSchedule called")
	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input: "+err.Error(), http.StatusBadRequest)
		return
	}

	userID, err := user.GetUseridFromDB(req["user_id"].(string))
	if err != nil {
		http.Error(w, "Failed to fetch user ID: "+err.Error(), http.StatusInternalServerError)
		return
	}

	query := `DELETE FROM work_schedule WHERE id = $1 AND user_id = $2`
	result, err := database.GetConn().Exec(query, req["id"], userID)
	if err != nil {
		http.Error(w, "Delete failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	rowsAffected, _ := result.RowsAffected()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"deleted": rowsAffected,
	})
}
