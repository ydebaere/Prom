package routes

import (
	"net/http"

	"backend/internal/handlers/appointment"
	"backend/internal/handlers/resource"
	"backend/internal/handlers/school"
	"backend/internal/handlers/service"
	"backend/internal/handlers/unavailabilities"
	"backend/internal/handlers/user"
	"backend/internal/handlers/userSchoolResource"
	workschedule "backend/internal/handlers/workSchedule"
	"backend/internal/middleware"
)

func Routes() {

	// Routes non sécurisées
	// Horaires de rendez-vous anonyme
	http.HandleFunc("/anonymous-schedules", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		// cfg := config.LoadConfig()
		service.GetAnonymousAvailableSlots(w, r)
	})
	// Prise de rendez-vous anonyme
	http.HandleFunc("/anonymous-appointment", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		// cfg := config.LoadConfig()
		service.CreateAnonymousAppointment(w, r)
	})

	// Informations apres prise de rendez-vous
	http.HandleFunc("/validate-appointment", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			service.ValidateAnonymousAppointment(w, r)
		case "POST":
			appointment.ConfirmAppointment(w, r)
		case "DELETE":
			appointment.DeleteAppointmentByToken(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Exportation du calendrier
	http.HandleFunc(("/calendar-ics"), func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			service.ExportCalendar(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Routes sécurisées
	// Récupération des slots disponibles
	http.HandleFunc("/availability", middleware.WithJWTMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		appointment.GetAvailableSlots(w, r)
	}))

	// Gestion des indisponibilités
	http.HandleFunc("/unavailabilities", middleware.WithJWTMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			if r.URL.Query().Get("userID") != "" {
				unavailabilities.GetUnavailabilitiesByUser(w, r)
			} else {
				unavailabilities.GetUnavailabilities(w, r)
			}
		case "POST":
			unavailabilities.CreateUnavailability(w, r)
		case "DELETE":
			unavailabilities.DeleteUnavailability(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	// Gestion des ressources
	http.HandleFunc("/resources", middleware.WithJWTMiddleware(func(w http.ResponseWriter, r *http.Request) {
		schoolID := r.URL.Query().Get("schoolID")
		switch r.Method {
		case "GET":
			if schoolID != "" {
				resource.FetchResourcesBySchool(w, r)
			} else {
				resource.FetchResources(w, r)
			}
		case "POST":
			resource.CreateResource(w, r)
		case "PUT":
			resource.UpdateResource(w, r)
		case "DELETE":
			resource.DeleteResource(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	// Gestion des utilisateurs
	http.HandleFunc("/users", middleware.WithJWTMiddleware(func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Query().Get("userID")
		switch r.Method {
		case "GET":
			if userID != "" {
				user.FetchUser(w, r)
			} else {
				user.FetchUsers(w, r)
			}
		case "POST":
			user.CreateUser(w, r)
		case "PUT":
			user.UpdateUser(w, r)
		case "DELETE":
			user.DeleteUser(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	// Gestion des écoles
	http.HandleFunc("/schools", middleware.WithJWTMiddleware((func(w http.ResponseWriter, r *http.Request) {
		schoolID := r.URL.Query().Get("schoolID")
		dirID := r.URL.Query().Get("dirID")
		switch r.Method {
		case "GET":
			if schoolID != "" {
				school.FetchSchool(w, r)
			} else {
				if dirID != "" {
					school.FetchSchoolByDirector(w, r)
				} else {
					school.FetchSchools(w, r)
				}
			}
		case "POST":
			school.CreateSchool(w, r)
		case "PUT":
			school.UpdateSchool(w, r)
		case "DELETE":
			school.DeleteSchool(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}

	})))

	// Gestion des rendez-vous
	http.HandleFunc("/appointments", middleware.WithJWTMiddleware(func(w http.ResponseWriter, r *http.Request) {
		appointmentID := r.URL.Query().Get("appointmentID")
		userID := r.URL.Query().Get("userID")
		schoolID := r.URL.Query().Get("schoolID")
		// cfg := config.LoadConfig()

		switch r.Method {
		case "GET":
			if appointmentID == "" && schoolID == "" && userID == "" {
				appointment.FetchAppointments(w, r)
			} else if appointmentID != "" && schoolID == "" && userID == "" {
				appointment.FetchAppointment(w, r)
			} else if appointmentID == "" && schoolID != "" && userID == "" {
				appointment.FetchAppointmentsBySchoolID(w, r)
			} else if appointmentID == "" && schoolID == "" && userID != "" {
				appointment.FetchAppointmentsByUserID(w, r)
			}

		case "POST":
			appointment.CreateAppointment(w, r)
		case "PUT":
			appointment.UpdateAppointment(w, r)
		case "DELETE":
			appointment.DeleteAppointment(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	// Gestion des ressources scolaires des utilisateurs
	http.HandleFunc("/user-school-resource", middleware.WithJWTMiddleware(func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Query().Get("userID")
		schoolID := r.URL.Query().Get("schoolID")
		resourceID := r.URL.Query().Get("resourceID")
		switch r.Method {
		case "GET":
			if schoolID != "" && userID == "" && resourceID == "" {
				userSchoolResource.FetchUserSchoolResourceWithSchoolID(w, r)
			} else if resourceID != "" && userID == "" && schoolID == "" {
				userSchoolResource.FetchUserSchoolResourceWithResourceID(w, r)
			} else if userID != "" && schoolID == "" && resourceID == "" {
				userSchoolResource.FetchUserSchoolResourceWithUserID(w, r)
			} else if schoolID != "" && resourceID != "" && userID == "" {
				userSchoolResource.FetchUserFromUSR(w, r)
			} else {
				userSchoolResource.FetchUserSchoolResource(w, r)
			}
		case "POST":
			userSchoolResource.AddUserSchoolResource(w, r)
		case "PUT":
			userSchoolResource.UpdateUserSchoolResource(w, r)
		case "DELETE":
			userSchoolResource.DeleteUserSchoolResource(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	// Gestion des emplois du temps
	http.HandleFunc("/workschedule", middleware.WithJWTMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			workschedule.GetWorkSchedulesByUser(w, r)
		case "POST":
			workschedule.CreateWorkSchedule(w, r)
		case "PUT":
			workschedule.UpdateWorkSchedule(w, r)
		case "DELETE":
			workschedule.DeleteWorkSchedule(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	// Interrogation de Azure AD
	http.HandleFunc("/get-azure-users", middleware.WithJWTMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			service.GetAzureUsers(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
}
