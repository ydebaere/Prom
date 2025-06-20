package user

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"backend/internal/database"
	userSchoolResource "backend/internal/handlers/userSchoolResource"

	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

type Claims struct {
	Unique_name string `json:"unique_name"`
	Firstname   string `json:"firstname"`
	Lastname    string `json:"lastname"`
	Id          int    `json:"id"`
	Role        int    `json:"role"`
	School      int    `json:"school"`
	jwt.StandardClaims
}

func FetchUser(w http.ResponseWriter, r *http.Request) {
	fmt.Println("DEBUGG: FetchUser called")
	userID := r.URL.Query().Get("userID")

	query := `
	SELECT * 
	FROM usr 
	where id = $1`

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

	var users []map[string]interface{}
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

		user := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				user[col] = string(b)
			} else {
				user[col] = val
			}
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

func FetchUsers(w http.ResponseWriter, r *http.Request) {
	fmt.Println("DEBUGG: FetchUsers called")
	query := `
	SELECT * 
	FROM usr`

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

	var users []map[string]interface{}
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

		user := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				user[col] = string(b)
			} else {
				user[col] = val
			}
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

func FetchAgent(w http.ResponseWriter, r *http.Request) {
	fmt.Println("DEBUGG: FetchAgent called")
	schoolID := r.URL.Query().Get("schoolID")
	resourceID := r.URL.Query().Get("roleID")

	var query string

	if resourceID == "2" {
		query = `
			SELECT u.id, u.unique_name, given_name, family_name
			FROM usr u
			JOIN user_school_resource usr 
			ON u.id = usr.user_id
			WHERE (resource_id = 2) AND school_id = $1`
	} else {
		query = `
			SELECT u.id, u.unique_name, given_name, family_name
			FROM usr u
			JOIN user_school_resource usr 
			ON u.id = usr.user_id
			WHERE resource_id = 3 AND school_id = $1`
	}

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

	var users []map[string]interface{}
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

		user := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				user[col] = string(b)
			} else {
				user[col] = val
			}
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

func CreateUser(w http.ResponseWriter, r *http.Request) {
	fmt.Println("DEBUGG: CreateUser called")
	var user map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Failed to decode request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Récupérer le dernier ID utilisateur et l'incrémenter
	idQuery := `
    SELECT COALESCE(MAX(id), 0) 
    FROM usr`
	var lastID int
	err := database.GetConn().QueryRow(idQuery).Scan(&lastID)
	if err != nil {
		http.Error(w, "Failed to fetch last ID: "+err.Error(), http.StatusInternalServerError)
		return
	}
	newID := lastID + 1

	// Insérer le nouvel utilisateur avec le nouvel ID
	query := `
    INSERT INTO 
    usr (id, given_name, family_name, unique_name) 
    VALUES ($1, $2, $3, $4)
	`
	result, err := database.GetConn().Exec(query, newID, user["given_name"], user["family_name"], user["unique_name"])
	if err != nil {
		http.Error(w, "Failed to execute query: "+err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, "Failed to get rows affected: "+err.Error(), http.StatusInternalServerError)
		return
	}

	s := r.URL.Query().Get("schoolID")
	if s != "" {
		// Ajouter la liaison utilisateur-école-ressource
		userSchoolResource.AddUserSchoolResourceTrigger(w, r, newID)
	}

	// Retourner l'ID du nouvel utilisateur
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": rowsAffected,
		"userID":  newID,
	})
}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	fmt.Println("DEBUGG: UpdateUser called")
	var user map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Failed to decode request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Vérifier si le mot de passe est fourni et le hacher
	if password, ok := user["password"].(string); ok && password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Failed to hash password: "+err.Error(), http.StatusInternalServerError)
			return
		}
		user["password"] = string(hashedPassword)
	}

	// Construire la requête de mise à jour
	query := `
	UPDATE 
	usr 
	SET give_name = $1, family_name = $2, unique_name = $3`

	args := []interface{}{user["given_name"], user["family_name"], user["unique_name"]}

	// Ajouter le mot de passe à la requête si fourni
	if _, ok := user["password"].(string); ok {
		query += ", password = ?"
		args = append(args, user["password"])
	}

	query += `
	WHERE id = $1`

	args = append(args, user["id"])

	result, err := database.GetConn().Exec(query, args...)
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

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("userID")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	db := database.GetConn()
	tx, err := db.Begin()
	if err != nil {
		http.Error(w, "Failed to begin transaction: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Remplacer user par ID fantôme dans appointments
	_, err = tx.Exec(`
		UPDATE appointment 
		SET 
			host = CASE WHEN host = $1 THEN 0 ELSE host END,
			guest = CASE WHEN guest = $1 THEN 0 ELSE guest END
		WHERE host = $1 OR guest = $1
	`, userID)
	if err != nil {
		http.Error(w, "Failed to update appointments: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Supprimer le user
	_, err = tx.Exec(`DELETE FROM usr WHERE id = $1`, userID)
	if err != nil {
		http.Error(w, "Failed to delete user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err = tx.Commit(); err != nil {
		http.Error(w, "Transaction commit failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Utilisateur supprimé avec succès"})
}

func GetUseridFromDB(uniqueName string) (int, error) {
	query := `
	SELECT id 
	FROM usr 
	WHERE unique_name = $1`
	var userID int
	err := database.GetConn().QueryRow(query, uniqueName).Scan(&userID)
	if err != nil {
		return 0, err
	}
	return userID, nil
}
