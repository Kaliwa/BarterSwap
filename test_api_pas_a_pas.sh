#!/usr/bin/env bash
# =============================================================================
#  BarterSwap — démo API en mode PAS À PAS (curl)
# =============================================================================
#  Le script s'arrête avant chaque appel : tu lis l'action annoncée, tu appuies
#  sur [Entrée], l'appel part, tu commentes le résultat, puis tu continues.
#
#  Scénario : Alice (fournit) — Bob (consomme) — Carol (règle de conflit).
#
#  AVANT LA DÉMO — repartir d'une base VIERGE (ids SERIAL = 1, 2, 3...) :
#       docker compose down -v && docker compose up -d
#       sleep 3 && curl -s http://localhost:8080/health
#
#  LANCEMENT :
#       bash test_api_pas_a_pas.sh        # mode pas à pas (pause à chaque étape)
#       AUTO=1 bash test_api_pas_a_pas.sh # tout d'un trait (répétition/vérif)
# =============================================================================

BASE="http://localhost:8080"
H_JSON="Content-Type: application/json"

# --- Helpers d'affichage ------------------------------------------------------
# pause : affiche l'action à venir puis attend [Entrée].
# - ne bloque pas si l'entrée n'est pas un terminal (ex. sortie redirigée)
# - désactivable avec AUTO=1 pour tout dérouler d'un coup
pause() {
  echo
  echo "▶ $1"
  if [ -z "$AUTO" ] && [ -t 0 ]; then
    read -rp "   [Entrée] pour lancer... " _
  fi
}

sep() {
  echo
  echo "════════════════════════════════════════════════════════════"
  echo "  $1"
  echo "════════════════════════════════════════════════════════════"
  if [ -z "$AUTO" ] && [ -t 0 ]; then
    read -rp "   [Entrée] pour démarrer cette section... " _
  fi
}

# =============================================================================
sep "0. HEALTHCHECK"
# =============================================================================
pause "Vérifier que l'API répond"
curl -s -w "\nHTTP %{http_code}\n" "$BASE/health"


# =============================================================================
sep "1. UTILISATEURS"
# =============================================================================
pause "Créer Alice (id 1) — reçoit 10 crédits de bienvenue"
curl -s -w "\nHTTP %{http_code}\n" -X POST "$BASE/api/users" -H "$H_JSON" \
  -d '{"pseudo":"Alice","bio":"Développeuse","ville":"Lyon"}'

pause "Créer Bob (id 2)"
curl -s -w "\nHTTP %{http_code}\n" -X POST "$BASE/api/users" -H "$H_JSON" \
  -d '{"pseudo":"Bob","ville":"Lyon"}'

pause "Créer Carol (id 3)"
curl -s -w "\nHTTP %{http_code}\n" -X POST "$BASE/api/users" -H "$H_JSON" \
  -d '{"pseudo":"Carol","ville":"Paris"}'

pause "[ERREUR 400] pseudo vide"
curl -s -w "\nHTTP %{http_code}\n" -X POST "$BASE/api/users" -H "$H_JSON" \
  -d '{"pseudo":"   "}'

pause "Profil public d'Alice"
curl -s -w "\nHTTP %{http_code}\n" "$BASE/api/users/1"

pause "Alice modifie son profil (X-User-ID: 1)"
curl -s -w "\nHTTP %{http_code}\n" -X PUT "$BASE/api/users/1" -H "$H_JSON" \
  -H "X-User-ID: 1" \
  -d '{"pseudo":"Alice","bio":"Développeuse Go","ville":"Lyon"}'

pause "[ERREUR 403] Bob tente de modifier le profil d'Alice"
curl -s -w "\nHTTP %{http_code}\n" -X PUT "$BASE/api/users/1" -H "$H_JSON" \
  -H "X-User-ID: 2" \
  -d '{"pseudo":"Hacked"}'

