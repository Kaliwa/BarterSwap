package main

import "context"

// Storer is the storage contract the business layer depends on. It is defined
// here, on the consumer side, so App can be exercised in tests with a fake
// implementation instead of a real database. *Store satisfies it.
type Storer interface {
	// Users
	CreateUser(ctx context.Context, u User) (User, error)
	GetUser(ctx context.Context, id int) (User, error)
	UpdateUser(ctx context.Context, u User) (User, error)
	UserExists(ctx context.Context, id int) (bool, error)
	GetUserSkills(ctx context.Context, userID int) ([]Skill, error)
	SetUserSkills(ctx context.Context, userID int, skills []Skill) error

	// Services
	CreateService(ctx context.Context, in Service) (Service, error)
	GetService(ctx context.Context, id int) (Service, error)
	UpdateService(ctx context.Context, in Service) (Service, error)
	DeleteService(ctx context.Context, id int) error
	ListServices(ctx context.Context, f ServiceFilter) ([]Service, error)

	// Exchanges
	CreateExchange(ctx context.Context, serviceID, requesterID, ownerID int) (Exchange, error)
	GetExchange(ctx context.Context, id int) (Exchange, error)
	ListExchanges(ctx context.Context, userID int, status string) ([]Exchange, error)
	HasActiveExchange(ctx context.Context, serviceID int) (bool, error)
	UserBalance(ctx context.Context, userID int) (int, error)
	AcceptExchange(ctx context.Context, id int) (Exchange, error)
	RejectExchange(ctx context.Context, id int) (Exchange, error)
	CompleteExchange(ctx context.Context, id int) (Exchange, error)
	CancelExchange(ctx context.Context, id int) (Exchange, error)

	// Reviews & stats
	CreateReview(ctx context.Context, in Review) (Review, error)
	ListUserReviews(ctx context.Context, revieweeID int) ([]Review, error)
	ListServiceReviews(ctx context.Context, serviceID int) ([]Review, error)
	GetUserStats(ctx context.Context, userID int) (UserStats, error)
}

// App is the business layer: it validates input, enforces the domain rules and
// orchestrates the Storer. It holds no HTTP nor SQL concerns.
type App struct {
	store Storer
}

// NewApp wires the business layer to its storage.
func NewApp(store Storer) *App {
	return &App{store: store}
}
