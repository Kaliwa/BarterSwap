package main

import "context"

// fakeStore implements Storer with overridable function fields, so each test
// stubs exactly the calls it expects. An un-stubbed call panics: the recover
// middleware turns it into a 500, which makes the test fail loudly.
type fakeStore struct {
	createUser    func(ctx context.Context, u User) (User, error)
	getUser       func(ctx context.Context, id int) (User, error)
	updateUser    func(ctx context.Context, u User) (User, error)
	userExists    func(ctx context.Context, id int) (bool, error)
	getUserSkills func(ctx context.Context, userID int) ([]Skill, error)
	setUserSkills func(ctx context.Context, userID int, skills []Skill) error

	createService func(ctx context.Context, in Service) (Service, error)
	getService    func(ctx context.Context, id int) (Service, error)
	updateService func(ctx context.Context, in Service) (Service, error)
	deleteService func(ctx context.Context, id int) error
	listServices  func(ctx context.Context, f ServiceFilter) ([]Service, error)

	createExchange    func(ctx context.Context, serviceID, requesterID, ownerID int) (Exchange, error)
	getExchange       func(ctx context.Context, id int) (Exchange, error)
	listExchanges     func(ctx context.Context, userID int, status string) ([]Exchange, error)
	hasActiveExchange func(ctx context.Context, serviceID int) (bool, error)
	userBalance       func(ctx context.Context, userID int) (int, error)
	acceptExchange    func(ctx context.Context, id int) (Exchange, error)
	rejectExchange    func(ctx context.Context, id int) (Exchange, error)
	completeExchange  func(ctx context.Context, id int) (Exchange, error)
	cancelExchange    func(ctx context.Context, id int) (Exchange, error)

	createReview       func(ctx context.Context, in Review) (Review, error)
	listUserReviews    func(ctx context.Context, revieweeID int) ([]Review, error)
	listServiceReviews func(ctx context.Context, serviceID int) ([]Review, error)
	getUserStats       func(ctx context.Context, userID int) (UserStats, error)
}

func (f *fakeStore) CreateUser(ctx context.Context, u User) (User, error) {
	if f.createUser == nil {
		panic("appel inattendu : CreateUser")
	}
	return f.createUser(ctx, u)
}

func (f *fakeStore) GetUser(ctx context.Context, id int) (User, error) {
	if f.getUser == nil {
		panic("appel inattendu : GetUser")
	}
	return f.getUser(ctx, id)
}

func (f *fakeStore) UpdateUser(ctx context.Context, u User) (User, error) {
	if f.updateUser == nil {
		panic("appel inattendu : UpdateUser")
	}
	return f.updateUser(ctx, u)
}

func (f *fakeStore) UserExists(ctx context.Context, id int) (bool, error) {
	if f.userExists == nil {
		panic("appel inattendu : UserExists")
	}
	return f.userExists(ctx, id)
}

func (f *fakeStore) GetUserSkills(ctx context.Context, userID int) ([]Skill, error) {
	if f.getUserSkills == nil {
		panic("appel inattendu : GetUserSkills")
	}
	return f.getUserSkills(ctx, userID)
}

func (f *fakeStore) SetUserSkills(ctx context.Context, userID int, skills []Skill) error {
	if f.setUserSkills == nil {
		panic("appel inattendu : SetUserSkills")
	}
	return f.setUserSkills(ctx, userID, skills)
}

func (f *fakeStore) CreateService(ctx context.Context, in Service) (Service, error) {
	if f.createService == nil {
		panic("appel inattendu : CreateService")
	}
	return f.createService(ctx, in)
}

func (f *fakeStore) GetService(ctx context.Context, id int) (Service, error) {
	if f.getService == nil {
		panic("appel inattendu : GetService")
	}
	return f.getService(ctx, id)
}

func (f *fakeStore) UpdateService(ctx context.Context, in Service) (Service, error) {
	if f.updateService == nil {
		panic("appel inattendu : UpdateService")
	}
	return f.updateService(ctx, in)
}

func (f *fakeStore) DeleteService(ctx context.Context, id int) error {
	if f.deleteService == nil {
		panic("appel inattendu : DeleteService")
	}
	return f.deleteService(ctx, id)
}

func (f *fakeStore) ListServices(ctx context.Context, filter ServiceFilter) ([]Service, error) {
	if f.listServices == nil {
		panic("appel inattendu : ListServices")
	}
	return f.listServices(ctx, filter)
}

func (f *fakeStore) CreateExchange(ctx context.Context, serviceID, requesterID, ownerID int) (Exchange, error) {
	if f.createExchange == nil {
		panic("appel inattendu : CreateExchange")
	}
	return f.createExchange(ctx, serviceID, requesterID, ownerID)
}

func (f *fakeStore) GetExchange(ctx context.Context, id int) (Exchange, error) {
	if f.getExchange == nil {
		panic("appel inattendu : GetExchange")
	}
	return f.getExchange(ctx, id)
}

func (f *fakeStore) ListExchanges(ctx context.Context, userID int, status string) ([]Exchange, error) {
	if f.listExchanges == nil {
		panic("appel inattendu : ListExchanges")
	}
	return f.listExchanges(ctx, userID, status)
}

func (f *fakeStore) HasActiveExchange(ctx context.Context, serviceID int) (bool, error) {
	if f.hasActiveExchange == nil {
		panic("appel inattendu : HasActiveExchange")
	}
	return f.hasActiveExchange(ctx, serviceID)
}

func (f *fakeStore) UserBalance(ctx context.Context, userID int) (int, error) {
	if f.userBalance == nil {
		panic("appel inattendu : UserBalance")
	}
	return f.userBalance(ctx, userID)
}

func (f *fakeStore) AcceptExchange(ctx context.Context, id int) (Exchange, error) {
	if f.acceptExchange == nil {
		panic("appel inattendu : AcceptExchange")
	}
	return f.acceptExchange(ctx, id)
}

func (f *fakeStore) RejectExchange(ctx context.Context, id int) (Exchange, error) {
	if f.rejectExchange == nil {
		panic("appel inattendu : RejectExchange")
	}
	return f.rejectExchange(ctx, id)
}

func (f *fakeStore) CompleteExchange(ctx context.Context, id int) (Exchange, error) {
	if f.completeExchange == nil {
		panic("appel inattendu : CompleteExchange")
	}
	return f.completeExchange(ctx, id)
}

func (f *fakeStore) CancelExchange(ctx context.Context, id int) (Exchange, error) {
	if f.cancelExchange == nil {
		panic("appel inattendu : CancelExchange")
	}
	return f.cancelExchange(ctx, id)
}

func (f *fakeStore) CreateReview(ctx context.Context, in Review) (Review, error) {
	if f.createReview == nil {
		panic("appel inattendu : CreateReview")
	}
	return f.createReview(ctx, in)
}

func (f *fakeStore) ListUserReviews(ctx context.Context, revieweeID int) ([]Review, error) {
	if f.listUserReviews == nil {
		panic("appel inattendu : ListUserReviews")
	}
	return f.listUserReviews(ctx, revieweeID)
}

func (f *fakeStore) ListServiceReviews(ctx context.Context, serviceID int) ([]Review, error) {
	if f.listServiceReviews == nil {
		panic("appel inattendu : ListServiceReviews")
	}
	return f.listServiceReviews(ctx, serviceID)
}

func (f *fakeStore) GetUserStats(ctx context.Context, userID int) (UserStats, error) {
	if f.getUserStats == nil {
		panic("appel inattendu : GetUserStats")
	}
	return f.getUserStats(ctx, userID)
}
