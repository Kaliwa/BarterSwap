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

## Endpoints

_(Tableau récapitulatif à compléter au fil de l'implémentation.)_

## Exemples d'utilisation

_(3-4 exemples complets avec curl à venir.)_

## Tests

```bash
go test -v -cover ./...
```
