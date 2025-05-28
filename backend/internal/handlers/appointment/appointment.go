package appointment

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"backend/internal/config"
	"backend/internal/database"
	"backend/internal/handlers/service"
)

// Représente un créneau horaire avec heure de début et de fin
type TimeSlot struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// Représente un rendez-vous
type Appointment struct {
	ID        int       `json:"id,omitempty"`
	Host      string    `json:"host,omitempty"`
	Guest     string    `json:"guest,omitempty"`
	School    string    `json:"school,omitempty"`
	Resource  string    `json:"resource,omitempty"`
	StartTime time.Time `json:"start_time,omitempty"`
	EndTime   time.Time `json:"end_time,omitempty"`
	Title     string    `json:"title,omitempty"`
	Status    bool      `json:"status,omitempty"`
	Token     string    `json:"token,omitempty"`
}

// Soustraire des créneaux horaires
func subtractSlots(base []TimeSlot, removes []TimeSlot) []TimeSlot {
	var result []TimeSlot

	for _, slot := range base {
		subtracted := []TimeSlot{slot}

		for _, rem := range removes {
			var temp []TimeSlot
			for _, s := range subtracted {
				temp = append(temp, subtractTwo(s, rem)...)
			}
			subtracted = temp
		}
		result = append(result, subtracted...)
	}

	return result
}

// Soustraire deux créneaux horaires
func subtractTwo(a, b TimeSlot) []TimeSlot {
	if b.End.Before(a.Start) || b.Start.After(a.End) {
		return []TimeSlot{a}
	}

	var slots []TimeSlot
	if b.Start.After(a.Start) {
		slots = append(slots, TimeSlot{Start: a.Start, End: b.Start})
	}
	if b.End.Before(a.End) {
		slots = append(slots, TimeSlot{Start: b.End, End: a.End})
	}
	return slots
}

// Générer des créneaux horaires à partir de plages horaires
func generateSlotsFromRanges(ranges []TimeSlot, duration time.Duration) []TimeSlot {
	var slots []TimeSlot
	for _, r := range ranges {
		start := r.Start
		for start.Add(duration).Before(r.End) || start.Add(duration).Equal(r.End) {
			end := start.Add(duration)
			slots = append(slots, TimeSlot{Start: start, End: end})
			start = end
		}
	}
	return slots
}

