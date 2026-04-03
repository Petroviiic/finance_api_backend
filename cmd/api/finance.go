package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Petroviiic/finance_api_backend/internal/storage"
	"github.com/go-chi/chi/v5"
)

var _ = storage.FinancialRecord{}

// ListRecords godoc
// @Summary      List financial records
// @Description  Get financial entries for current user with optional filters
// @Tags         finance
// @Security     BearerAuth
// @Param        category  query  string  false  "Filter by category"
// @Param        type      query  string  false  "Filter by type (income/expense)"
// @Param        from      query  string  false  "Start date (YYYY-MM-DD)"
// @Param        to        query  string  false  "End date (YYYY-MM-DD)"
// @Param        offset  query  string  false  "Result offset"
// @Param        limit      query  string  false  "Result limit"
// @Produce      json
// @Success      200       {array}   storage.FinancialRecord
// @Failure      401  {object}  error   "Unauthorized"
// @Failure      403  {object}  error   "Forbidden - Admin only"
// @Failure      500  {object}  error   "Internal server error"
// @Router       /finance/list_records [get]
func (app *Application) ListRecords(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r)

	category := r.URL.Query().Get("category")
	recordType := r.URL.Query().Get("type")
	startDate := r.URL.Query().Get("from")
	endDate := r.URL.Query().Get("to")

	limitStr := r.URL.Query().Get("limit")
	if limitStr == "" {
		limitStr = "10"
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	offsetStr := r.URL.Query().Get("offset")
	if offsetStr == "" {
		offsetStr = "0"
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	records, err := app.storage.FinancialStorage.GetAllRecords(r.Context(), user.Role == "admin", user.ID, category, recordType, startDate, endDate, limit, offset)
	if err != nil {
		app.internalServerErrorJson(w, r, err)
		return
	}

	if err := jsonResponse(w, http.StatusOK, records); err != nil {
		app.internalServerErrorJson(w, r, err)
	}
}

// GetFinancialTrends godoc
// @Summary      Retrieves financial trends
// @Tags         finance
// @Security     BearerAuth
// @Param        months_back      query  string  false  "Months back"
// @Produce      json
// @Success      200       {array}   storage.Trend
// @Failure      401  {object}  error   "Unauthorized"
// @Failure      403  {object}  error   "Forbidden - Admin only"
// @Failure      500  {object}  error   "Internal server error"
// @Router       /finance/trends [get]
func (app *Application) GetFinancialTrends(w http.ResponseWriter, r *http.Request) {
	monthsBackStr := r.URL.Query().Get("months_back")
	if monthsBackStr == "" {
		monthsBackStr = "3"
	}
	monthsBack, err := strconv.Atoi(monthsBackStr)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	if monthsBack < 3 {
		monthsBack = 3
	}

	trends, err := app.storage.FinancialStorage.MonthlyFinancialTrends(r.Context(), monthsBack)
	if err != nil {
		app.internalServerErrorJson(w, r, err)
		return
	}

	if err := jsonResponse(w, http.StatusOK, trends); err != nil {
		app.internalServerErrorJson(w, r, err)
	}
}

type CreateRecordRequest struct {
	Amount       float64 `json:"amount" validate:"required,gt=0"`
	TargetUserID int64   `json:"target_user_id" validate:"required"`
	Type         string  `json:"type" validate:"required,oneof=income expense"`
	Category     string  `json:"category" validate:"required,max=50"`
	EntryDate    string  `json:"entry_date" validate:"required,datetime=2006-01-02"`
	Description  string  `json:"description" validate:"omitempty,max=500"`
}

// CreateRecord godoc
// @Summary      Create a financial record
// @Description  Add a new income or expense entry. Only Admin.
// @Tags         finance
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        payload  body  CreateRecordRequest  true  "Record details"
// @Success      201      {object}  map[string]interface{}
// @Failure      400  {object}  error   "Bad payload"
// @Failure      401  {object}  error   "Unauthorized"
// @Failure      403  {object}  error   "Forbidden - Admin only"
// @Failure      500  {object}  error   "Internal server error"
// @Router       /finance/create_record [post]
func (app *Application) CreateRecord(w http.ResponseWriter, r *http.Request) {
	var input CreateRecordRequest
	if err := readJson(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	entryDate, err := time.ParseInLocation(time.DateOnly, input.EntryDate, time.Local)

	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	id, err := app.storage.FinancialStorage.CreateRecord(r.Context(), input.TargetUserID, input.Amount, input.Type, input.Category, entryDate, input.Description)
	if err != nil {
		app.internalServerErrorJson(w, r, err)
		return
	}

	if err := jsonResponse(w, http.StatusOK, fmt.Sprintf("record %d created", id)); err != nil {
		app.internalServerErrorJson(w, r, err)
	}
}

// UpdateRecord godoc
// @Summary      Update an existing financial record
// @Description  Updates a financial record by its ID. Only accessible by admins.
// @Tags         finance
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        recordID  path      int                  true  "Record ID"
// @Param        request   body      CreateRecordRequest  true  "Update Record Request"
// @Success      200    {object}  map[string]string    "record updated successfully"
// @Failure      400  	{object}  error   "Bad payload"
// @Failure      401  	{object}  error   "Unauthorized"
// @Failure      403  	{object}  error   "Forbidden - Admin only"
// @Failure      500  	{object}  error   "Internal server error"
// @Router       /finance/update_record/{recordID} [put]
func (app *Application) UpdateRecord(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "recordID")
	recordID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	var input CreateRecordRequest
	if err := readJson(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	entryDate, err := time.ParseInLocation(time.DateOnly, input.EntryDate, time.Local)

	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	err = app.storage.FinancialStorage.UpdateRecord(r.Context(), recordID, input.TargetUserID, input.Amount, input.Type, input.Category, entryDate, input.Description)
	if err != nil {
		app.internalServerErrorJson(w, r, err)
		return
	}

	if err := jsonResponse(w, http.StatusOK, "record updated successfully"); err != nil {
		app.internalServerErrorJson(w, r, err)
	}
}

// DeleteRecord godoc
// @Summary      Delete a financial record
// @Description  Deletes a financial record from the system by its ID.
// @Tags         finance
// @Produce      json
// @Security     BearerAuth
// @Param        recordID  path      int                true  "Record ID"
// @Success      200       {object}  map[string]string  "record deleted successfully"
// @Failure      400  	   {object}  error   "Bad payload"
// @Failure      401  	   {object}  error   "Unauthorized"
// @Failure      403  	   {object}  error   "Forbidden - Admin only"
// @Failure      500  	   {object}  error   "Internal server error"
// @Router       /finance/delete_record/{recordID} [delete]
func (app *Application) DeleteRecord(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "recordID")
	recordID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	err = app.storage.FinancialStorage.DeleteRecord(r.Context(), recordID)

	if err != nil {
		app.internalServerErrorJson(w, r, err)
		return
	}

	if err := jsonResponse(w, http.StatusOK, "record deleted successfully"); err != nil {
		app.internalServerErrorJson(w, r, err)
		return
	}
}

type DashboardSummary struct {
	TotalIncome   float64                    `json:"total_income"`
	TotalExpenses float64                    `json:"total_expenses"`
	NetBalance    float64                    `json:"net_balance"`
	RecentRecords []*storage.FinancialRecord `json:"recent_records"`
}

// GetDashboardSummary godoc
// @Summary      Get financial dashboard summary
// @Description  Returns aggregated totals (income, expense, balance) and the 5 most recent records
// @Tags         finance
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{} "Returns TotalIncome, TotalExpense, NetBalance and RecentRecords"
// @Failure      401  {object}  error "Unauthorized"
// @Failure      500  {object}  error "Internal server error"
// @Router       /finance/summary [get]
func (app *Application) GetDashboardSummary(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r)
	isAdmin := user.Role == "admin"

	sums, err := app.storage.FinancialStorage.GetFinancialSums(r.Context(), user.ID, isAdmin)
	if err != nil {
		app.internalServerErrorJson(w, r, err)
		return
	}

	recent, err := app.storage.FinancialStorage.GetAllRecords(r.Context(), isAdmin, user.ID, "", "", "", "", app.config.dashboard.NumberOfRecentRecords, 0)
	if err != nil {
		app.internalServerErrorJson(w, r, err)
		return
	}

	response := DashboardSummary{
		TotalIncome:   sums.TotalIncome,
		TotalExpenses: sums.TotalExpense,
		NetBalance:    sums.TotalIncome - sums.TotalExpense,
		RecentRecords: recent,
	}

	if err := jsonResponse(w, http.StatusOK, response); err != nil {
		app.internalServerErrorJson(w, r, err)
		return
	}
}
