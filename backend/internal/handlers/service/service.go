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
	"os"
	"strings"
	"time"

	"backend/internal/config"
	"backend/internal/database"

	ical "github.com/arran4/golang-ical"

	"strconv"
)

// Structure pour capturer l'unique_name envoyé dans la requête
type AnonymousAppointmentRequest struct {
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

	fmt.Println("DEBUGG : Email sent successfully")
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{to}, []byte(msg))
	if err != nil {
		fmt.Printf("DEBUGG : Failed to send email: %v\n", err)
		return err
	}
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
func CreateAnonymousAppointment(config *config.Config, w http.ResponseWriter, r *http.Request) {
	fmt.Println("DEBUGG : CreateAnonymousAppointment function called")
	// Décoder les données de la requête
	var req AnonymousAppointmentRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.Unique_name == "" {
		http.Error(w, "Invalid request body or unique_name missing", http.StatusBadRequest)
		return
	}

	// Générer un token unique
	token, err := GenerateUniqueToken()
	if err != nil {
		http.Error(w, "Failed to generate unique token", http.StatusInternalServerError)
		return
	}

	var userID int

	// Vérifier si l'utilisateur existe déjà
	var userExists bool
	userExistsQuery := `
	SELECT EXISTS (
		SELECT 1
		FROM usr
		WHERE unique_name = $1
	)`
	err = database.GetConn().QueryRow(userExistsQuery, req.Unique_name).Scan(&userExists)
	if err != nil {
		http.Error(w, "Failed to query database: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if userExists {
		userIDQuery := `
		SELECT id 
		FROM usr 
		WHERE unique_name = $1 
		ORDER BY id 
		DESC LIMIT 1`

		err = database.GetConn().QueryRow(userIDQuery, req.Unique_name).Scan(&userID)
		if err != nil {
			http.Error(w, "Failed to query database: "+err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Printf("DEBUGG : User ID found: %d\n", userID)
	} else {
		fmt.Printf("DEBUGG : User does not exist, creating new user\n")
		userQuery := `
		INSERT INTO usr 
		(unique_name) 
		VALUES ($1)`

		_, err = database.GetConn().Exec(userQuery, req.Unique_name)
		if err != nil {
			http.Error(w, "Failed to insert data into database: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Récupérer l'ID de l'utilisateur nouvellement inséré
		userIDQuery := `
		SELECT id 
		FROM usr 
		WHERE unique_name = $1 
		ORDER BY id 
		DESC LIMIT 1`

		err = database.GetConn().QueryRow(userIDQuery, req.Unique_name).Scan(&userID)
		if err != nil {
			http.Error(w, "Failed to query database: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// insérer le rôle de l'utilisateur
		userRoleQuery := `
		INSERT INTO 
		user_school_resource (user_id, school_id, resource_id) 
		VALUES ($1, $2, $3)`

		_, err = database.GetConn().Exec(userRoleQuery, userID, req.SchoolID, 4)
		if err != nil {
			http.Error(w, "Failed to insert data into database: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Calculer l'heure de fin d'un rendez-vous
	startTime, err := time.Parse("2006-01-02 15:04:05", req.StartTime)
	if err != nil {
		http.Error(w, "Failed to parse start time: "+err.Error(), http.StatusBadRequest)
		return
	}
	endTime := startTime.Add(30 * time.Minute) // Ajoute 30 minutes à l'heure de début par défaut (@secretariat)

	// Requête pour trouver les utilisateurs ayant le rôle de secrétaire, liés à l’école, et disponibles
	hostQuery := `
	SELECT usr.id
	FROM usr
	JOIN user_school_resource usr_res ON usr.id = usr_res.user_id
	JOIN work_schedule ws ON usr.id = ws.user_id
	WHERE usr_res.school_id = $1
	AND usr_res.resource_id = 0
	AND ws.day_of_week = $2
	AND ws.start_time <= $3::time
	OR ws.end_time >= $4::time
	AND usr.id NOT IN (
		SELECT user_id
		FROM unavailabilities
		WHERE date = $5::date
		AND ($3::time, $4::time) OVERLAPS (start_time, end_time)
	)
	ORDER BY RANDOM()
	LIMIT 1;
	`

	dayOfWeek := strings.ToLower(startTime.Weekday().String())
	var hostID int
	err = database.GetConn().QueryRow(hostQuery, req.SchoolID, dayOfWeek, startTime.Format("15:04:05"), endTime.Format("15:04:05"), startTime.Format("2006-01-02")).Scan(&hostID)
	if err != nil {
		fmt.Printf("hostQuery : %s\n", hostQuery)
		fmt.Printf("With params : %d, %s, %s, %s, %s\n", req.SchoolID, dayOfWeek, startTime.Format("15:04:05"), endTime.Format("15:04:05"), startTime.Format("2006-01-02"))
		http.Error(w, "No available host found: "+err.Error(), http.StatusNotFound)
		return
	}

	fmt.Printf("DEBUGG : Host ID found: %d\n", hostID)

	// Insérer le rendez-vous
	query := `
		INSERT INTO appointment (host, guest, start_time, end_time, school, title, resource, token) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err = database.GetConn().Exec(query, hostID, userID, req.StartTime, endTime, req.SchoolID, req.Unique_name, 2, token)
	if err != nil {
		http.Error(w, "Failed to insert data into database: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Récupérer le nom de l'école
	var schoolName string
	schoolQuery := `
	SELECT name
	FROM school
	WHERE id = $1`
	err = database.GetConn().QueryRow(schoolQuery, req.SchoolID).Scan(&schoolName)
	if err != nil {
		http.Error(w, "Failed to query database: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Générer le lien unique
	link := fmt.Sprintf("http://localhost:9000/validate-appointment?token=%s", token)

	// Corps du mail HTML
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
			<img src="https://upload.wikimedia.org/wikipedia/commons/2/29/Logo_Province_de_Hainaut.png" alt="Logo E-Cale" style="height: 40px;">
			</div>
			<div style="padding: 20px;">
			<h2 style="color: #26a69a;">Confirmation de rendez-vous</h2>
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
		req.Unique_name,
		startTime.Format("02-01-2006"),
		startTime.Format("15:04"),
		endTime.Format("15:04"),
		schoolName,
		link,
	)

	// Envoi du mail HTML
	err = SendEmail(config, req.Unique_name, "Votre rendez-vous via Prometheus", body)
	if err != nil {
		http.Error(w, "Failed to send email: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Répondre avec un statut de succès
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Anonymous appointment created and email sent successfully"})
}

// Fonction pour confirmer un rendez-vous
func ConfirmAppointment(w http.ResponseWriter, r *http.Request) {
	// Récupérer le token de la requête
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Token is missing", http.StatusBadRequest)
		return
	}

	// Rechercher le rendez-vous correspondant au token
	var appointmentID int
	query := `
	SELECT id 
	FROM appointment 
	WHERE token = $1`
	err := database.GetConn().QueryRow(query, token).Scan(&appointmentID)
	if err == sql.ErrNoRows {
		http.Error(w, "Invalid token or appointment not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Failed to query database: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Mettre à jour le statut du rendez-vous
	updateQuery := `
	UPDATE appointment 
	SET status = 'true' 
	WHERE id = $1`
	_, err = database.GetConn().Exec(updateQuery, appointmentID)
	if err != nil {
		http.Error(w, "Failed to update appointment status: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Appointment confirmed successfully",
		"status":  http.StatusOK,
	})
}

// Fonction pour récupérer les groupes ou rôles d'un utilisateur via Microsoft Graph
func GetUserRolesFromGraph(accessToken, userID string) (map[string]interface{}, error) {
	// URL de l'API Microsoft Graph pour récupérer les groupes de l'utilisateur
	url := fmt.Sprintf("https://graph.microsoft.com/v1.0/users/%s/memberOf", userID)
	// Créer une requête HTTP GET
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Ajouter le token d'accès dans l'en-tête Authorization
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	// Effectuer la requête
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Lire la réponse
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	// Vérifier le statut de la réponse
	if resp.StatusCode != http.StatusOK {
		// Si le statut n'est pas 200, retourner une erreur
		// et afficher le corps de la réponse pour le débogage
		fmt.Printf("DEBUGG : Status code: %d\n", resp.StatusCode)
		fmt.Printf("DEBUGG : Response body: %s\n", string(body))
		return nil, fmt.Errorf("failed to fetch user roles, status: %s, response: %s", resp.Status, string(body))
	}

	// Décoder la réponse JSON
	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response JSON: %v", err)
	}

	return result, nil
}

var appointment struct {
	ID        int       `json:"id"`
	Guest     string    `json:"guest"`
	Host      string    `json:"host"`
	Type      int       `json:"type"`
	School    string    `json:"school"`
	Title     string    `json:"title"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

// Fonction pour exporter le calendrier au format ICS
func ExportCalendarICSFile(w http.ResponseWriter, r *http.Request) {
	unique_name := r.URL.Query().Get("unique_name")
	if unique_name == "" {
		http.Error(w, "unique_name is missing", http.StatusBadRequest)
		return
	}
	// Récupérer les rendez-vous de l'utilisateur
	query := `
	SELECT
		a.id,
		a.guest,
		a.host,
		a.type,
		a.school,
		a.title,
		a.start_time,
		a.end_time
		FROM
		appointment a
		JOIN usr g ON a.guest = g.id
		JOIN usr h ON a.host = h.id
		WHERE g.unique_name = $1
		ORDER BY a.start_time DESC`
	// Exécuter la requête
	err := database.GetConn().QueryRow(query, unique_name).Scan(
		&appointment.ID,
		&appointment.Guest,
		&appointment.Host,
		&appointment.Type,
		&appointment.School,
		&appointment.Title,
		&appointment.StartTime,
		&appointment.EndTime)
	if err == sql.ErrNoRows {
		http.Error(w, "No appointments found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Failed to query database: "+err.Error(), http.StatusInternalServerError)
		return
	}
	// Créer le contenu du fichier ICS
	icsContent := fmt.Sprintf(`BEGIN:VCALENDAR
		VERSION:2.0
		PRODID:-//Ilulu//NONSGML v1.0//EN
		METHOD:PUBLISH
		BEGIN:VEVENT
		UID:%d
		SUMMARY:%s
		DTSTART;TZID=Europe/Paris:%s
		DTEND;TZID=Europe/Paris:%s
		DESCRIPTION:%s
		STATUS:CONFIRMED
		SEQUENCE:0
		BEGIN:VALARM
		TRIGGER:-PT15M
		DESCRIPTION:Reminder
		ACTION:DISPLAY
		END:VALARM
		END:VEVENT
		END:VCALENDAR`, appointment.ID, appointment.Title, appointment.StartTime.Format("20060102T150405"), appointment.EndTime.Format("20060102T150405"), appointment.Title)
	// Définir les en-têtes de la réponse
	w.Header().Set("Content-Type", "text/calendar")
	w.Header().Set("Content-Disposition", "attachment; filename=appointment.ics")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(icsContent)))
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Last-Modified", time.Now().Format(http.TimeFormat))
	w.Header().Set("ETag", fmt.Sprintf("%x", time.Now().Unix()))
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")

	fmt.Printf("ICS Content: %s\n", icsContent)
}

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

// / import calendrier
func ImportCalendar(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20) // Limite de 10 Mo

	file, handler, err := r.FormFile("icsFile")
	if err != nil {
		http.Error(w, "Erreur lors du téléchargement du fichier", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Optionnel : sauvegarder temporairement le fichier
	tempFile, err := os.CreateTemp("", "upload-*.ics")
	if err != nil {
		http.Error(w, "Erreur lors de la création du fichier temporaire", http.StatusInternalServerError)
		return
	}
	defer os.Remove(tempFile.Name()) // Nettoyage
	defer tempFile.Close()

	io.Copy(tempFile, file)
	tempFile.Seek(0, 0)

	cal, err := ical.ParseCalendar(tempFile)
	if err != nil {
		http.Error(w, "Erreur lors de la lecture du fichier ICS", http.StatusInternalServerError)
		return
	}

	for _, event := range cal.Events() {
		summary := event.GetProperty("SUMMARY") // Accès au résumé
		start := event.GetProperty("DTSTART")   // Accès à la date de début
		end := event.GetProperty("DTEND")       // Accès à la date de fin

		fmt.Printf("Event: %s, Start: %s, End: %s\n",
			summary.Value,
			start.Value,
			end.Value,
		)

		// Ici : insérez les données dans votre base (dans un modèle Appointment par ex.)
	}

	response := map[string]string{
		"message": fmt.Sprintf("Fichier %s traité avec succès", handler.Filename),
		"status":  "success",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
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