pause "[ERREUR 401] modification sans header X-User-ID"
curl -s -w "\nHTTP %{http_code}\n" -X PUT "$BASE/api/users/1" -H "$H_JSON" \
  -d '{"pseudo":"Alice"}'

pause "[ERREUR 404] utilisateur inexistant"
curl -s -w "\nHTTP %{http_code}\n" "$BASE/api/users/9999"

pause "[ERREUR 400] identifiant non numérique"
curl -s -w "\nHTTP %{http_code}\n" "$BASE/api/users/abc"


# =============================================================================
sep "2. COMPÉTENCES"
# =============================================================================
pause "Alice déclare ses compétences (écrase tout)"
curl -s -w "\nHTTP %{http_code}\n" -X PUT "$BASE/api/users/1/skills" -H "$H_JSON" \
  -H "X-User-ID: 1" \
  -d '[{"nom":"Informatique","niveau":"expert"},{"nom":"Musique","niveau":"intermédiaire"}]'

pause "Lire les compétences d'Alice"
curl -s -w "\nHTTP %{http_code}\n" "$BASE/api/users/1/skills"

pause "[ERREUR 400] niveau invalide"
curl -s -w "\nHTTP %{http_code}\n" -X PUT "$BASE/api/users/1/skills" -H "$H_JSON" \
  -H "X-User-ID: 1" \
  -d '[{"nom":"Cuisine","niveau":"ninja"}]'


# =============================================================================
sep "3. SERVICES (ANNONCES)"
# =============================================================================
pause "S1 : Alice publie un service Informatique à 3 crédits (id 1)"
curl -s -w "\nHTTP %{http_code}\n" -X POST "$BASE/api/services" -H "$H_JSON" \
  -H "X-User-ID: 1" \
  -d '{"titre":"Dépannage PC","description":"Diagnostic et réparation","categorie":"Informatique","duree_minutes":60,"credits":3,"ville":"Lyon"}'

pause "S2 : service Informatique à 15 crédits (id 2) — servira au test crédits"
curl -s -w "\nHTTP %{http_code}\n" -X POST "$BASE/api/services" -H "$H_JSON" \
  -H "X-User-ID: 1" \
  -d '{"titre":"Formation Go complète","categorie":"Informatique","duree_minutes":300,"credits":15,"ville":"Lyon"}'

pause "S3 : service Informatique à 2 crédits (id 3) — servira aux règles échange"
curl -s -w "\nHTTP %{http_code}\n" -X POST "$BASE/api/services" -H "$H_JSON" \
  -H "X-User-ID: 1" \
  -d '{"titre":"Install Linux","categorie":"Informatique","duree_minutes":45,"credits":2,"ville":"Lyon"}'

pause "S4 : service Informatique à 4 crédits (id 4) — servira au test annulation"
curl -s -w "\nHTTP %{http_code}\n" -X POST "$BASE/api/services" -H "$H_JSON" \
  -H "X-User-ID: 1" \
  -d '{"titre":"Config réseau","categorie":"Informatique","duree_minutes":90,"credits":4,"ville":"Lyon"}'

pause "[ERREUR 400] catégorie sans compétence correspondante (Alice n'a pas 'Jardinage')"
curl -s -w "\nHTTP %{http_code}\n" -X POST "$BASE/api/services" -H "$H_JSON" \
  -H "X-User-ID: 1" \
  -d '{"titre":"Tonte pelouse","categorie":"Jardinage","duree_minutes":30,"credits":2}'

pause "[ERREUR 400] crédits <= 0"
curl -s -w "\nHTTP %{http_code}\n" -X POST "$BASE/api/services" -H "$H_JSON" \
  -H "X-User-ID: 1" \
  -d '{"titre":"Gratuit","categorie":"Informatique","duree_minutes":30,"credits":0}'

pause "Lister tous les services"
curl -s -w "\nHTTP %{http_code}\n" "$BASE/api/services"

pause "Filtre serveur : ?categorie=Informatique"
curl -s -w "\nHTTP %{http_code}\n" "$BASE/api/services?categorie=Informatique"

