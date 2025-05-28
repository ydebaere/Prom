## 📝 Présentation du projet

**Prometheus** est une application web fullstack conçue dans le cadre de mon projet de fin d'études en informatique de gestion.  
Elle permet de gérer la prise de rendez-vous, gérer des utilisateurs et des groupes.

Le projet est divisé en deux parties :
- Un **frontend** en Quasar (Vue.js) pour l'interface utilisateur
- Un **backend** en Go qui expose une API REST sécurisée


## ✨ Fonctionnalités principales

- Authentification des utilisateurs
- Visualisation de [données / graphiques / listes]
- API RESTful avec validation
- Gestion des erreurs

## ✅ Tests

Le backend a été testé via des requêtes Postman / Curl.  
Les cas de test principaux incluent :
- Authentification
- Création / lecture / mise à jour / suppression d'objets
- Gestion des erreurs

> Aucun framework de test automatique n’est utilisé

## 🔐 Sécurité

Les clés, tokens et paramètres sensibles sont gérés via des fichiers `.env` (non versionnés).  
Exemple de contenu possible :

API_PORT=8080
JWT_SECRET=your-secret-key

## 🧱 Architecture technique

- **Frontend (Quasar)** : Application SPA
- **Backend (Go)** : Serveur HTTP avec routes REST
- **Échange de données** : Format JSON


## 📦 Dépendances

- Quasar CLI v2.x (Vue 3)
- Go
- Node.js 
- 

## 🚀 Déploiement

### 🔸 1. **frontend**

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

### 🔸 2. **backend**

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

## 🎓 Contexte académique

Ce projet a été réalisé dans le cadre de mon **travail de fin d'études** en Bachelier en Informatique de Gestion.  
Il démontre mes compétences en :
- développement fullstack
- architecture d’application
- gestion de configuration et sécurité
