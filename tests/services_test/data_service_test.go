package services_test

import (
	"context"
	"regexp"
	"testing"

	"megga-backend/internal/services"

	"github.com/pashagolub/pgxmock"
)

func TestSaveBLSData_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mock.Close()

	blsData := map[string]struct {
		Value  float64
		Year   string
		Period string
	}{
		"APU0000708111": {Value: 4.15, Year: "2024", Period: "M12"},
		"APU0000702111": {Value: 1.91, Year: "2024", Period: "M12"},
	}

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT data_id, latest_value, previous_value, year, period FROM data WHERE series_id = $1`)).
		WithArgs("APU0000708111").
		WillReturnRows(pgxmock.NewRows([]string{"data_id", "latest_value", "previous_value", "year", "period"}).
			AddRow(1, 4.1, 4.0, "2024", "M11"))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT data_id, latest_value, previous_value, year, period FROM data WHERE series_id = $1`)).
		WithArgs("APU0000702111").
		WillReturnRows(pgxmock.NewRows([]string{"data_id", "latest_value", "previous_value", "year", "period"}).
			AddRow(2, 1.8, 1.7, "2024", "M11"))

	mock.ExpectBegin()

	mock.ExpectExec(regexp.QuoteMeta(`UPDATE data SET previous_value = latest_value, latest_value = $1, year = $2, period = $3, last_updated = NOW() WHERE data_id = $4`)).
		WithArgs(4.15, "2024", "M12", 1).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	mock.ExpectExec(regexp.QuoteMeta(`UPDATE data SET previous_value = latest_value, latest_value = $1, year = $2, period = $3, last_updated = NOW() WHERE data_id = $4`)).
		WithArgs(1.91, "2024", "M12", 2).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	mock.ExpectCommit()

	mock.MatchExpectationsInOrder(false)

	err = services.SaveBLSData(mock, blsData)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("❌ Unmet mock expectations: %v", err)
	}
}

func TestSaveBLSData_PreventDuplicateEntries(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mock.Close()

	blsData := map[string]struct {
		Value  float64
		Year   string
		Period string
	}{
		"APU0000708111": {Value: 4.15, Year: "2024", Period: "M12"},
	}

	mock.ExpectBegin()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT data_id, latest_value, previous_value, year, period FROM data WHERE series_id = $1`)).
		WithArgs("APU0000708111").
		WillReturnRows(pgxmock.NewRows([]string{"data_id", "latest_value", "previous_value", "year", "period"}).
			AddRow(1, 4.15, 4.10, "2024", "M12"))

	mock.ExpectRollback()

	err = services.SaveBLSData(mock, blsData)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("❌ Unmet mock expectations: %v", err)
	}
}

func TestSaveBLSData_DBError(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mock.Close()

	blsData := map[string]struct {
		Value  float64
		Year   string
		Period string
	}{
		"APU0000708111": {Value: 4.146, Year: "2024", Period: "M12"},
	}

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT data_id, latest_value, previous_value, year, period FROM data WHERE series_id = $1`)).
		WithArgs("APU0000708111").
		WillReturnRows(pgxmock.NewRows([]string{"data_id", "latest_value", "previous_value", "year", "period"}).
			AddRow(1, 4.1, 4.0, "2024", "M11"))

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE data SET previous_value = latest_value, latest_value = $1, year = $2, period = $3, last_updated = NOW() WHERE data_id = $4`)).
		WithArgs(4.146, "2024", "M12", 1).
		WillReturnError(context.DeadlineExceeded)

	mock.ExpectRollback()

	err = services.SaveBLSData(mock, blsData)
	if err == nil {
		t.Errorf("Expected an error, got nil")
	}
}

func TestSaveBLSData_IgnoreOldData(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mock.Close()

	blsData := map[string]struct {
		Value  float64
		Year   string
		Period string
	}{
		"APU0000708111": {Value: 4.146, Year: "2024", Period: "M12"},
	}

	mock.ExpectBegin()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT data_id, latest_value, previous_value, year, period FROM data WHERE series_id = $1`)).
		WithArgs("APU0000708111").
		WillReturnRows(pgxmock.NewRows([]string{"data_id", "latest_value", "previous_value", "year", "period"}).
			AddRow(1, 4.146, 4.100, "2024", "M12"))

	mock.ExpectRollback()

	err = services.SaveBLSData(mock, blsData)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("❌ Unmet mock expectations: %v", err)
	}
}

func TestSaveBLSData_EmptyInput(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mock.Close()

	blsData := make(map[string]struct {
		Value  float64
		Year   string
		Period string
	})

	err = services.SaveBLSData(mock, blsData)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("❌ Unmet mock expectations: %v", err)
	}
}

func TestSaveBLSData_UnexpectedAPIData(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mock.Close()

	blsData := map[string]struct {
		Value  float64
		Year   string
		Period string
	}{
		"UNKNOWN_SERIES": {Value: 10.50, Year: "2024", Period: "M12"},
	}

	err = services.SaveBLSData(mock, blsData)
	if err == nil {
		t.Errorf("Expected an error for unknown series, got nil")
	}
}
