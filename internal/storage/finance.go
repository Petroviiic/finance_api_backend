package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type FinancialStorage struct {
	db *sql.DB
}
type FinancialRecord struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	Amount      float64   `json:"amount"`
	Type        string    `json:"type"`
	Category    string    `json:"category"`
	EntryDate   time.Time `json:"entry_date"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

func (s *FinancialStorage) GetAllRecords(ctx context.Context, isAdmin bool, userID int64, category, recordType, from, to string, limit, offset int) ([]*FinancialRecord, error) {
	query := `
        SELECT id, user_id, amount, type, category, entry_date, description, created_at 
        FROM financial_records 
        WHERE 1=1`

	args := []any{}
	argCount := 0

	if !isAdmin {
		argCount++
		query += fmt.Sprintf(" AND user_id = $%d", argCount)
		args = append(args, userID)
	}

	if category != "" {
		argCount++
		query += fmt.Sprintf(" AND category = $%d", argCount)
		args = append(args, category)
	}

	if recordType != "" {
		argCount++
		query += fmt.Sprintf(" AND type = $%d", argCount)
		args = append(args, recordType)
	}

	if from != "" {
		argCount++
		query += fmt.Sprintf(" AND entry_date >= $%d", argCount)
		args = append(args, from)
	}
	if to != "" {
		argCount++
		query += fmt.Sprintf(" AND entry_date <= $%d", argCount)
		args = append(args, to)
	}

	query += " ORDER BY entry_date DESC"
	argCount++
	query += fmt.Sprintf(" LIMIT $%d", argCount)
	args = append(args, limit)

	argCount++
	query += fmt.Sprintf(" OFFSET $%d", argCount)
	args = append(args, offset)

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rows, err := s.db.QueryContext(
		ctx,
		query,
		args...,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*FinancialRecord
	for rows.Next() {
		r := &FinancialRecord{}
		err := rows.Scan(
			&r.ID,
			&r.UserID,
			&r.Amount,
			&r.Type,
			&r.Category,
			&r.EntryDate,
			&r.Description,
			&r.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, nil
}

func (s *FinancialStorage) CreateRecord(ctx context.Context, userID int64, Amount float64, Type string, Category string, EntryDate time.Time, Description string) (int64, error) {
	query := `
			INSERT INTO financial_records
			(user_id, amount, type, category, entry_date, description) VALUES 
			($1, $2, $3, $4, $5, $6)
			RETURNING id; 
		`
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var id int64
	err := s.db.QueryRowContext(
		ctx,
		query,
		userID,
		Amount,
		Type,
		Category,
		EntryDate,
		Description,
	).Scan(
		&id,
	)

	if err != nil {
		return 0, err
	}
	return id, nil

}
func (s *FinancialStorage) DeleteRecord(ctx context.Context, recordID int64) error {
	query := `DELETE FROM financial_records WHERE id = $1`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := s.db.ExecContext(
		ctx,
		query,
		recordID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ERROR_NO_ROWS_AFFECTED
	}

	return nil
}

func (s *FinancialStorage) UpdateRecord(ctx context.Context, recordID int64, userID int64, Amount float64, Type string, Category string, EntryDate time.Time, Description string) error {
	query := `
			UPDATE financial_records SET
			user_id = $1, 
			amount = $2, 
			type = $3, 
			category = $4, 
			entry_date = $5, 
			description = $6
			WHERE id = $7;
		`
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := s.db.ExecContext(
		ctx,
		query,
		userID,
		Amount,
		Type,
		Category,
		EntryDate,
		Description,
		recordID,
	)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ERROR_NO_ROWS_AFFECTED
	}
	return nil
}

type Trend struct {
	Date         time.Time `json:"date"`
	TotalIncome  float64   `json:"total_income"`
	TotalExpense float64   `json:"total_expense"`
}

func (s *FinancialStorage) MonthlyFinancialTrends(ctx context.Context, monthsBack int) ([]*Trend, error) {
	query := `
		SELECT 
			date_trunc('month', entry_date) AS month,
			SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END) as total_income,
			SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END) as total_expense
		FROM financial_records
		WHERE entry_date >= CURRENT_DATE - ($1 || ' month')::INTERVAL
		GROUP BY month
		ORDER BY month DESC
	`
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rows, err := s.db.QueryContext(
		ctx,
		query,
		monthsBack,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trends []*Trend
	for rows.Next() {
		r := &Trend{}
		err := rows.Scan(
			&r.Date,
			&r.TotalIncome,
			&r.TotalExpense,
		)
		if err != nil {
			return nil, err
		}
		trends = append(trends, r)
	}
	return trends, nil
}

type FinancalSum struct {
	TotalIncome  float64 `json:"total_income"`
	TotalExpense float64 `json:"total_expense"`
}

func (s *FinancialStorage) GetFinancialSums(ctx context.Context, userID int64, isAdmin bool) (*FinancalSum, error) {
	query := `SELECT 
				COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END), 0) as total_income,
				COALESCE(SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END), 0) as total_expenses
			FROM financial_records
			WHERE (user_id = $1 OR $2 = TRUE);
			`
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	sum := &FinancalSum{}

	err := s.db.QueryRowContext(ctx, query, userID, isAdmin).Scan(
		&sum.TotalIncome,
		&sum.TotalExpense,
	)

	if err != nil {
		return nil, err
	}

	return sum, nil
}
