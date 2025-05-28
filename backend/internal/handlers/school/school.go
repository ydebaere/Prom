package school

import (
	"encoding/json"
	"fmt"
	"net/http"

	"backend/internal/database"
)

func FetchSchoolByDirector(w http.ResponseWriter, r *http.Request) {
	directorID := r.URL.Query().Get("dirID")
	query := `
	SELECT * 
	FROM school
	WHERE director = $1`

	rows, err := database.GetConn().Query(query, directorID)
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

	var schools []map[string]interface{}
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

		school := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				school[col] = string(b)
			} else {
				school[col] = val
			}
			if col == "director" {
				directorID := val
				var unique_name, given_name, family_name string
				var id int
				err := database.GetConn().QueryRow(`
				SELECT id, unique_name, given_name, family_name
				FROM usr
				WHERE unique_name = $1`, directorID).Scan(&id, &unique_name, &given_name, &family_name)
				if err != nil {
					http.Error(w, "Failed to find guest name: "+err.Error(), http.StatusInternalServerError)
					return
				}
				school["director_name"] = given_name + " " + family_name
			}
		}
		schools = append(schools, school)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Error iterating rows: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(schools)
}

func FetchSchools(w http.ResponseWriter, r *http.Request) {
	query := `
	SELECT * 
	FROM school`

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

	var schools []map[string]interface{}
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

		school := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				school[col] = string(b)
			} else {
				school[col] = val
			}
			if col == "director" {
				directorID := val
				var unique_name, given_name, family_name string
				var id int
				err := database.GetConn().QueryRow(`
				SELECT id, unique_name, given_name, family_name
				FROM usr
				WHERE unique_name = $1`, directorID).Scan(&id, &unique_name, &given_name, &family_name)
				if err != nil {
					http.Error(w, "Failed to find guest name: "+err.Error(), http.StatusInternalServerError)
					return
				}
				school["director_name"] = given_name + " " + family_name
			}
		}
		schools = append(schools, school)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Error iterating rows: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(schools)
}

func FetchSchool(w http.ResponseWriter, r *http.Request) {
	schoolID := r.URL.Query().Get("schoolID")
	if schoolID == "" {
		http.Error(w, "Missing schoolID", http.StatusBadRequest)
		return
	}

	query := `
	SELECT * 
	FROM school 
	WHERE id = $1`

	row := database.GetConn().QueryRow(query, schoolID)

	var school map[string]interface{}
	err := row.Scan()
	if err != nil {
		http.Error(w, "Failed to scan row: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(school)
}

func CreateSchool(w http.ResponseWriter, r *http.Request) {
	fmt.Println("CreateSchool called")
	var school map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&school); err != nil {
		http.Error(w, "Failed to decode request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Récupérer le dernier ID school et l'incrémenter
	idQuery := `
    SELECT COALESCE(MAX(id), 0) 
    FROM school`
	var lastID int
	err := database.GetConn().QueryRow(idQuery).Scan(&lastID)
	if err != nil {
		http.Error(w, "Failed to fetch last ID: "+err.Error(), http.StatusInternalServerError)
		return
	}
	newID := lastID + 1

	// Insérer la nouvelle école avec le nouvel ID
	query := `
	INSERT INTO 
	school (id, name, director) 
	VALUES ($1, $2, $3)`
	result, err := database.GetConn().Exec(query, newID, school["name"], school["director"])
	if err != nil {
		http.Error(w, "Failed to execute query: "+err.Error(), http.StatusInternalServerError)
		return
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, "Failed to get rows affected: "+err.Error(), http.StatusInternalServerError)
		return
	}

	afterInsertSchoolTrigger(newID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"rowsAffected": rowsAffected})
}

func UpdateSchool(w http.ResponseWriter, r *http.Request) {
	var school map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&school); err != nil {
		http.Error(w, "Failed to decode request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	schoolID := r.URL.Query().Get("schoolID")
	if schoolID == "" {
		http.Error(w, "Missing schoolID", http.StatusBadRequest)
		return
	}

	query := `
	UPDATE 
	school 
	SET WHERE id = $1`

	result, err := database.GetConn().Exec(query, schoolID)
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

func DeleteSchool(w http.ResponseWriter, r *http.Request) {
	schoolID := r.URL.Query().Get("schoolID")
	if schoolID == "" {
		http.Error(w, "Missing schoolID", http.StatusBadRequest)
		return
	}

	query := `
	DELETE 
	FROM 
	school 
	WHERE id = $1`

	result, err := database.GetConn().Exec(query, schoolID)
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

func afterInsertSchoolTrigger(id int) {
	fmt.Println("afterInsertSchoolTrigger called for school ID:", id)
	schoolNameQuery := `
	SELECT name
	FROM school
	WHERE id = $1`
	var schoolName string
	err := database.GetConn().QueryRow(schoolNameQuery, id).Scan(&schoolName)
	if err != nil {
		fmt.Println("Error fetching school name for new school:", err)
		return
	}

	resourceIDQuery := `
	SELECT COALESCE(MAX(id), 0) + 1
	FROM resource`
	var resourceID1 int
	err = database.GetConn().QueryRow(resourceIDQuery).Scan(&resourceID1)
	if err != nil {
		fmt.Println("Error fetching next resource ID for new school:", err)
		return
	}
	resourceID2 := resourceID1 + 1

	var name1, name2, description1, description2 string
	var duration int
	var visible bool
	// Define default values for the resource
	name1 = "guest " + schoolName
	description1 = "Invité " + schoolName
	name2 = "user " + schoolName
	description2 = "Utilisateur " + schoolName

	query := `
	INSERT INTO resource (id, name, description, school, duration, visible)
	VALUES ($1, $2, $3, $4, $5, $6)`
	_, err = database.GetConn().Exec(query, resourceID1, name1, description1, id, duration, visible)
	if err != nil {
		fmt.Println("Error inserting default resource for new school:", err)
		return
	}
	_, err = database.GetConn().Exec(query, resourceID2, name2, description2, id, duration, visible)
	if err != nil {
		fmt.Println("Error inserting user resource for new school:", err)
		return
	}
	fmt.Println("Default resource inserted for new school with ID:", id)
}
