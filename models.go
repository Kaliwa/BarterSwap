package main

// welcomeCredits is the amount of time-credits granted at account creation,
// letting a new member request services before having rendered any.
const welcomeCredits = 10

// Credit-transaction types recorded in the credit journal.
const (
	txEarn   = "earn"
	txSpend  = "spend"
	txRefund = "refund"
)

// validNiveaux is the closed set of accepted skill levels.
var validNiveaux = map[string]bool{
	"débutant":      true,
	"intermédiaire": true,
	"expert":        true,
}

// User represents a member of the platform. CreditBalance is derived from the
// credit-transaction journal, not stored as a column.
type User struct {
	ID            int     `json:"id"`
	Pseudo        string  `json:"pseudo"`
	Bio           string  `json:"bio,omitempty"`
	Ville         string  `json:"ville,omitempty"`
	Skills        []Skill `json:"skills,omitempty"`
	CreditBalance int     `json:"credit_balance"`
	CreatedAt     string  `json:"created_at"`
}

// Skill is a competence a user can offer. A user may declare several; they are
// replaced wholesale on each PUT (no individual add).
type Skill struct {
	Nom    string `json:"nom"`
	Niveau string `json:"niveau"`
}

// validCategories is the closed set of service categories.
var validCategories = map[string]bool{
	"Informatique": true,
	"Jardinage":    true,
	"Bricolage":    true,
	"Cuisine":      true,
	"Musique":      true,
	"Langues":      true,
	"Sport":        true,
	"Tutorat":      true,
	"Déménagement": true,
	"Photographie": true,
	"Animalier":    true,
	"Couture":      true,
	"Autre":        true,
}

// Service is a public announcement: a member offers a service tied to one of
// their skills, priced in time-credits.
type Service struct {
	ID           int    `json:"id"`
	ProviderID   int    `json:"provider_id"`
	Titre        string `json:"titre"`
	Description  string `json:"description,omitempty"`
	Categorie    string `json:"categorie"`
	DureeMinutes int    `json:"duree_minutes"`
	Credits      int    `json:"credits"`
	Ville        string `json:"ville,omitempty"`
	Actif        bool   `json:"actif"`
	CreatedAt    string `json:"created_at"`
}

// ServiceFilter holds the optional, server-side filters for listing services.
type ServiceFilter struct {
	Categorie string
	Ville     string
	Search    string
}

// Exchange statuses and their lifecycle:
//
//	pending → accepted → completed
//	   ↓         ↓
//	rejected  cancelled
const (
	statusPending   = "pending"
	statusAccepted  = "accepted"
	statusRejected  = "rejected"
	statusCancelled = "cancelled"
	statusCompleted = "completed"
)

// Exchange is a reservation between a requester (who asks) and the service
// owner (who offers).
type Exchange struct {
	ID          int    `json:"id"`
	ServiceID   int    `json:"service_id"`
	RequesterID int    `json:"requester_id"`
	OwnerID     int    `json:"owner_id"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}
