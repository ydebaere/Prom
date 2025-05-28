# 📅 Prometheus

> **Projet de fin d'études – Bachelier en Informatique de Gestion**

## 🧩 Présentation

**Prometheus** est une application web complète destinée à faciliter la prise de rendez-vous entre différentes parties (ex. : élèves et écoles). Elle propose une gestion sécurisée des utilisateurs, des disponibilités et des réservations via une interface moderne développée avec **Quasar (Vue 3)**, et un serveur backend performant écrit en **Go**.

---

## 🚀 Fonctionnalités principales

- 🔐 Authentification sécurisée par **JWT**
- 👥 Gestion des utilisateurs avec rôles
- 📅 Création, visualisation et suppression de rendez-vous
- ⛔ Gestion des **indisponibilités**
- 📱 Interface responsive (mobile/desktop) avec **Quasar Framework**
- 🛠️ API REST sécurisée et structurée
- 🗃️ Persistance des données avec **PostgreSQL**

---

## 🗂️ Architecture du projet

```
project/
├── frontend/   # Application Vue 3 avec Quasar
├── backend/    # Serveur HTTP Go + API REST
├── .env        # Fichier de configuration (non versionné)
```

### Technologies clés

| Frontend         | Backend      | Base de données |
|------------------|--------------|------------------|
| Quasar (Vue 3)   | Go (Golang)  | PostgreSQL       |

---

## ⚙️ Installation et lancement

### 1. Prérequis

- Node.js (v16+ recommandé)
- Go (v1.18+)
- Quasar CLI (`npm install -g @quasar/cli`)
- PostgreSQL

---

### 2. Frontend (Quasar)

**Installation des dépendances :**

```bash
cd frontend
npm install
```

**Démarrage en mode développement :**

```bash
quasar dev
```

**Build production :**

```bash
quasar build
```

---

### 3. Backend (Go)

**Lancement en mode développement :**

```bash
cd backend
go run main.go
```

**Compilation pour production :**

```bash
GOOS=linux GOARCH=amd64 go build -o prometheus-api
```

**Redémarrage du service (si installé via systemd) :**

```bash
sudo systemctl restart prometheus-api.service
```

---

## 🔐 Configuration (.env)

Les informations sensibles (clés, tokens, connexions) sont stockées dans un fichier `.env` (non versionné).

**Exemple de contenu :**

```env
JWT_SECRET=your_secret_key
DATABASE_URL=postgres://user:password@host:port/dbname?sslmode=disable
```

---

## 🎓 Contexte académique

Ce projet a été réalisé dans le cadre de mon **travail de fin d’études** en **Bachelier en Informatique de Gestion**.  
Il illustre mes compétences dans les domaines suivants :

- Développement **Fullstack (Vue + Go)**
- Conception et architecture logicielle
- Gestion de projet et déploiement
- Sécurité des API et authentification
- Utilisation de conteneurs et configuration système

---

## 📌 À venir (Roadmap)

- 📆 Synchronisation avec des calendriers externes

---