pause "Filtre serveur : ?ville=Lyon"
curl -s -w "\nHTTP %{http_code}\n" "$BASE/api/services?ville=Lyon"

pause "Filtre serveur : ?search=linux (titre/description)"
curl -s -w "\nHTTP %{http_code}\n" "$BASE/api/services?search=linux"

pause "Détail du service 1"
curl -s -w "\nHTTP %{http_code}\n" "$BASE/api/services/1"

pause "Alice met à jour S1"
curl -s -w "\nHTTP %{http_code}\n" -X PUT "$BASE/api/services/1" -H "$H_JSON" \
  -H "X-User-ID: 1" \
  -d '{"titre":"Dépannage PC à domicile","categorie":"Informatique","duree_minutes":60,"credits":3,"ville":"Lyon","actif":true}'

pause "[ERREUR 403] Bob tente de modifier le service d'Alice"
curl -s -w "\nHTTP %{http_code}\n" -X PUT "$BASE/api/services/1" -H "$H_JSON" \
  -H "X-User-ID: 2" \
  -d '{"titre":"Pirate","categorie":"Informatique","duree_minutes":60,"credits":1,"actif":true}'


# =============================================================================
sep "4. ÉCHANGES — PARCOURS NOMINAL (E1 : Bob demande S1)"
# =============================================================================
pause "Bob demande le service S1 -> échange 1 (pending)"
curl -s -w "\nHTTP %{http_code}\n" -X POST "$BASE/api/exchanges" -H "$H_JSON" \
  -H "X-User-ID: 2" \
  -d '{"service_id":1}'

pause "[ERREUR 400] Alice demande son propre service"
curl -s -w "\nHTTP %{http_code}\n" -X POST "$BASE/api/exchanges" -H "$H_JSON" \
  -H "X-User-ID: 1" \
  -d '{"service_id":1}'

pause "[ERREUR 400] Bob demande S2 (15 crédits > son solde) -> crédits insuffisants"
curl -s -w "\nHTTP %{http_code}\n" -X POST "$BASE/api/exchanges" -H "$H_JSON" \
  -H "X-User-ID: 2" \
  -d '{"service_id":2}'

pause "[ERREUR 401] demande sans X-User-ID"
curl -s -w "\nHTTP %{http_code}\n" -X POST "$BASE/api/exchanges" -H "$H_JSON" \
  -d '{"service_id":1}'

pause "[ERREUR 403] Bob (demandeur) tente d'accepter — seul le propriétaire peut"
curl -s -w "\nHTTP %{http_code}\n" -X PUT "$BASE/api/exchanges/1/accept" -H "X-User-ID: 2"

pause "Alice (propriétaire) accepte l'échange 1 -> accepted (Bob débité de 3)"
curl -s -w "\nHTTP %{http_code}\n" -X PUT "$BASE/api/exchanges/1/accept" -H "X-User-ID: 1"

pause "Stats de Bob : solde attendu 10 - 3 = 7"
curl -s -w "\nHTTP %{http_code}\n" "$BASE/api/users/2/stats"

pause "Détail de l'échange 1 (réservé aux participants)"
curl -s -w "\nHTTP %{http_code}\n" "$BASE/api/exchanges/1" -H "X-User-ID: 2"

pause "Compléter l'échange 1 -> completed (Alice créditée de 3)"
curl -s -w "\nHTTP %{http_code}\n" -X PUT "$BASE/api/exchanges/1/complete" -H "X-User-ID: 1"

pause "Stats d'Alice : solde attendu 10 + 3 = 13"
curl -s -w "\nHTTP %{http_code}\n" "$BASE/api/users/1/stats"

pause "Lister les échanges de Bob (filtre ?status=completed)"
curl -s -w "\nHTTP %{http_code}\n" "$BASE/api/exchanges?status=completed" -H "X-User-ID: 2"


