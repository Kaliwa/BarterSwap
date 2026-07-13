# BarterSwap — API d'échange de compétences

BarterSwap est une **banque de temps** : une API REST qui permet à des particuliers
d'échanger leurs compétences sans transaction monétaire. Chaque heure de service
rendue donne droit à une heure de service reçue, via un système de **crédits-temps**.

> Projet de fin de module. Écrit en **Go** avec la bibliothèque standard
> (`net/http`, `encoding/json`, `database/sql`, `context`) — sans ORM ni framework externe.

## Stack technique

- **Langage** : Go 1.22
- **Base de données** : PostgreSQL 16 (`database/sql`, driver `github.com/lib/pq`)
- **Conteneurisation** : Docker + Docker Compose

## Installation

### Avec Docker (recommandé)

```bash
git clone <url>
cd BarterSwap
docker compose up --build
```

L'API est disponible sur `http://localhost:8080`, la base PostgreSQL sur le port `5432`.

Vérifier que le service répond :

```bash
curl http://localhost:8080/health
# {"status":"ok"}
```

### En local (sans Docker)

```bash
cp .env.example .env   # ajuster DATABASE_URL si besoin
go mod tidy
go run .
```

## Authentification

Pas de système avancé : les endpoints qui modifient une ressource attendent un
header `X-User-ID` identifiant l'appelant. Un utilisateur ne peut modifier que
ses propres données.

## Endpoints

### Gestion des utilisateurs

| Méthode | Endpoint | Auth | Description |
|---------|----------|------|-------------|
| `GET`  | `/health`                  | –          | Vérifier que l'API répond |
| `POST` | `/api/users`               | –          | Créer un compte (10 crédits de bienvenue) |
| `GET`  | `/api/users/{id}`          | –          | Profil public d'un utilisateur |
| `PUT`  | `/api/users/{id}`          | `X-User-ID` | Modifier son profil |
| `GET`  | `/api/users/{id}/skills`   | –          | Compétences d'un utilisateur |
| `PUT`  | `/api/users/{id}/skills`   | `X-User-ID` | Définir ses compétences (écrase tout) |

Niveaux de compétence acceptés : `débutant`, `intermédiaire`, `expert`.

### Annonces de services

| Méthode | Endpoint | Auth | Description |
|---------|----------|------|-------------|
| `GET`    | `/api/services`      | –          | Liste des annonces (filtres optionnels) |
| `POST`   | `/api/services`      | `X-User-ID` | Créer une annonce |
| `GET`    | `/api/services/{id}` | –          | Détail d'une annonce |
| `PUT`    | `/api/services/{id}` | `X-User-ID` | Modifier son annonce |
| `DELETE` | `/api/services/{id}` | `X-User-ID` | Supprimer son annonce |

Filtres cumulables (côté serveur) : `?categorie=`, `?ville=`, `?search=`.

Catégories (liste fermée) : `Informatique`, `Jardinage`, `Bricolage`, `Cuisine`,
`Musique`, `Langues`, `Sport`, `Tutorat`, `Déménagement`, `Photographie`,
`Animalier`, `Couture`, `Autre`.

> Pour publier une annonce dans une catégorie, l'utilisateur doit avoir déclaré
> une compétence portant le même nom que la catégorie.

_Échanges, évaluations et statistiques : à venir._

## Exemples d'utilisation

Créer un compte :

```bash
curl -X POST http://localhost:8080/api/users \
  -H 'Content-Type: application/json' \
  -d '{"pseudo":"alice","bio":"Jardinière passionnée","ville":"Lyon"}'
# 201 → {"id":1,"pseudo":"alice",...,"credit_balance":10,"created_at":"..."}
```

Consulter un profil public :

```bash
curl http://localhost:8080/api/users/1
```

Mettre à jour son profil (authentifié) :

```bash
curl -X PUT http://localhost:8080/api/users/1 \
  -H 'X-User-ID: 1' -H 'Content-Type: application/json' \
  -d '{"pseudo":"alice","bio":"Nouvelle bio","ville":"Paris"}'
```

Définir ses compétences (remplace la liste existante) :

```bash
curl -X PUT http://localhost:8080/api/users/1/skills \
  -H 'X-User-ID: 1' -H 'Content-Type: application/json' \
  -d '[{"nom":"Jardinage","niveau":"expert"},{"nom":"Cuisine","niveau":"débutant"}]'
```

Publier une annonce de service :

```bash
curl -X POST http://localhost:8080/api/services \
  -H 'X-User-ID: 1' -H 'Content-Type: application/json' \
  -d '{"titre":"Taille de haies","categorie":"Jardinage","duree_minutes":90,"credits":2,"ville":"Lyon"}'
```

Rechercher / filtrer les annonces :

```bash
curl 'http://localhost:8080/api/services?categorie=Jardinage&ville=Lyon'
curl 'http://localhost:8080/api/services?search=haies'
```

## Tests

```bash
go test -v -cover ./...
```