// Obtenir les créneaux disponibles
func GetAvailableSlots(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("userID")
	dateStr := r.URL.Query().Get("date")
	durationStr := r.URL.Query().Get("duration")

	duration, err := strconv.Atoi(durationStr)
	if err != nil || duration <= 0 {
		http.Error(w, "Paramètre de durée invalide", http.StatusBadRequest)
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		http.Error(w, "Format de date invalide. Format attendu : YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	dayOfWeekMap := map[string]string{
		"monday":    "lundi",
		"tuesday":   "mardi",
		"wednesday": "mercredi",
		"thursday":  "jeudi",
		"friday":    "vendredi",
		"saturday":  "samedi",
		"sunday":    "dimanche",
	}
	dayOfWeek := strings.ToLower(date.Weekday().String())
	dayOfWeek = dayOfWeekMap[dayOfWeek]

	// 1. Horaires
	queryWork := `SELECT start_time, end_time FROM work_schedule WHERE user_id = $1 AND day_of_week = $2`
	rowsWork, err := database.GetConn().Query(queryWork, userID, dayOfWeek)
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des horaires de travail : "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rowsWork.Close()

	var workSlots []TimeSlot
	for rowsWork.Next() {
		var start, end time.Time
		if err := rowsWork.Scan(&start, &end); err != nil {
			http.Error(w, "Erreur lors de la lecture des horaires de travail : "+err.Error(), http.StatusInternalServerError)
			return
		}
		start = time.Date(date.Year(), date.Month(), date.Day(), start.Hour(), start.Minute(), 0, 0, time.Local)
		end = time.Date(date.Year(), date.Month(), date.Day(), end.Hour(), end.Minute(), 0, 0, time.Local)
		workSlots = append(workSlots, TimeSlot{Start: start, End: end})
	}

	// 2. Indispos
	queryUnavail := `SELECT start_time, end_time FROM unavailabilities WHERE user_id = $1 AND date = $2`
	rowsUnavail, err := database.GetConn().Query(queryUnavail, userID, dateStr)
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des indisponibilités : "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rowsUnavail.Close()

	var unavailSlots []TimeSlot
	for rowsUnavail.Next() {
		var start, end time.Time
		if err := rowsUnavail.Scan(&start, &end); err != nil {
			http.Error(w, "Erreur lors de la lecture des indisponibilités : "+err.Error(), http.StatusInternalServerError)
			return
		}
		start = time.Date(date.Year(), date.Month(), date.Day(), start.Hour(), start.Minute(), 0, 0, time.Local)
		end = time.Date(date.Year(), date.Month(), date.Day(), end.Hour(), end.Minute(), 0, 0, time.Local)
		unavailSlots = append(unavailSlots, TimeSlot{Start: start, End: end})
	}

	// 3. Calcul des créneaux nets
	availableRanges := subtractSlots(workSlots, unavailSlots)

	// 4. Découpage par tranches de durée
	finalSlots := generateSlotsFromRanges(availableRanges, time.Duration(duration)*time.Minute)

	// 5. Format final avec label
	var response []map[string]string
	for _, slot := range finalSlots {
		startStr := slot.Start.Format("15:04")
		endStr := slot.End.Format("15:04")
		response = append(response, map[string]string{
			"start": startStr,
			"end":   endStr,
			"label": fmt.Sprintf("%s - %s", startStr, endStr),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Obtenir les prochains rendez-vous par schoolID
func FetchAppointmentsBySchoolID(w http.ResponseWriter, r *http.Request) {
	schoolID := r.URL.Query().Get("schoolID")
	query := `
    SELECT a.*
    FROM appointment a
    JOIN user u
    ON a.host = u.id
    JOIN school s
    ON a.school = s.id
    WHERE s.id = $1
    AND a.start_time >= NOW()
    ORDER BY a.start_time 
	ASC`

	rows, err := database.GetConn().Query(query, schoolID)
	if err != nil {
		http.Error(w, "Échec de l'exécution de la requête : "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		http.Error(w, "Échec de la récupération des colonnes : "+err.Error(), http.StatusInternalServerError)
		return
	}

	var appointments []map[string]interface{}

	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			http.Error(w, "Échec de la lecture de la ligne : "+err.Error(), http.StatusInternalServerError)
			return
		}

		appointment := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				appointment[col] = string(b)
			} else {
				appointment[col] = val
			}
			// Si la colonne est 'host', trouver le nom de l'hôte dans la table user
			if col == "host" {
				hostID := val
				var firstName string
				var lastName string
				err := database.GetConn().QueryRow(`
				SELECT firstname, lastname 
				FROM user 
				WHERE id = $1`, hostID).Scan(&firstName, &lastName)
				if err != nil {
					http.Error(w, "Impossible de trouver le nom de l'hôte : "+err.Error(), http.StatusInternalServerError)
					return
				}
				appointment["host_name"] = lastName + " " + firstName
			}
			// If the column is 'guest', find the guest name from the user table
			if col == "guest" {
				guestID := val
				var firstName string
				var lastName string
				err := database.GetConn().QueryRow(`
				SELECT firstname, lastname 
				FROM user WHERE id = $1`, guestID).Scan(&firstName, &lastName)
				if err != nil {
					http.Error(w, "Impossible de trouver le nom de l'invité : "+err.Error(), http.StatusInternalServerError)
					return
				}
				appointment["guest_name"] = lastName + " " + firstName
			}
			// If the column is 'school', find the school name from the school table
			if col == "school" {
				schoolID := val
				var name string
				err := database.GetConn().QueryRow(`
				SELECT name 
				FROM school 
				WHERE id = $1`, schoolID).Scan(&name)
				if err != nil {
					http.Error(w, "Impossible de trouver le nom de l'école : "+err.Error(), http.StatusInternalServerError)
					return
				}
				appointment["school_name"] = name
			}

		}

		appointments = append(appointments, appointment)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Erreur lors de l'itération des lignes : "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(appointments); err != nil {
		http.Error(w, "Échec de l'encodage de la réponse : "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// Obtenir les prochains rendez-vous par userID
func FetchAppointmentsByUserID(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("userID")
	query := `
	SELECT a.id, a.host, a.guest, a.school, a.resource, a.start_time, a.end_time, a.title, a.token, a.status
	FROM appointment a
	JOIN usr u1 ON a.host = u1.id
	JOIN usr u2 ON a.guest = u2.id
	JOIN school s ON a.school = s.id
	WHERE (u1.unique_name = $1 OR u2.unique_name = $1)
	AND a.start_time >= now()
	ORDER BY a.start_time ASC`

	rows, err := database.GetConn().Query(query, userID)
	if err != nil {
		http.Error(w, "Échec de l'exécution de la requête : "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var appointments []Appointment

	for rows.Next() {
		var appt Appointment
		err := rows.Scan(
			&appt.ID,
			&appt.Host,
			&appt.Guest,
			&appt.School,
			&appt.Resource,
			&appt.StartTime,
			&appt.EndTime,
			&appt.Title,
			&appt.Token,
			&appt.Status,
		)
		if err != nil {
			http.Error(w, "Échec de la lecture de la ligne : "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Récupérer les noms associés
		var hostFirstName, hostLastName string
		err = database.GetConn().QueryRow(`SELECT given_name, family_name FROM usr WHERE id = $1`, appt.Host).Scan(&hostFirstName, &hostLastName)
		if err == nil {
			appt.Host = hostFirstName + " " + hostLastName
		}

		var guestFirstName, guestLastName string
		err = database.GetConn().QueryRow(`SELECT given_name, family_name FROM usr WHERE id = $1`, appt.Guest).Scan(&guestFirstName, &guestLastName)
		if err == nil {
			appt.Guest = guestFirstName + " " + guestLastName
		}

		var schoolName string
		err = database.GetConn().QueryRow(`SELECT name FROM school WHERE id = $1`, appt.School).Scan(&schoolName)
		if err == nil {
			appt.School = schoolName
		}

		appointments = append(appointments, appt)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Erreur lors de l'itération des lignes : "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(appointments); err != nil {
		http.Error(w, "Échec de l'encodage de la réponse : "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// Récupérer les rendez-vous
func FetchAppointments(w http.ResponseWriter, r *http.Request) {
	query := `
	SELECT id, host, guest, school, resource, start_time, end_time, title, token, status
	FROM appointment`

	rows, err := database.GetConn().Query(query)
	if err != nil {
		http.Error(w, "Échec de l'exécution de la requête : "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var appointments []Appointment

	for rows.Next() {
		var appt Appointment
		err := rows.Scan(
			&appt.ID,
			&appt.Host,
			&appt.Guest,
			&appt.School,
			&appt.Resource,
			&appt.StartTime,
			&appt.EndTime,
			&appt.Title,
			&appt.Token,
			&appt.Status,
		)
		if err != nil {
			http.Error(w, "Échec de la lecture de la ligne : "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Récupérer les noms associés
		var hostFirstName, hostLastName string
		err = database.GetConn().QueryRow(`SELECT given_name, family_name FROM usr WHERE id = $1`, appt.Host).Scan(&hostFirstName, &hostLastName)
		if err == nil {
			appt.Host = hostFirstName + " " + hostLastName
		}

		var guestFirstName, guestLastName string
		err = database.GetConn().QueryRow(`SELECT given_name, family_name FROM usr WHERE id = $1`, appt.Guest).Scan(&guestFirstName, &guestLastName)
		if err == nil {
			appt.Guest = guestFirstName + " " + guestLastName
		}

		var schoolName string
		err = database.GetConn().QueryRow(`SELECT name FROM school WHERE id = $1`, appt.School).Scan(&schoolName)
		if err == nil {
			appt.School = schoolName
		}

		appointments = append(appointments, appt)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Erreur lors de l'itération des lignes : "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(appointments)
}

// Récupérer un rendez-vous par ID
func FetchAppointment(w http.ResponseWriter, r *http.Request) {
	appointmentID := r.URL.Query().Get("appointmentID")
	query := `
	SELECT id, host, guest, school, resource, start_time, end_time, title, token, status
	FROM appointment
	WHERE id = $1`

	var appt Appointment
	err := database.GetConn().QueryRow(query, appointmentID).Scan(
		&appt.ID,
		&appt.Host,
		&appt.Guest,
		&appt.School,
		&appt.Resource,
		&appt.StartTime,
		&appt.EndTime,
		&appt.Title,
		&appt.Token,
		&appt.Status,
	)
	if err != nil {
		http.Error(w, "Échec de la lecture de la ligne : "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Récupérer les noms associés
	var hostFirstName, hostLastName string
	err = database.GetConn().QueryRow(`SELECT given_name, family_name FROM usr WHERE id = $1`, appt.Host).Scan(&hostFirstName, &hostLastName)
	if err == nil {
		appt.Host = hostFirstName + " " + hostLastName
	}

	var guestFirstName, guestLastName string
	err = database.GetConn().QueryRow(`SELECT given_name, family_name FROM usr WHERE id = $1`, appt.Guest).Scan(&guestFirstName, &guestLastName)
	if err == nil {
		appt.Guest = guestFirstName + " " + guestLastName
	}

	var schoolName string
	err = database.GetConn().QueryRow(`SELECT name FROM school WHERE id = $1`, appt.School).Scan(&schoolName)
	if err == nil {
		appt.School = schoolName
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(appt)
}

// Créer un rendez-vous
func CreateAppointment(w http.ResponseWriter, r *http.Request) {
	type AppointmentPayload struct {
		School    int    `json:"school"`
		Host      int    `json:"host"`
		Guest     string `json:"guest"`
		Resource  int    `json:"resource"`
		Title     string `json:"title"`
		Date      string `json:"date"`
		StartTime int64  `json:"start_time"`
		EndTime   int64  `json:"end_time"`
	}

	var payload AppointmentPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Échec du décodage du corps de la requête : "+err.Error(), http.StatusBadRequest)
		return
	}

	startTime := time.Unix(payload.StartTime, 0).Format("2006-01-02 15:04:05")
	endTime := time.Unix(payload.EndTime, 0).Format("2006-01-02 15:04:05")
	guestEmail := payload.Guest
	schoolID := payload.School

	var emailUrl string
	if strings.Contains(r.Host, "localhost") {
		emailUrl = "http://localhost:9000/validate-appointment?"
	} else {
		emailUrl = "https://prometheus.hainaut-promsoc.be/validate-appointment?"
	}

	var userExists bool
	var userID int
	userExistsQuery := `
	SELECT EXISTS (
		SELECT 1
		FROM usr
		WHERE unique_name = $1
	)`
	err := database.GetConn().QueryRow(userExistsQuery, guestEmail).Scan(&userExists)
	if err != nil {
		http.Error(w, "Échec de la requête à la base de données : "+err.Error(), http.StatusInternalServerError)
		return
	}
	if userExists {
		userIDQuery := `
		SELECT id 
		FROM usr 
		WHERE unique_name = $1 
		ORDER BY id 
		DESC LIMIT 1`
		err = database.GetConn().QueryRow(userIDQuery, guestEmail).Scan(&userID)
		if err != nil {
			http.Error(w, "Échec de la requête à la base de données : "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		guestGivenName := ""
		guestFamilyName := ""
		var parts []string
		if strings.Contains(guestEmail, ".") {
			parts = strings.SplitN(guestEmail, ".", 2)
		} else {
			parts = []string{"", ""}
		}
		if len(parts) == 2 {
			guestGivenName = parts[0]
			guestFamilyName = parts[1]
		} else {
			guestGivenName = guestEmail
			guestFamilyName = ""
		}
		userQuery := `
		INSERT INTO usr 
		(unique_name, given_name, family_name) 
		VALUES ($1, $2, $3)`
		_, err = database.GetConn().Exec(userQuery, guestEmail, guestGivenName, guestFamilyName)
		if err != nil {
			http.Error(w, "Échec de l'insertion des données dans la base de données : "+err.Error(), http.StatusInternalServerError)
			return
		}
		userIDQuery := `
		SELECT id 
		FROM usr 
		WHERE unique_name = $1 
		ORDER BY id 
		DESC LIMIT 1`
		err = database.GetConn().QueryRow(userIDQuery, guestEmail).Scan(&userID)
		if err != nil {
			http.Error(w, "Échec de la requête à la base de données : "+err.Error(), http.StatusInternalServerError)
			return
		}
		userRoleQuery := `
		INSERT INTO 
		user_school_resource (user_id, school_id, resource_id) 
		VALUES ($1, $2, $3)`
		_, err = database.GetConn().Exec(userRoleQuery, userID, schoolID, 4)
		if err != nil {
			http.Error(w, "Échec de l'insertion des données dans la base de données : "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	var guestID int
	query := `SELECT id FROM usr WHERE unique_name = $1`
	err = database.GetConn().QueryRow(query, guestEmail).Scan(&guestID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Invité non trouvé : "+err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "Échec de l'exécution de la requête : "+err.Error(), http.StatusInternalServerError)
		return
	}

	if CheckAppointmentsStatus(guestID) {
		tx, err := database.GetConn().Begin()
		if err != nil {
			http.Error(w, "Échec du démarrage de la transaction : "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer func() {
			if p := recover(); p != nil {
				tx.Rollback()
				http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
				return
			} else if err != nil {
				tx.Rollback()
			} else {
				err = tx.Commit()
			}
		}()

		var userID int
		userQuery := `SELECT id FROM usr WHERE unique_name = $1`
		err = tx.QueryRow(userQuery, guestEmail).Scan(&userID)
		if err != nil {
			userInsert := `INSERT INTO usr (unique_name) VALUES ($1) RETURNING id`
			err = tx.QueryRow(userInsert, guestEmail).Scan(&userID)
			if err != nil {
				http.Error(w, "Échec de l'insertion du nouvel utilisateur : "+err.Error(), http.StatusInternalServerError)
				return
			}
		}

		token, err := service.GenerateUniqueToken()
		if err != nil {
			http.Error(w, "Échec de la génération du token unique : "+err.Error(), http.StatusInternalServerError)
			return
		}

		insertQuery := `
		INSERT INTO appointment (school, host, guest, start_time, end_time, resource, title, token) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, 
		(SELECT unique_name FROM usr WHERE id = $2), 
		(SELECT given_name FROM usr WHERE id = $2), 
		(SELECT family_name FROM usr WHERE id = $2), 
		(SELECT name FROM school WHERE id = $1)`

		var appointmentID int
		var hostUniqueName, given_name, family_name, school_name string
		err = tx.QueryRow(
			insertQuery,
			payload.School,
			payload.Host,
			userID,
			startTime,
			endTime,
			payload.Resource,
			payload.Title,
			token,
		).Scan(&appointmentID, &hostUniqueName, &given_name, &family_name, &school_name)

		if err != nil {
			http.Error(w, "Échec de l'insertion du rendez-vous : "+err.Error(), http.StatusInternalServerError)
			return
		}

		resp := map[string]interface{}{
			"id":          appointmentID,
			"host":        hostUniqueName,
			"guest":       guestEmail,
			"school":      payload.School,
			"school_name": school_name,
			"resource":    payload.Resource,
			"title":       payload.Title,
			"start_time":  payload.StartTime,
			"end_time":    payload.EndTime,
			"token":       token,
			"date":        payload.Date,
		}
		escapedToken := url.QueryEscape(fmt.Sprintf("%v", token))
		fullURL := fmt.Sprintf("%stoken=%s", emailUrl, escapedToken)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)

		body := fmt.Sprintf(`
		<!DOCTYPE html>
		<html lang="fr">
		<head>
		<meta charset="UTF-8">
		<title>Confirmation de rendez-vous</title>
		</head>
		<body style="font-family: Arial, sans-serif; color: #333; padding: 20px;">
		<div style="max-width: 600px; margin: auto; border: 1px solid #eee; border-radius: 8px; overflow: hidden; background-color: #fafafa;">
			<div style="background-color: #26a69a; padding: 10px;">
			<img src="https://upload.wikimedia.org/wikipedia/commons/2/29/Logo_Province_de_Hainaut.png" alt="Logo Prometheus" style="height: 40px;">
			</div>
			<div style="padding: 20px;">
			<h2 style="color: #26a69a;">Demande de confirmation de rendez-vous</h2>
			<p><strong>Date :</strong> %s</p>
			<p><strong>Début :</strong> %s</p>
			<p><strong>Fin :</strong> %s</p>
			<p><strong>Lieu :</strong> %s</p>
			<p><strong>Votre contact :</strong> %s %s</p>
			<p><strong>Adresse e-mail :</strong> %s</p>
			<p style="text-align: center;">
				<a href="%s" 
				style="background-color: #26a69a; color: white; padding: 12px 20px; text-decoration: none; border-radius: 4px; display: inline-block;">
				Confirmer mon rendez-vous
				</a>
			</p>
			<p style="text-align: center; font-size: 0.9em; color: #666;">
				Si vous ne confirmez pas votre rendez-vous dans les 24 heures, il sera annulé automatiquement.
			</p>
			<p style="margin-top: 30px;">Merci de votre confiance !<br>L'équipe de <strong>Prometheus</strong></p>
			</div>
			<div style="background-color: #eee; padding: 10px; font-size: 0.9em; text-align: center;">
			Vous n'avez pas demandé ce rendez-vous ? 
			<a href="mailto:ydebare@gmail.com" style="color: #d93025;">Cliquez ici</a>.
			</div>
		</div>
		</body>
		</html>
		`,
			payload.Date,
			time.Unix(payload.StartTime, 0).Format("15:04"),
			time.Unix(payload.EndTime, 0).Format("15:04"),
			school_name,
			given_name,
			family_name,
			hostUniqueName,
			fullURL,
		)
		mailTitle := "Demande de confirmation de rendez-vous"

		service.SendEmail(config.LoadConfig(), guestEmail, mailTitle, body)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]interface{}{"message": "Des rendez-vous non confirmés existent pour cet utilisateur"})
		return
	}
}

// Mettre à jour un rendez-vous
func UpdateAppointment(w http.ResponseWriter, r *http.Request) {
	type AppointmentPayload struct {
		Appointment struct {
			ID    int    `json:"id"`
			Date  string `json:"date"`
			Start string `json:"start"`
			End   string `json:"end"`
		} `json:"appointment"`
	}

	var payload AppointmentPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "JSON invalide : "+err.Error(), http.StatusBadRequest)
		return
	}

	appt := payload.Appointment

	startTimeStr := fmt.Sprintf("%sT%s:00", appt.Date, appt.Start)
	endTimeStr := fmt.Sprintf("%sT%s:00", appt.Date, appt.End)

	startTime, err := time.Parse("2006-01-02T15:04:05", startTimeStr)
	endTime, err2 := time.Parse("2006-01-02T15:04:05", endTimeStr)
	if err != nil || err2 != nil {
		http.Error(w, "Format de date/heure invalide", http.StatusBadRequest)
		return
	}

	query := `
	UPDATE appointment 
	SET start_time = $1, end_time = $2
	WHERE id = $3`

	result, err := database.GetConn().Exec(query, startTime, endTime, appt.ID)
	if err != nil {
		http.Error(w, "Échec de l'exécution de la requête : "+err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, "Échec de la récupération du nombre de lignes affectées : "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"lignesAffectees": rowsAffected})
}

// Supprimer un rendez-vous
func DeleteAppointment(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("appointmentID")

	query := `
	DELETE FROM appointment 
	WHERE id = $1`

	result, err := database.GetConn().Exec(query, userID)
	if err != nil {
		http.Error(w, "Échec de l'exécution de la requête : "+err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, "Échec de la récupération du nombre de lignes affectées : "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"lignesAffectees": rowsAffected,
		"statut":          http.StatusOK,
	})
}

// Vérifier si l'utilisateur a des rendez-vous non confirmés
func CheckAppointmentsStatus(userID int) bool {
	query := `
	SELECT 1
	FROM appointment
	WHERE status = false AND guest = $1
	LIMIT 1
	`
	row := database.GetConn().QueryRow(query, userID)

	var dummy int
	err := row.Scan(&dummy)
	if err == sql.ErrNoRows {
		return true
	} else if err != nil {
		return false
	}

	return false
}

// Fonction pour confirmer un rendez-vous
func ConfirmAppointment(w http.ResponseWriter, r *http.Request) {
	// Récupérer le token de la requête
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Le token est manquant", http.StatusBadRequest)
		return
	}
	var emailUrl string
	if strings.Contains(r.Host, "localhost") {
		emailUrl = "http://localhost:9000/validate-appointment?"
	} else {
		emailUrl = "https://prometheus.hainaut-promsoc.be/validate-appointment?"
	}
	// Rechercher le rendez-vous correspondant au token
	var appointmentID int
	query := `
	SELECT id 
	FROM appointment 
	WHERE token = $1`
	err := database.GetConn().QueryRow(query, token).Scan(&appointmentID)
	if err == sql.ErrNoRows {
		http.Error(w, "Token invalide ou rendez-vous introuvable", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Échec de la requête à la base de données : "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Mettre à jour le statut du rendez-vous
	updateQuery := `
	UPDATE appointment 
	SET status = true 
	WHERE id = $1
	RETURNING id, 
		(SELECT unique_name FROM usr WHERE id = appointment.guest), 
		(SELECT unique_name FROM usr WHERE id = appointment.host),
		(SELECT given_name FROM usr WHERE id = appointment.guest), 
		(SELECT family_name FROM usr WHERE id = appointment.guest), 
		(SELECT name FROM school WHERE id = appointment.school),
		(SELECT start_time from appointment WHERE id = $1),
		(SELECT end_time from appointment WHERE id = $1)`

	var confirmedID int
	var guestUniqueName, hostUniqueName, given_name, family_name, school_name, hostTitle string
	var date, start_time, end_time time.Time
	err = database.GetConn().QueryRow(updateQuery, appointmentID).Scan(
		&confirmedID,
		&guestUniqueName,
		&hostUniqueName,
		&given_name,
		&family_name,
		&school_name,
		&start_time,
		&end_time,
	)
	if err != nil {
		http.Error(w, "Échec de la mise à jour du statut du rendez-vous : "+err.Error(), http.StatusInternalServerError)
		return
	}

	hostTitle = "Vous avez un nouveau rendez-vous!"
	guestTitle := "Votre rendez-vous est confirmé"
	// extraire la date des champs de date et d'heure
	date = start_time.Truncate(24 * time.Hour)

	hostBody := fmt.Sprintf(`
		<!DOCTYPE html>
		<html lang="fr">
		<head>
		<meta charset="UTF-8">
		<title>Nouveau rendez-vous</title>
		</head>
		<body style="font-family: Arial, sans-serif; color: #333; padding: 20px;">
		<div style="max-width: 600px; margin: auto; border: 1px solid #eee; border-radius: 8px; overflow: hidden; background-color: #fafafa;">
			<div style="background-color: #26a69a; padding: 10px;">
			<img src="https://upload.wikimedia.org/wikipedia/commons/2/29/Logo_Province_de_Hainaut.png" alt="Logo Prometheus" style="height: 40px;">
			</div>
			<div style="padding: 20px;">
				<h2 style="color: #26a69a;">Vous avez un nouveau rendez-vous</h2>
				<p><strong>Date :</strong> %s</p>
				<p><strong>Début :</strong> %s</p>
				<p><strong>Fin :</strong> %s</p>
				<p><strong>Lieu :</strong> %s</p>
				<p><strong>Votre contact :</strong> %s %s</p>
				<p><strong>Adresse e-mail :</strong> %s</p>
				<p style="text-align: center;">
				</p>
				<p style="margin-top: 30px;">Merci de votre confiance !<br>L'équipe de <strong>Prometheus</strong></p>
			</div>
		</div>
		</body>
		</html>
		`,
		date.Format("2006-01-02"),
		start_time.Format("15:04"),
		end_time.Format("15:04"),
		school_name,
		given_name,
		family_name,
		guestUniqueName,
	)

	escapedToken := url.QueryEscape(token)
	fullURL := fmt.Sprintf("%stoken=%s", emailUrl, escapedToken)

	guestBody := fmt.Sprintf(`
		<!DOCTYPE html>
		<html lang="fr">
		<head>
		<meta charset="UTF-8">
		<title>Nouveau rendez-vous</title>
		</head>
		<body style="font-family: Arial, sans-serif; color: #333; padding: 20px;">
		<div style="max-width: 600px; margin: auto; border: 1px solid #eee; border-radius: 8px; overflow: hidden; background-color: #fafafa;">
			<div style="background-color: #26a69a; padding: 10px;">
			<img src="https://upload.wikimedia.org/wikipedia/commons/2/29/Logo_Province_de_Hainaut.png" alt="Logo Prometheus" style="height: 40px;">
			</div>
			<div style="padding: 20px;">
				<h2 style="color: #26a69a;">Votre rendez-vous est confirmé</h2>
				<p><strong>Date :</strong> %s</p>
				<p><strong>Début :</strong> %s</p>
				<p><strong>Fin :</strong> %s</p>
				<p><strong>Lieu :</strong> %s</p>
				<p><strong>Votre contact :</strong> %s %s</p>
				<p><strong>Adresse e-mail :</strong> %s</p>
				<p style="text-align: center;">
				</p>
				<p style="text-align: center;">
				<a href="%s" 
				style="background-color: #26a69a; color: white; padding: 12px 20px; text-decoration: none; border-radius: 4px; display: inline-block;">
				Détails du rendez-vous
				</a>
			</p>
				<p style="margin-top: 30px;">Merci de votre confiance !<br>L'équipe de <strong>Prometheus</strong></p>
			</div>
		</div>
		</body>
		</html>
		`,
		date.Format("2006-01-02"),
		start_time.Format("15:04"),
		end_time.Format("15:04"),
		school_name,
		given_name,
		family_name,
		guestUniqueName,
		fullURL,
	)

	service.SendEmail(config.LoadConfig(), hostUniqueName, hostTitle, hostBody)
	service.SendEmail(config.LoadConfig(), guestUniqueName, guestTitle, guestBody)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Rendez-vous confirmé avec succès",
		"status":  http.StatusOK,
	})
}
