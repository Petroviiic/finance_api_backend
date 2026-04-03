package storage

import (
	"context"
	"fmt"
	"time"
)

func NewMockStorage() *Storage {
	return &Storage{
		UserStorage:      &MockUserStorage{},
		FinancialStorage: &MockFinancialStorage{},
	}
}

type MockUserStorage struct {
}
type MockFinancialStorage struct {
}

func (u *MockUserStorage) GetByUsername(ctx context.Context, username string) (*User, error) {
	user := &User{
		ID:       1,
		Username: "admin_user",
		Role:     "admin",
		IsActive: true,
	}
	_ = user.Password.Set("password123")
	switch username {
	case "admin_user":
		return user, nil
	default:
		return nil, fmt.Errorf("not found") // Simuliraj da korisnik ne postoji
	}
}

func (u *MockUserStorage) GetById(ctx context.Context, userId int64) (*User, error) {
	switch userId {
	case 1:
		return &User{
			ID:       1,
			Username: "admin_user",
			Role:     "admin",
			IsActive: true,
		}, nil
	case 2:
		return &User{
			ID:       2,
			Username: "analyst_user",
			Role:     "analyst",
			IsActive: true,
		}, nil
	case 3:
		return &User{
			ID:       3,
			Username: "viewer_user",
			Role:     "viewer",
			IsActive: true,
		}, nil
	default:
		return nil, fmt.Errorf("not found") // Simuliraj da korisnik ne postoji
	}
}

func (u *MockUserStorage) RegisterUser(ctx context.Context, user *User) (int64, error) {
	return -1, nil
}

func (u *MockUserStorage) UpdateUserStatus(ctx context.Context, id int64, isActive bool) error {
	return nil
}

func (u *MockUserStorage) UpdateUserRole(ctx context.Context, id int64, role string) error {
	return nil
}

func (u *MockUserStorage) GetAllUsers(ctx context.Context) ([]*User, error) {
	return make([]*User, 0), nil
}

func (u *MockUserStorage) DeleteUser(ctx context.Context, id int64) error {
	return nil
}

func (s *MockFinancialStorage) GetAllRecords(ctx context.Context, isAdmin bool, userID int64, category, recordType, from, to string, limit, offset int) ([]*FinancialRecord, error) {
	return make([]*FinancialRecord, 0), nil
}

func (s *MockFinancialStorage) CreateRecord(ctx context.Context, userID int64, Amount float64, Type string, Category string, EntryDate time.Time, Description string) (int64, error) {
	return 0, nil

}
func (s *MockFinancialStorage) DeleteRecord(context.Context, int64) error {
	return nil
}

func (s *MockFinancialStorage) UpdateRecord(context.Context, int64, int64, float64, string, string, time.Time, string) error {
	return nil
}

func (s *MockFinancialStorage) MonthlyFinancialTrends(context.Context, int) ([]*Trend, error) {
	return make([]*Trend, 0), nil
}

func (s *MockFinancialStorage) GetFinancialSums(context.Context, int64, bool) (*FinancalSum, error) {
	return &FinancalSum{}, nil
}
