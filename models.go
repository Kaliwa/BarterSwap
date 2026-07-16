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

// Review bounds for the 1-5 rating scale.
const (
	minNote = 1
	maxNote = 5
)

// Review is the rating left by the requester on a completed exchange. It is
// immutable: no update nor delete once posted, and one review per exchange.
type Review struct {
	ID          int    `json:"id"`
	ExchangeID  int    `json:"exchange_id"`
	ServiceID   int    `json:"service_id"`
	ReviewerID  int    `json:"reviewer_id"`
	RevieweeID  int    `json:"reviewee_id"`
	Note        int    `json:"note"`
	Commentaire string `json:"commentaire,omitempty"`
	CreatedAt   string `json:"created_at"`
}

// UserStats aggregates a user's activity: credit balance, completed exchanges
// (as requester or owner), active service announcements, average rating
// received and credits earned/spent through exchanges.
type UserStats struct {
	UserID             int     `json:"user_id"`
	CreditBalance      int     `json:"credit_balance"`
	CompletedExchanges int     `json:"completed_exchanges"`
	ActiveServices     int     `json:"active_services"`
	AverageRating      float64 `json:"average_rating"`
	ReviewCount        int     `json:"review_count"`
	TotalEarned        int     `json:"total_earned"`
	TotalSpent         int     `json:"total_spent"`
}

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
