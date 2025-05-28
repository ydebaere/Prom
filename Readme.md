Ce projet est une application web complète composée de :

- 🔧 Un **frontend** développé avec [Quasar Framework (Vue.js)](https://quasar.dev/)
- 🖥️ Un **backend** en [Go (Golang)](https://golang.org/), servant d'API REST

---

## 📁 Structure du projet

Prometheus
    frontend/ # Interface utilisateur (SPA avec Quasar)
    backend/  # Serveur API en Go

---

## 🚀 Installation rapide

### 🔸 1. Prérequis

- [Node.js](https://nodejs.org/) (v16+ recommandé)
- [Quasar CLI](https://quasar.dev/start/pick-quasar-flavour)
- [Go](https://golang.org/doc/install) (v1.18+)
- Un fichier `.env` pour chaque environnement (voir ci-dessous)

---

### 🔸 2. Lancer le **frontend**

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

### 🔸 2. Lancer le **backend**

Dev ::
```bash
cd backend
go run main.go
```

Build ::
```bash
GOOS=linux GOARCH=amd64 go build -o prometheus-api
```