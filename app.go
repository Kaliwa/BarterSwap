package main

// App is the business layer: it validates input, enforces the domain rules and
// orchestrates the Store. It holds no HTTP nor SQL concerns.
type App struct {
	store *Store
}

// NewApp wires the business layer to its Store.
func NewApp(store *Store) *App {
	return &App{store: store}
}
