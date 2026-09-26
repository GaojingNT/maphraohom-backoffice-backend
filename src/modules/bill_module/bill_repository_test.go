package bill_module

import (
	"database/sql/driver"
	"regexp"
	"testing"
	"time"

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

// (a) A new bill records who created it, and starts never-edited with no
// slip timestamp.
func TestNewBill_RecordsCreatorAndStartsUnedited(t *testing.T) {
	userID := 7
	bill := newBill(CreateBillInput{StoreID: 1, Type: models.BillTypeReceipt, CreatedBy: &userID}, 1, 1, 3)

	if bill.CreatedBy == nil || *bill.CreatedBy != userID {
		t.Fatalf("CreatedBy = %v, want %d", bill.CreatedBy, userID)
	}
	if bill.EditedAt != nil {
		t.Errorf("EditedAt = %v, want nil on a new bill", bill.EditedAt)
	}
	if bill.SlipUploadedAt != nil || bill.Slip != nil {
		t.Errorf("new bill has slip fields set: slip=%v slipUploadedAt=%v", bill.Slip, bill.SlipUploadedAt)
	}
}

// (b) An edit stamps EditedAt and leaves the slip fields as they were.
func TestApplyBillUpdate_StampsEditedAt(t *testing.T) {
	key := "bills/slips/a.png"
	uploaded := time.Date(2026, 9, 26, 7, 32, 0, 0, time.UTC)
	bill := models.Bill{Slip: &key, SlipUploadedAt: &uploaded}
	now := time.Date(2026, 9, 26, 10, 2, 0, 0, time.UTC)

	applyBillUpdate(&bill, UpdateBillInput{CustomerName: "ป้ามาลี"}, 3, now)

	if bill.EditedAt == nil || !bill.EditedAt.Equal(now) {
		t.Fatalf("EditedAt = %v, want %v", bill.EditedAt, now)
	}
	if bill.Slip != &key || bill.SlipUploadedAt == nil || !bill.SlipUploadedAt.Equal(uploaded) {
		t.Errorf("edit changed slip fields: slip=%v slipUploadedAt=%v", bill.Slip, bill.SlipUploadedAt)
	}
}

// setBillSlipQuery pins the full SET list, so (d) is checked too: a slip
// change writes slip + slip_uploaded_at (+ GORM's updated_at) and never
// edited_at. The table prefix is optional: the mock DB doesn't get the
// app's tbl_ naming strategy.
var setBillSlipQuery = `UPDATE "(tbl_)?bills" ` + regexp.QuoteMeta(`SET "slip"=$1,"slip_uploaded_at"=$2,"updated_at"=$3 WHERE id = $4`)

// nonNilTime matches any non-nil time.Time argument.
type nonNilTime struct{}

func (nonNilTime) Match(v driver.Value) bool {
	_, ok := v.(time.Time)
	return ok
}

// (c)+(d) Attaching a slip sets slip_uploaded_at, not edited_at.
func TestSetBillSlip_AttachStampsSlipUploadedAt(t *testing.T) {
	gormDB, mock := newMockGormDB(t)
	key := "bills/slips/a.png"

	mock.ExpectBegin()
	mock.ExpectExec(setBillSlipQuery).
		WithArgs(key, nonNilTime{}, sqlmock.AnyArg(), 5).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := setBillSlip(gormDB, 5, &key, time.Now()); err != nil {
		t.Fatalf("setBillSlip returned error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// (c) Removing the slip clears slip_uploaded_at back to NULL.
func TestSetBillSlip_RemoveClearsSlipUploadedAt(t *testing.T) {
	gormDB, mock := newMockGormDB(t)

	mock.ExpectBegin()
	mock.ExpectExec(setBillSlipQuery).
		WithArgs(nil, nil, sqlmock.AnyArg(), 5).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := setBillSlip(gormDB, 5, nil, time.Now()); err != nil {
		t.Fatalf("setBillSlip returned error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}
