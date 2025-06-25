package service

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/smtp"
	"net/url"
	"strings"
	"time"

	"backend/internal/config"
	"backend/internal/database"

	ical "github.com/arran4/golang-ical"

	"strconv"
)

type TimeSlot struct {
	Start string `json:"start"`
	End   string `json:"end"`
	Label string `json:"label"`
}

func GetAnonymousAvailableSlots(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("DEBUGG : GetAnonymousAvailableSlots function called")
	schoolID := r.URL.Query().Get("schoolID")
	resourceID := r.URL.Query().Get("resourceID")
	dateStr := r.URL.Query().Get("date")

	if schoolID == "" || resourceID == "" || dateStr == "" {
		http.Error(w, "Paramètres manquants : schoolID, resourceID, date requis", http.StatusBadRequest)
		return
	}

	schoolIDInt, err := strconv.Atoi(schoolID)
	if err != nil {
		http.Error(w, "Format de schoolID invalide", http.StatusBadRequest)
		return
	}

	resourceIDInt, err := strconv.Atoi(resourceID)
	if err != nil {
		resourceIDInt = 12 // Secretariat par défaut
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		http.Error(w, "Format de date invalide. Format attendu : YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	dayOfWeek := strings.ToLower(date.Weekday().String())
	dayMap := map[string]string{
		"monday":    "lundi",
		"tuesday":   "mardi",
		"wednesday": "mercredi",
		"thursday":  "jeudi",
		"friday":    "vendredi",
		"saturday":  "samedi",
		"sunday":    "dimanche",
	}
	day := dayMap[dayOfWeek]

	db := database.GetConn()

	// Étape 1 : récupérer la durée du slot et les agents liés
	queryAgents := `
		SELECT DISTINCT usr.id, r.duration
		FROM user_school_resource usr_res
		JOIN resource r ON usr_res.resource_id = r.id
		JOIN usr ON usr.id = usr_res.user_id
		WHERE r.school = $1 AND r.id = $2
	`
	rows, err := db.Query(queryAgents, schoolIDInt, resourceIDInt)
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des agents : "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type Agent struct {
		ID       int
		Duration int
	}
	var agents []Agent
	for rows.Next() {
		var a Agent
		if err := rows.Scan(&a.ID, &a.Duration); err != nil {
			http.Error(w, "Erreur lors du scan des agents : "+err.Error(), http.StatusInternalServerError)
			return
		}
		agents = append(agents, a)
	}

	var slots []TimeSlot

	for _, agent := range agents {
		// Étape 2 : récupérer les horaires de travail
		queryWork := `
			SELECT start_time, end_time
			FROM work_schedule
			WHERE user_id = $1 AND day_of_week = $2
		`
		workRows, err := db.Query(queryWork, agent.ID, day)
		if err != nil {
			continue
		}
		defer workRows.Close()

		for workRows.Next() {
			var start, end time.Time
			if err := workRows.Scan(&start, &end); err != nil {
				continue
			}

			// reconstruire les timestamps avec la bonne date
			start = time.Date(date.Year(), date.Month(), date.Day(), start.Hour(), start.Minute(), 0, 0, time.Local)
			end = time.Date(date.Year(), date.Month(), date.Day(), end.Hour(), end.Minute(), 0, 0, time.Local)

			// Étape 3 : récupérer indisponibilités et rendez-vous
			queryUnavail := `
				SELECT start_time, end_time
				FROM unavailabilities
				WHERE user_id = $1 AND date = $2
			`
			unav, _ := db.Query(queryUnavail, agent.ID, dateStr)
			var blocks []TimeSlot
			for unav.Next() {
				var bStart, bEnd time.Time
				unav.Scan(&bStart, &bEnd)
				bStart = time.Date(date.Year(), date.Month(), date.Day(), bStart.Hour(), bStart.Minute(), 0, 0, time.Local)
				bEnd = time.Date(date.Year(), date.Month(), date.Day(), bEnd.Hour(), bEnd.Minute(), 0, 0, time.Local)
				blocks = append(blocks, TimeSlot{Start: bStart.Format(time.RFC3339), End: bEnd.Format(time.RFC3339)})
			}
			unav.Close()

			queryApp := `
				SELECT start_time, end_time
				FROM appointment
				WHERE host = $1 AND DATE(start_time) = $2 AND status = true
			`
			apps, _ := db.Query(queryApp, agent.ID, dateStr)
			for apps.Next() {
				var aStart, aEnd time.Time
				apps.Scan(&aStart, &aEnd)
				blocks = append(blocks, TimeSlot{Start: aStart.Format(time.RFC3339), End: aEnd.Format(time.RFC3339)})
			}
			apps.Close()

			// Étape 4 : découper les créneaux
			for t := start; t.Add(time.Duration(agent.Duration)*time.Minute).Before(end) || t.Add(time.Duration(agent.Duration)*time.Minute).Equal(end); t = t.Add(time.Duration(agent.Duration) * time.Minute) {
				slotStart := t
				slotEnd := t.Add(time.Duration(agent.Duration) * time.Minute)

				conflict := false
				for _, b := range blocks {
					bStart, _ := time.Parse(time.RFC3339, b.Start)
					bEnd, _ := time.Parse(time.RFC3339, b.End)
					if slotStart.Before(bEnd) && slotEnd.After(bStart) {
						conflict = true
						break
					}
				}
				if !conflict {
					slots = append(slots, TimeSlot{
						Start: slotStart.Format("15:04"),
						End:   slotEnd.Format("15:04"),
						Label: fmt.Sprintf("%s - %s", slotStart.Format("15:04"), slotEnd.Format("15:04")),
					})
				}
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(slots)
}

// Structure pour capturer l'unique_name envoyé dans la requête
type AnonymousAppointmentPayload struct {
	Unique_name string `json:"unique_name"`
	StartTime   string `json:"start_time"`
	EndTime     string `json:"end_time"`
	SchoolID    int    `json:"school_id"`
}

// Fonction pour générer un identifiant unique
func GenerateUniqueToken() (string, error) {
	bytes := make([]byte, 16)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// Fonction pour envoyer un email
func SendEmail(config *config.Config, to, subject, body string) error {
	fmt.Println("DEBUGG : SendEmail function called")
	from := config.SMTPUser
	password := config.SMTPPassword
	smtpHost := config.SMTPHost
	smtpPort := config.SMTPPort

	// Construction du message email avec headers MIME
	msg := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=\"UTF-8\"\r\n\r\n%s",
		from, to, subject, body)

	auth := smtp.PlainAuth("", from, password, smtpHost)

	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{to}, []byte(msg))
	if err != nil {
		fmt.Printf("DEBUGG : Failed to send email: %v\n", err)
		return err
	}
	fmt.Println("DEBUGG : Email sent successfully")
	return nil
}

// Fonction pour valider un rendez-vous anonyme
func ValidateAnonymousAppointment(w http.ResponseWriter, r *http.Request) {
	fmt.Println("DEBUGG : ValidateAnonymousAppointmentHandler called")
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Token is missing", http.StatusBadRequest)
		return
	}

	// Rechercher le rendez-vous correspondant au token
	var appointment struct {
		ID          int    `json:"id"`
		Host        string `json:"host"`
		unique_name string `json:"unique_name"`
		School      string `json:"school"`
		Title       string `json:"title"`
		StartTime   string `json:"start_time"`
		EndTime     string `json:"end_time"`
		Token       string `json:"token"`
		Status      string `json:"status"`
	}
	query := `
	SELECT
	a.id,
	h.unique_name AS host_unique_name, 
	g.unique_name AS guest_unique_name, 
	a.title, 
	a.start_time, 
	a.end_time,
	s.name AS school,
	a.token,
	a.status
	FROM 
	appointment a
	JOIN 
	usr g 
	ON a.guest = g.id
	JOIN 
	usr h 
	ON a.host = h.id
	JOIN 
	school s 
	ON a.school = s.id
	WHERE 
	a.token = $1`

	err := database.GetConn().QueryRow(query, token).Scan(
		&appointment.ID,
		&appointment.Host,
		&appointment.unique_name,
		&appointment.Title,
		&appointment.StartTime,
		&appointment.EndTime,
		&appointment.School,
		&appointment.Token,
		&appointment.Status)
	if err == sql.ErrNoRows {
		http.Error(w, "Invalid token or appointment not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Failed to query database: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Renvoyer les détails du rendez-vous en réponse JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(appointment)
}

// Fonction pour créer un rendez-vous anonyme
func CreateAnonymousAppointment(w http.ResponseWriter, r *http.Request) {
	fmt.Println("DEBUGG : CreateAnonymousAppointment function called")
	// Décoder les données de la requête
	var payload AnonymousAppointmentPayload
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil || payload.Unique_name == "" {
		http.Error(w, "Échec du décodage du corps de la requête", http.StatusBadRequest)
		return
	}
	fmt.Printf("DEBUGG : Payload received: %+v\n", payload)

	// Générer un token unique
	token, err := GenerateUniqueToken()
	if err != nil {
		http.Error(w, "Échec de la génération du token unique", http.StatusInternalServerError)
		return
	}

	// Vérifier si l'utilisateur existe déjà
	var userID int
	var userExists bool
	userExistsQuery := `
	SELECT EXISTS (
		SELECT 1
		FROM usr
		WHERE unique_name ILIKE $1
	)`
	err = database.GetConn().QueryRow(userExistsQuery, payload.Unique_name).Scan(&userExists)
	if err != nil {
		http.Error(w, "Échec de la requête à la base de données: "+err.Error(), http.StatusInternalServerError)
		return
	}
	// Si l'utilisateur existe, on récupère son ID
	if userExists {
		userIDQuery := `
		SELECT id 
		FROM usr 
		WHERE unique_name ILIKE $1 
		ORDER BY id 
		DESC LIMIT 1`

		err = database.GetConn().QueryRow(userIDQuery, payload.Unique_name).Scan(&userID)
		if err != nil {
			http.Error(w, "Échec de la requête à la base de données: "+err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Printf("DEBUGG : User ID found: %d\n", userID)
	} else {
		// Si l'utilisateur n'existe pas, on le crée
		fmt.Printf("DEBUGG : L'utilisateur n'existe pas, création d'un nouvel utilisateur\n")
		userQuery := `
		INSERT INTO usr 
		(unique_name) 
		VALUES ($1)
		RETURNING id`

		err = database.GetConn().QueryRow(userQuery, payload.Unique_name).Scan(&userID)
		if err != nil {
			http.Error(w, "Échec de l'insertion des données dans la base de données: "+err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Printf("DEBUGG : New user created with ID: %d\n", userID)

		resourceQuery := `
		SELECT id
		FROM resource
		WHERE school = $1 
		AND name ILIKE '%guest%'
		`
		var resourceID int
		err = database.GetConn().QueryRow(resourceQuery, payload.SchoolID).Scan(&resourceID)
		if err != nil && err != sql.ErrNoRows {
			http.Error(w, "Échec de la requête à la base de données: "+err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Printf("DEBUGG : Resource ID found: %d\n", resourceID)
		// insérer le rôle de l'utilisateur
		userRoleQuery := `
		INSERT INTO 
		user_school_resource (user_id, school_id, resource_id) 
		VALUES ($1, $2, $3)`

		_, err = database.GetConn().Exec(userRoleQuery, userID, payload.SchoolID, resourceID)
		if err != nil {
			http.Error(w, "Échec de l'insertion des données dans la base de données: "+err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Printf("DEBUGG : User role inserted for user ID: %d\n", userID)
	}

	var status bool
	// Vérifier si l'utilisateur a déjà un rendez-vous non confirmé
	existsQuery := `
	SELECT EXISTS (
		SELECT 1
		FROM appointment
		WHERE guest = $1
		AND status = false
	)`
	err = database.GetConn().QueryRow(existsQuery, userID).Scan(&status)
	if err != nil {
		http.Error(w, "Erreur lors de la vérification de l'existence du rendez-vous: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if status {
		// Si l'utilisateur a déjà un rendez-vous, on renvoie une erreur
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{"error": "Vous avez déjà un rendez-vous non confirmé. Veuillez vérifier vos emails/spams."})
		return
	}
	if !status {
		// Calculer l'heure de fin d'un rendez-vous
		startTime, err := time.Parse("2006-01-02 15:04:05", payload.StartTime)
		if err != nil {
			http.Error(w, "Échec de l'analyse de l'heure de début: "+err.Error(), http.StatusBadRequest)
			return
		}
		// Ajoute 30 minutes à l'heure de début par défaut (@secretariat)
		endTime := startTime.Add(30 * time.Minute)

		// Requete pour trouver le ressourceID du secrétariat
		secretariatResourceName := "%secretariat%"

		resourceQuery := `
		SELECT id
		FROM resource
		WHERE school = $1 
		AND name ILIKE $2
		`
		var resourceID int
		err = database.GetConn().QueryRow(resourceQuery, payload.SchoolID, secretariatResourceName).Scan(&resourceID)
		if err != nil && err != sql.ErrNoRows {
			http.Error(w, "Échec de la requête à la base de données: "+err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Printf("DEBUGG : Resource ID found: %d\n", resourceID)

		// Requête pour trouver les utilisateurs ayant le rôle de secrétaire, liés à l’école, et disponibles
		hostQuery := `
		SELECT usr.id
		FROM usr
			JOIN user_school_resource usr_res ON usr.id = usr_res.user_id
			JOIN work_schedule ws ON usr.id = ws.user_id
		WHERE usr_res.school_id = $1
		AND usr_res.resource_id = $2
		AND ws.day_of_week = $3
		AND ws.start_time <= $4::time
		AND ws.end_time >= $5::time
		AND usr.id NOT IN (
			SELECT user_id
			FROM unavailabilities
			WHERE date = $6::date
			AND (start_time, end_time) OVERLAPS ($4::time, $5::time)
		)
		ORDER BY RANDOM()
		LIMIT 1`
		dayOfWeek := strings.ToLower(startTime.Weekday().String())
		dayMap := map[string]string{
			"monday":    "lundi",
			"tuesday":   "mardi",
			"wednesday": "mercredi",
			"thursday":  "jeudi",
			"friday":    "vendredi",
			"saturday":  "samedi",
			"sunday":    "dimanche",
		}
		day := dayMap[dayOfWeek]

		var hostID int
		fmt.Printf("DEBUGG : with params schoolID=%d, resourceID=%d, dayOfWeek=%s, startTime=%s, endTime=%s\n", payload.SchoolID, resourceID, dayOfWeek, startTime.Format("15:04"), endTime.Format("15:04"))
		err = database.GetConn().QueryRow(hostQuery, payload.SchoolID, resourceID, day, startTime, endTime, startTime).Scan(&hostID)
		if err != nil {
			// Utilisation du code d'erreur HTTP 422 (Unprocessable Entity) pour indiquer qu'aucun hôte n'est disponible
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnprocessableEntity)
			json.NewEncoder(w).Encode(map[string]string{"error": "Aucun hôte disponible trouvé : " + err.Error()})
			return
		}

		// Insérer le rendez-vous
		query := `
		INSERT INTO appointment (host, guest, start_time, end_time, school, title, resource, token) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id,
		(SELECT name FROM school WHERE id = $5)`

		_, err = database.GetConn().Exec(query, hostID, userID, payload.StartTime, endTime, payload.SchoolID, payload.Unique_name, resourceID, token)
		if err != nil {
			http.Error(w, "Échec de l'insertion des données dans la base de données: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Récupérer le nom de l'école
		var schoolName string
		schoolQuery := `
		SELECT name
		FROM school
		WHERE id = $1`
		err = database.GetConn().QueryRow(schoolQuery, payload.SchoolID).Scan(&schoolName)
		if err != nil {
			http.Error(w, "Échec de la requête à la base de données: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Générer le lien unique
		var emailUrl string
		if strings.Contains(r.Host, "localhost") {
			emailUrl = "http://localhost:9000/validate-appointment?"
		} else {
			emailUrl = "https://prometheus.hainaut-promsoc.be/validate-appointment?"
		}
		fullURL := fmt.Sprintf("%stoken=%s", emailUrl, token)
		// Corps du mail HTML

		body := fmt.Sprintf(`
		<!DOCTYPE html>
		<html lang="fr">
		<head>
		<meta charset="UTF-8">
		<title>Demande de confirmation de rendez-vous</title>
		</head>
		<body style="font-family: Arial, sans-serif; color: #333; padding: 20px;">
		<div style="max-width: 600px; margin: auto; border: 1px solid #eee; border-radius: 8px; overflow: hidden; background-color: #fafafa;">
			<div style="background-color: #26a69a; padding: 10px;">
			<img src="https://raw.githubusercontent.com/ydebaere/Prom/main/assets/Prometheus_logo.png" alt="Logo E-Cale" style="height: 40px;">
			</div>
			<div style="padding: 20px;">
			<h2 style="color: #26a69a;">Demande de confirmation de rendez-vous</h2>
			<p><strong>Identité :</strong> %s</p>
			<p><strong>Date :</strong> %s</p>
			<p><strong>Début :</strong> %s</p>
			<p><strong>Fin :</strong> %s</p>
			<p><strong>Établissement :</strong> %s</p>
			<p style="text-align: center;">
				<a href="%s"
				style="background-color: #26a69a; color: white; padding: 12px 20px; text-decoration: none; border-radius: 4px; display: inline-block;">
				Confirmer mon rendez-vous
				</a>
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
			payload.Unique_name,
			startTime.Format("02-01-2006"),
			startTime.Format("15:04"),
			endTime.Format("15:04"),
			schoolName,
			fullURL,
		)
		// Titre du mail
		mailTitle := "Demande de confirmation de rendez-vous"

		// Envoi du mail
		err = SendEmail(config.LoadConfig(), payload.Unique_name, mailTitle, body)
		if err != nil {
			http.Error(w, "Échec de l'envoi de l'email : "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Répondre avec un statut de succès et les détails du rendez-vous
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message":          "Rendez-vous anonyme créé et email envoyé avec succès",
			"unique_name":      payload.Unique_name,
			"start_time":       startTime.Format("2006-01-02 15:04:05"),
			"end_time":         endTime.Format("2006-01-02 15:04:05"),
			"school_name":      schoolName,
			"school_id":        payload.SchoolID,
			"token":            token,
			"confirmation_url": fullURL,
			"host_id":          hostID,
			"resource_id":      resourceID,
			"guest_id":         userID,
		})
	} else {
		// Si l'utilisateur a déjà un rendez-vous, on renvoie une erreur
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(map[string]string{"error": "Vous avez déjà un rendez-vous en cours. Veuillez contacter le secrétariat pour plus d'informations."})
		return
	}
}

// Fonction pour exporter le calendrier au format ICS
func ExportCalendar(w http.ResponseWriter, r *http.Request) {
	uniqueName := r.URL.Query().Get("unique_name")
	if uniqueName == "" {
		http.Error(w, "Missing unique_name", http.StatusBadRequest)
		return
	}

	query := `
		SELECT a.id, a.title, a.start_time, a.end_time
		FROM appointment a
		JOIN usr g ON a.guest = g.id
		JOIN usr h ON a.host = h.id
		WHERE g.unique_name = $1 OR h.unique_name = $1
		ORDER BY a.start_time DESC`
	rows, err := database.GetConn().Query(query, uniqueName)
	if err != nil {
		http.Error(w, "Failed to query database: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type Appointment struct {
		ID    int
		Title string
		Start string
		End   string
	}
	var appointments []Appointment

	for rows.Next() {
		var appt Appointment
		err := rows.Scan(&appt.ID, &appt.Title, &appt.Start, &appt.End)
		if err != nil {
			http.Error(w, "Failed to scan row: "+err.Error(), http.StatusInternalServerError)
			return
		}
		appointments = append(appointments, appt)
	}

	cal := ical.NewCalendar()
	cal.SetMethod(ical.MethodRequest)

	for _, appt := range appointments {
		event := cal.AddEvent(strconv.Itoa(appt.ID))
		event.SetSummary(appt.Title)
		event.SetDescription("Rendez-vous Prometheus")
		event.SetLocation("Avenue du Tir 10 Mons, Belgique")

		startTime, err := time.Parse(time.RFC3339, appt.Start)
		if err != nil {
			http.Error(w, "Failed to parse start time: "+err.Error(), http.StatusInternalServerError)
			return
		}
		endTime, err := time.Parse(time.RFC3339, appt.End)
		if err != nil {
			http.Error(w, "Failed to parse end time: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Convertir en fuseau Europe/Paris
		event.SetStartAt(startTime.Add(-2 * time.Hour))
		event.SetEndAt(endTime.Add(-2 * time.Hour))
	}

	w.Header().Set("Content-Type", "text/calendar")
	w.Header().Set("Content-Disposition", "attachment; filename=mon_calendrier.ics")
	w.Write([]byte(cal.Serialize()))
}

// //// AUTOCOMPLETE : GetAzureUsers function
func GetAzureUsers(w http.ResponseWriter, r *http.Request) {
	fmt.Println("DEBUGG : GetAzureUsers function called")

	groupID := "bf19de93-1022-4a91-8335-598cb585d15d"

	query := r.URL.Query().Get("search")
	if query == "" {
		http.Error(w, "Missing query", http.StatusBadRequest)
		return
	}

	token, err := getAzureToken()
	if err != nil {
		http.Error(w, "Failed to get token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	users, err := getGroupMembersFiltered(groupID, token, query)
	if err != nil {
		http.Error(w, "Failed to get users: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(users)
	if err != nil {
		http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
func getAzureToken() (string, error) {
	fmt.Println("DEBUGG : Starting getAzureToken function")

	azureCfg := config.LoadAzureConfig()
	if azureCfg == nil {
		return "", fmt.Errorf("failed to load Azure config")
	}

	data := url.Values{}
	data.Set("client_id", azureCfg.ClientID)
	data.Set("client_secret", azureCfg.Secret)
	data.Set("grant_type", "client_credentials")
	data.Set("scope", "https://graph.microsoft.com/.default")

	tokenURL := fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", azureCfg.TenantID)
	resp, err := http.PostForm(tokenURL, data)
	if err != nil {
		fmt.Printf("DEBUGG : HTTP Request Error: %v\n", err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("DEBUGG : Token request failed: %s\n", string(body))
		return "", fmt.Errorf("token request failed: %s", resp.Status)
	}

	var result struct {
		AccessToken string `json:"access_token"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("DEBUGG : JSON Decode Error: %v\n", err)
		return "", err
	}
	return result.AccessToken, nil
}
func getGroupMembersFiltered(groupID, token, filter string) ([]map[string]string, error) {
	baseURL := fmt.Sprintf("https://graph.microsoft.com/v1.0/groups/%s/members?$select=id,mail,givenName,surname", groupID)
	users := []map[string]string{}

	for url := baseURL; url != ""; {
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			fmt.Printf("DEBUGG : Failed to read response body: %v\n", err)
			return nil, err
		}

		var result struct {
			Value []struct {
				ID        string `json:"id"`
				Mail      string `json:"mail"`
				GivenName string `json:"givenName"`
				Surname   string `json:"surname"`
			} `json:"value"`
			NextLink string `json:"@odata.nextLink"`
		}

		if err := json.Unmarshal(body, &result); err != nil {
			fmt.Printf("DEBUGG : JSON Decode Error: %v\n", err)
			return nil, err
		}

		for _, u := range result.Value {
			if u.Mail == "" {
				continue
			}
			if strings.HasPrefix(strings.ToLower(u.Mail), strings.ToLower(filter)) ||
				strings.HasPrefix(strings.ToLower(u.GivenName), strings.ToLower(filter)) ||
				strings.HasPrefix(strings.ToLower(u.Surname), strings.ToLower(filter)) {
				users = append(users, map[string]string{
					"id":          u.ID,
					"mail":        u.Mail,
					"given_name":  u.GivenName,
					"family_name": u.Surname,
				})
			}
		}

		url = result.NextLink // continuer si `@odata.nextLink` est présent
	}

	fmt.Printf("DEBUGG : Filtered Users: %+v\n", users)
	return users, nil
}
