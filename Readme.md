# 📅 Appointment Booking App

## 🧩 Présentation du projet

Cette application permet la gestion de rendez-vous entre des utilisateurs (par exemple, élèves et écoles). Elle permet l’authentification, la réservation, et la gestion des disponibilités via une interface utilisateur moderne basée sur Quasar, avec un backend robuste en Go.

## 🚀 Fonctionnalités principales

- Authentification JWT
- Gestion des utilisateurs et des rôles
- Création et consultation de rendez-vous
- Gestion des indisponibilités
- Interface responsive avec Quasar Framework
- API REST sécurisée
- Stockage des données avec PostgreSQL

## 🗂 Structure du projet

- **Frontend (Quasar)** : Application SPA
- **Backend (Go)** : Serveur HTTP avec routes REST
- **Échange de données** : Format JSON


## 🛠️ Installation / Build


### 🔸 1. Dépendances

- Quasar CLI v2.x (Vue 3)
- Go
- Node.js 

### 🔸 2. **frontend**

Dev ::
```bash
cd frontend
npm install
quasar dev
```

Build ::
```bash
npm run build
```

### 🔸 3. **backend**

Dev ::
```bash
cd backend
go run main.go
```

Build ::
```bash
GOOS=linux GOARCH=amd64 go build -o prometheus-api
```

Relancer service backend ::
```bash
sudo systemctl restart prometheus-api.service
```

## 🔐 Utilisation des fichiers `.env`

Les clés, tokens et paramètres sensibles sont gérés via des fichiers `.env` (non versionnés).  
Exemple de contenu possible :

JWT_SECRET=your_secret_key
DATABASE_URL=postgres://user:password@host:port/dbname?sslmode=disable