# =============================================================================
sep "5. AVIS (sur l'échange 1 terminé)"
# =============================================================================
pause "Bob (demandeur) note l'échange terminé"
curl -s -w "\nHTTP %{http_code}\n" -X POST "$BASE/api/exchanges/1/review" -H "$H_JSON" \
  -H "X-User-ID: 2" \
  -d '{"note":5,"commentaire":"Super rapide, merci !"}'

pause "[ERREUR 409] Bob tente un second avis sur le même échange"
curl -s -w "\nHTTP %{http_code}\n" -X POST "$BASE/api/exchanges/1/review" -H "$H_JSON" \
  -H "X-User-ID: 2" \
  -d '{"note":1}'

pause "Avis reçus par Alice"
curl -s -w "\nHTTP %{http_code}\n" "$BASE/api/users/1/reviews"

pause "Avis sur le service 1"
curl -s -w "\nHTTP %{http_code}\n" "$BASE/api/services/1/reviews"


# =============================================================================
sep "6. RÈGLE : UN SEUL ÉCHANGE ACTIF PAR SERVICE (E2 sur S3)"
# =============================================================================
pause "Bob demande S3 -> échange 2 (pending)"
curl -s -w "\nHTTP %{http_code}\n" -X POST "$BASE/api/exchanges" -H "$H_JSON" \
  -H "X-User-ID: 2" \
  -d '{"service_id":3}'

pause "[ERREUR 409] Carol demande le même S3 (déjà un échange actif)"
curl -s -w "\nHTTP %{http_code}\n" -X POST "$BASE/api/exchanges" -H "$H_JSON" \
  -H "X-User-ID: 3" \
  -d '{"service_id":3}'

pause "[ERREUR 409] Alice tente de supprimer S3 qui a un échange actif"
curl -s -w "\nHTTP %{http_code}\n" -X DELETE "$BASE/api/services/3" -H "X-User-ID: 1"

pause "Alice (propriétaire) rejette la demande (échange 2) -> rejected"
curl -s -w "\nHTTP %{http_code}\n" -X PUT "$BASE/api/exchanges/2/reject" -H "X-User-ID: 1"

pause "Alice supprime S3 : maintenant possible -> 204 No Content"
curl -s -w "\nHTTP %{http_code}\n" -X DELETE "$BASE/api/services/3" -H "X-User-ID: 1"


# =============================================================================
sep "7. ANNULATION AVEC REMBOURSEMENT (E3 sur S4)"
# =============================================================================
pause "Bob demande S4 -> échange 3 (pending)"
curl -s -w "\nHTTP %{http_code}\n" -X POST "$BASE/api/exchanges" -H "$H_JSON" \
  -H "X-User-ID: 2" \
  -d '{"service_id":4}'

pause "Alice accepte l'échange 3 (Bob débité de 4)"
curl -s -w "\nHTTP %{http_code}\n" -X PUT "$BASE/api/exchanges/3/accept" -H "X-User-ID: 1"

pause "Stats de Bob : solde attendu 7 - 4 = 3"
curl -s -w "\nHTTP %{http_code}\n" "$BASE/api/users/2/stats"

pause "Bob annule l'échange 3 -> cancelled (remboursé de 4)"
curl -s -w "\nHTTP %{http_code}\n" -X PUT "$BASE/api/exchanges/3/cancel" -H "X-User-ID: 2"

pause "Stats de Bob : solde de nouveau à 7 (remboursement)"
curl -s -w "\nHTTP %{http_code}\n" "$BASE/api/users/2/stats"


# =============================================================================
sep "8. STATISTIQUES FINALES"
# =============================================================================
pause "Stats d'Alice (échanges terminés, note moyenne, crédits gagnés...)"
curl -s -w "\nHTTP %{http_code}\n" "$BASE/api/users/1/stats"

pause "Stats de Bob"
curl -s -w "\nHTTP %{http_code}\n" "$BASE/api/users/2/stats"

echo
echo "════════════════════════════════════════════════════════════"
echo "  FIN DE LA DÉMO"
echo "════════════════════════════════════════════════════════════"
