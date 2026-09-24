package bill_module

import (
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"maphraohom.app/maphraohom-backoffice/internal/exception"
	"maphraohom.app/maphraohom-backoffice/src/models"
)

// newMockGormDB wires a *gorm.DB to a sqlmock connection so nextBillNumbers'
// exact SQL and argument handling can be tested without a real database.
func newMockGormDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open: %v", err)
	}

	return gormDB, mock
}

// nextBillNumbersQuery matches nextBillNumbers' UPDATE ... RETURNING
// statement loosely (whitespace-insensitive) so the test doesn't break on
// harmless reformatting of the SQL literal.
var nextBillNumbersQuery = `(?s)UPDATE tbl_bill_sequences.*RETURNING last_book_no, last_receipt_no`

func TestNextBillNumbers_ReturnsRowFromDB(t *testing.T) {
	gormDB, mock := newMockGormDB(t)

	rows := sqlmock.NewRows([]string{"last_book_no", "last_receipt_no"}).AddRow(1, 2)
	mock.ExpectQuery(nextBillNumbersQuery).
		WithArgs(MaxReceiptNoPerBook, MaxReceiptNoPerBook, 1, models.BillTypeReceipt).
		WillReturnRows(rows)

	bookNo, receiptNo, err := nextBillNumbers(gormDB, 1, models.BillTypeReceipt)
	if err != nil {
		t.Fatalf("nextBillNumbers returned error: %v", err)
	}
	if bookNo != 1 || receiptNo != 2 {
		t.Fatalf("got (%d, %d), want (1, 2)", bookNo, receiptNo)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestNextBillNumbers_ScopedPerStoreAndType(t *testing.T) {
	gormDB, mock := newMockGormDB(t)

	// Store 1's receipt sequence and store 1's payment sequence are
	// independent rows — the WHERE clause must carry both store_id and type.
	mock.ExpectQuery(nextBillNumbersQuery).
		WithArgs(MaxReceiptNoPerBook, MaxReceiptNoPerBook, 1, models.BillTypeReceipt).
		WillReturnRows(sqlmock.NewRows([]string{"last_book_no", "last_receipt_no"}).AddRow(3, 10))
	mock.ExpectQuery(nextBillNumbersQuery).
		WithArgs(MaxReceiptNoPerBook, MaxReceiptNoPerBook, 1, models.BillTypePayment).
		WillReturnRows(sqlmock.NewRows([]string{"last_book_no", "last_receipt_no"}).AddRow(1, 4))

	bookNo, receiptNo, err := nextBillNumbers(gormDB, 1, models.BillTypeReceipt)
	if err != nil || bookNo != 3 || receiptNo != 10 {
		t.Fatalf("receipt sequence: got (%d, %d, %v), want (3, 10, nil)", bookNo, receiptNo, err)
	}

	bookNo, receiptNo, err = nextBillNumbers(gormDB, 1, models.BillTypePayment)
	if err != nil || bookNo != 1 || receiptNo != 4 {
		t.Fatalf("payment sequence: got (%d, %d, %v), want (1, 4, nil)", bookNo, receiptNo, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestNextBillNumbers_MissingSequenceRow(t *testing.T) {
	gormDB, mock := newMockGormDB(t)

	// No (store, type) row was seeded for this store — the UPDATE matches
	// zero rows and RETURNING yields nothing.
	mock.ExpectQuery(nextBillNumbersQuery).
		WithArgs(MaxReceiptNoPerBook, MaxReceiptNoPerBook, 99, models.BillTypeReceipt).
		WillReturnRows(sqlmock.NewRows([]string{"last_book_no", "last_receipt_no"}))

	_, _, err := nextBillNumbers(gormDB, 99, models.BillTypeReceipt)
	if err != exception.ErrBillSequenceMissing {
		t.Fatalf("got err = %v, want exception.ErrBillSequenceMissing", err)
	}
}

// regexpEscapeSanity guards against a future edit accidentally turning
// nextBillNumbersQuery into something that no longer compiles as a regexp.
func TestNextBillNumbersQueryIsValidRegexp(t *testing.T) {
	if _, err := regexp.Compile(nextBillNumbersQuery); err != nil {
		t.Fatalf("nextBillNumbersQuery is not a valid regexp: %v", err)
	}
}
