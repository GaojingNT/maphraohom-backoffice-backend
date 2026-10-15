package bill_module

import (
	"context"
	"database/sql"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/shopspring/decimal"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"maphraohom.app/maphraohom-backoffice/internal/exception"
	"maphraohom.app/maphraohom-backoffice/internal/utils"
	"maphraohom.app/maphraohom-backoffice/pkg/database/paginator"
	"maphraohom.app/maphraohom-backoffice/src/models"
)

// CreateBillItemInput is one already-priced, already-validated line item —
// Unit and Subtotal are resolved by the service before the repository ever
// sees them.
type CreateBillItemInput struct {
	ProductID int
	Unit      string
	Quantity  decimal.Decimal
	Price     decimal.Decimal
	Subtotal  decimal.Decimal
}

// CreateBillInput carries the already-validated, already-priced fields the
// service resolved from the request. Total is computed by the service (not
// re-derived here) so create and update share one formula. Book/receipt
// numbers and customer linkage are resolved inside the repository
// transaction so they stay consistent under concurrent writes.
type CreateBillInput struct {
	Type            string
	StoreID         int
	CustomerID      *int
	CustomerName    string
	CustomerAddress string
	CustomerPhone   string
	Discount        decimal.Decimal
	ShippingFee     decimal.Decimal
	Total           decimal.Decimal
	Items           []CreateBillItemInput
	// CreatedAt back-/post-dates the bill when set — nil means "now" (the
	// database column's own default). Never affects book/receipt numbering,
	// which is scoped by the real server clock's calendar year regardless.
	CreatedAt *time.Time
	// CreatedBy is the id of the logged-in user issuing the bill.
	CreatedBy *int
}

// UpdateBillInput carries the already-validated fields for replacing a
// bill's editable fields and its full set of line items (existing items are
// deleted and recreated, not diffed). Book/receipt numbers and the slip are
// never touched.
type UpdateBillInput struct {
	Type            string
	StoreID         int
	CustomerID      *int
	CustomerName    string
	CustomerAddress string
	CustomerPhone   string
	Discount        decimal.Decimal
	ShippingFee     decimal.Decimal
	Total           decimal.Decimal
	Items           []CreateBillItemInput
}

// billListFilters narrows GetBillPaginate beyond the generic search/period
// scopes — every field is optional (zero value = no filter on it).
type billListFilters struct {
	Type    string
	StoreID int
	From    time.Time
	To      time.Time
}

func (r Repository) GetBillPaginate(ctx context.Context, pagination *paginator.Pagination) (*paginator.Pagination, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetBillPaginateRepository", trace.WithAttributes(attribute.String("repository", "GetBillPaginate")))
		bills        = make([]models.Bill, 0)
		err          error
	)

	// Get attributes
	searchAttribute, _ := pagination.GetStringAttribute("search")
	searchByAttribute, _ := pagination.GetStringAttribute("search_by")
	typeAttribute, _ := pagination.GetStringAttribute("type")
	storeIDAttribute, _ := pagination.GetIntAttribute("store_id")
	fromAttribute, _ := pagination.GetStringAttribute("from")
	toAttribute, _ := pagination.GetStringAttribute("to")

	// Set tracing attributes
	r.tracer.SetAttributes(childSpan, attribute.String("search", searchAttribute))
	r.tracer.SetAttributes(childSpan, attribute.String("search_by", searchByAttribute))
	r.tracer.SetAttributes(childSpan, attribute.String("type", typeAttribute))

	tx := r.db
	if typeAttribute != "" {
		tx = tx.Where("type = ?", typeAttribute)
	}
	if storeIDAttribute != 0 {
		tx = tx.Where("store_id = ?", storeIDAttribute)
	}
	if from, err := time.ParseInLocation("2006-01-02", fromAttribute, time.Local); err == nil {
		tx = tx.Where("created_at >= ?", from)
	}
	if to, err := time.ParseInLocation("2006-01-02", toAttribute, time.Local); err == nil {
		tx = tx.Where("created_at < ?", to.AddDate(0, 0, 1))
	}

	utils.Block{
		Try: func() {
			// Execute query
			if err = tx.
				Preload("Items").
				Preload("Store").
				Scopes(models.SearchingScope(models.BillSearchable(), searchAttribute, searchByAttribute)).
				Scopes(paginator.Paginate(bills, pagination, tx)).
				Find(&bills).Error; err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			// Logging
			r.logger.Error(err.Error())
			sentry.CaptureException(err)
			exception.SqlErrorMessage = err.Error()
			err = exception.ErrDbQueryStatement
		},
		Finally: nil,
	}.Do()

	// Set data
	pagination.Data = bills

	r.tracer.TraceEnd(childSpan)

	// Check error
	if err != nil {
		return nil, err
	}

	return pagination, nil
}

func (r Repository) GetBillByID(ctx context.Context, id int) (models.Bill, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetBillByIDRepository", trace.WithAttributes(attribute.String("repository", "GetBillByID"), attribute.Int64("id", int64(id))))
		bill         models.Bill
		err          error
	)

	utils.Block{
		Try: func() {
			// Execute query
			if err = preloadBillDetail(r.db).First(&bill, id).Error; err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			if err == gorm.ErrRecordNotFound {
				err = exception.ErrRecordNotFound
			} else {
				err = e.(error)
				// Logging
				r.logger.Error(err.Error())
				sentry.CaptureException(err)
				exception.SqlErrorMessage = err.Error()
				err = exception.ErrDbQueryStatement
			}
		},
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)

	// Check error
	if err != nil {
		return bill, err
	}

	return bill, nil
}

// GetProductsByID returns every requested product that exists and hasn't
// been soft-deleted, keyed by id — a missing key in the result means the
// caller should treat that id as invalid.
func (r Repository) GetProductsByID(ctx context.Context, ids []int) (map[int]models.Product, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetProductsByIDRepository", trace.WithAttributes(attribute.String("repository", "GetProductsByID")))
		products     = make([]models.Product, 0, len(ids))
		err          error
	)

	if len(ids) == 0 {
		r.tracer.TraceEnd(childSpan)
		return map[int]models.Product{}, nil
	}

	utils.Block{
		Try: func() {
			if err = r.db.Where("id IN ?", ids).Find(&products).Error; err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			r.logger.Error(err.Error())
			sentry.CaptureException(err)
			exception.SqlErrorMessage = err.Error()
			err = exception.ErrDbQueryStatement
		},
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)

	if err != nil {
		return nil, err
	}

	byID := make(map[int]models.Product, len(products))
	for _, p := range products {
		byID[p.ID] = p
	}
	return byID, nil
}

// CustomerExists reports whether a (non-deleted) customer with this id
// exists.
func (r Repository) CustomerExists(ctx context.Context, id int) (bool, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "CustomerExistsRepository", trace.WithAttributes(attribute.String("repository", "CustomerExists"), attribute.Int("id", id)))
		count        int64
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Model(&models.Customer{}).Where("id = ?", id).Count(&count).Error; err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			r.logger.Error(err.Error())
			sentry.CaptureException(err)
			exception.SqlErrorMessage = err.Error()
			err = exception.ErrDbQueryStatement
		},
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// nextBillNumbers atomically advances tbl_bill_sequences for (storeID, type)
// and returns the book/receipt numbers the new bill should use. Must run
// inside the same transaction as the bill insert. The CASE expressions here
// must stay in sync with the pure nextBookReceipt function in calc.go (kept
// there purely so the rule has a unit test — this SQL is what actually runs,
// since only a single atomically-locked UPDATE is safe under concurrent
// requests: the row lock Postgres takes for the UPDATE serializes concurrent
// callers for the same (store, type), so no separate advisory lock is
// needed).
func nextBillNumbers(tx *gorm.DB, storeID int, billType string) (bookNo int, receiptNo int, err error) {
	row := tx.Raw(`
		UPDATE tbl_bill_sequences
		SET
			last_book_no = CASE
				WHEN last_book_no = 0 OR EXTRACT(YEAR FROM updated_at) <> EXTRACT(YEAR FROM now()) THEN 1
				WHEN last_receipt_no >= ? THEN last_book_no + 1
				ELSE last_book_no
			END,
			last_receipt_no = CASE
				WHEN last_book_no = 0 OR EXTRACT(YEAR FROM updated_at) <> EXTRACT(YEAR FROM now()) THEN 1
				WHEN last_receipt_no >= ? THEN 1
				ELSE last_receipt_no + 1
			END,
			updated_at = now()
		WHERE store_id = ? AND type = ?
		RETURNING last_book_no, last_receipt_no
	`, MaxReceiptNoPerBook, MaxReceiptNoPerBook, storeID, billType).Row()

	if scanErr := row.Scan(&bookNo, &receiptNo); scanErr != nil {
		if scanErr == sql.ErrNoRows {
			return 0, 0, exception.ErrBillSequenceMissing
		}
		return 0, 0, scanErr
	}

	return bookNo, receiptNo, nil
}

// preloadBillDetail loads everything responses.BillDetailResponse.Make
// reads: the store, each item's product, the slips (in attach order), and
// the creator's name — only id/first_name/last_name, so email, password
// hash and signature never leave the users table.
func preloadBillDetail(db *gorm.DB) *gorm.DB {
	return db.
		Preload("Store").
		Preload("Items.Product").
		Preload("Slips", func(db *gorm.DB) *gorm.DB {
			return db.Order("id")
		}).
		Preload("Creator", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "first_name", "last_name")
		})
}

// newBill builds the row CreateBill inserts. EditedAt and SlipUploadedAt
// start nil: a new bill has never been edited and has no slip yet.
func newBill(input CreateBillInput, bookNo int, receiptNo int, customerID int) models.Bill {
	bill := models.Bill{
		StoreID:         input.StoreID,
		Type:            input.Type,
		CustomerID:      &customerID,
		BookNo:          bookNo,
		ReceiptNo:       receiptNo,
		CustomerName:    input.CustomerName,
		CustomerAddress: input.CustomerAddress,
		CustomerPhone:   input.CustomerPhone,
		Discount:        input.Discount,
		ShippingFee:     input.ShippingFee,
		Total:           input.Total,
		Slip:            nil,
		CreatedBy:       input.CreatedBy,
	}
	// A caller-supplied CreatedAt back-/post-dates the bill — GORM only
	// auto-fills CreatedAt when it's still the zero value, so setting it
	// here before Create is enough to override the column's
	// CURRENT_TIMESTAMP default.
	if input.CreatedAt != nil {
		bill.CreatedAt = *input.CreatedAt
	}
	return bill
}

// applyBillUpdate copies an edit onto the loaded bill and stamps EditedAt.
// Slip and SlipUploadedAt are left as loaded — an edit is not a slip change.
func applyBillUpdate(bill *models.Bill, input UpdateBillInput, customerID int, now time.Time) {
	bill.CustomerID = &customerID
	bill.CustomerName = input.CustomerName
	bill.CustomerAddress = input.CustomerAddress
	bill.CustomerPhone = input.CustomerPhone
	bill.Discount = input.Discount
	bill.ShippingFee = input.ShippingFee
	bill.Total = input.Total
	bill.EditedAt = &now
}

// billSlipMirror is what tbl_bills.slip/slip_uploaded_at hold for a bill
// with these slips (in id order): the first slip's key and the most recent
// attach time, or nil/nil when there are none.
func billSlipMirror(slips []models.BillSlip) (*string, *time.Time) {
	if len(slips) == 0 {
		return nil, nil
	}
	key := slips[0].ObjectKey
	latest := slips[0].CreatedAt
	for _, slip := range slips[1:] {
		if slip.CreatedAt.After(latest) {
			latest = slip.CreatedAt
		}
	}
	return &key, &latest
}

// syncBillSlipMirror re-derives tbl_bills.slip/slip_uploaded_at from the
// bill's current tbl_bill_slips rows — run in the same transaction as every
// slip change. edited_at is not touched: a slip change is not an edit of
// the bill.
func syncBillSlipMirror(tx *gorm.DB, billID int) error {
	var slips []models.BillSlip
	if err := tx.Where("bill_id = ?", billID).Order("id").Find(&slips).Error; err != nil {
		return err
	}
	key, uploadedAt := billSlipMirror(slips)
	return tx.Model(&models.Bill{}).Where("id = ?", billID).Updates(map[string]any{
		"slip":             key,
		"slip_uploaded_at": uploadedAt,
	}).Error
}

// lockBill takes a row lock on a (non-deleted) bill for the rest of the
// transaction, so concurrent slip uploads to the same bill are serialized
// and can't push it past models.MaxSlipsPerBill together.
func lockBill(tx *gorm.DB, billID int) error {
	var bill models.Bill
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Select("id").First(&bill, billID).Error
	if err == gorm.ErrRecordNotFound {
		return exception.ErrRecordNotFound
	}
	return err
}

func (r Repository) CreateBill(ctx context.Context, input CreateBillInput) (models.Bill, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "CreateBillRepository", trace.WithAttributes(attribute.String("repository", "CreateBill")))
		bill         models.Bill
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Transaction(func(tx *gorm.DB) error {
				bookNo, receiptNo, txErr := nextBillNumbers(tx, input.StoreID, input.Type)
				if txErr != nil {
					return txErr
				}

				customerID, txErr := r.resolveCustomerID(tx, input.CustomerID, input.CustomerName, input.CustomerAddress, input.CustomerPhone)
				if txErr != nil {
					return txErr
				}

				billItems := make([]models.BillItem, 0, len(input.Items))
				for _, item := range input.Items {
					billItems = append(billItems, models.BillItem{
						ProductID: item.ProductID,
						Unit:      item.Unit,
						Quantity:  item.Quantity,
						Price:     item.Price,
						Subtotal:  item.Subtotal,
					})
				}

				bill = newBill(input, bookNo, receiptNo, customerID)

				if txErr := tx.Create(&bill).Error; txErr != nil {
					return txErr
				}

				for i := range billItems {
					billItems[i].BillID = bill.ID
				}

				if txErr := tx.Create(&billItems).Error; txErr != nil {
					return txErr
				}

				bill.Items = billItems

				return nil
			}); err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			switch e {
			case exception.ErrBillSequenceMissing:
				err = exception.ErrBillSequenceMissing
			default:
				err = e.(error)
				// Logging
				r.logger.Error(err.Error())
				sentry.CaptureException(err)
				exception.SqlErrorMessage = err.Error()
				err = exception.ErrDbQueryStatement
			}
		},
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)

	// Check error
	if err != nil {
		return bill, err
	}

	// Reload with Store + Items.Product preloaded to match the shape
	// responses.BillDetailResponse.Make expects (storeName, each item's
	// productName).
	if err = preloadBillDetail(r.db).First(&bill, bill.ID).Error; err != nil {
		return bill, err
	}

	return bill, nil
}

// UpdateBill replaces a bill's editable fields and line items in one
// transaction. Book/receipt numbers and the slip are never touched. Line
// items are deleted and recreated wholesale rather than diffed against the
// request. Returns exception.ErrBillFieldImmutable if input.StoreID or
// input.Type differ from the bill's current values.
func (r Repository) UpdateBill(ctx context.Context, id int, input UpdateBillInput) (models.Bill, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "UpdateBillRepository", trace.WithAttributes(attribute.String("repository", "UpdateBill"), attribute.Int64("id", int64(id))))
		bill         models.Bill
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Transaction(func(tx *gorm.DB) error {
				if txErr := tx.First(&bill, id).Error; txErr != nil {
					if txErr == gorm.ErrRecordNotFound {
						return exception.ErrRecordNotFound
					}
					return txErr
				}

				if bill.StoreID != input.StoreID || bill.Type != input.Type {
					return exception.ErrBillFieldImmutable
				}

				customerID, txErr := r.resolveCustomerID(tx, input.CustomerID, input.CustomerName, input.CustomerAddress, input.CustomerPhone)
				if txErr != nil {
					return txErr
				}

				billItems := make([]models.BillItem, 0, len(input.Items))
				for _, item := range input.Items {
					billItems = append(billItems, models.BillItem{
						BillID:    bill.ID,
						ProductID: item.ProductID,
						Unit:      item.Unit,
						Quantity:  item.Quantity,
						Price:     item.Price,
						Subtotal:  item.Subtotal,
					})
				}

				// Replace line items wholesale rather than diffing.
				if txErr := tx.Where("bill_id = ?", bill.ID).Delete(&models.BillItem{}).Error; txErr != nil {
					return txErr
				}
				if txErr := tx.Create(&billItems).Error; txErr != nil {
					return txErr
				}

				applyBillUpdate(&bill, input, customerID, time.Now())

				// The slip columns are owned by the slip endpoints — omit
				// them so a concurrent slip change isn't overwritten with
				// the values loaded above.
				if txErr := tx.Omit("slip", "slip_uploaded_at", "Slips").Save(&bill).Error; txErr != nil {
					return txErr
				}

				bill.Items = billItems

				return nil
			}); err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			switch e {
			case exception.ErrRecordNotFound:
				err = exception.ErrRecordNotFound
			case exception.ErrBillFieldImmutable:
				err = exception.ErrBillFieldImmutable
			default:
				err = e.(error)
				// Logging
				r.logger.Error(err.Error())
				sentry.CaptureException(err)
				exception.SqlErrorMessage = err.Error()
				err = exception.ErrDbQueryStatement
			}
		},
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)

	if err != nil {
		return bill, err
	}

	// Reload with Store + Items.Product preloaded to match the shape
	// responses.BillDetailResponse.Make expects.
	if err = preloadBillDetail(r.db).First(&bill, bill.ID).Error; err != nil {
		return bill, err
	}

	return bill, nil
}

// resolveCustomerID resolves the counterparty (customer on a receipt bill,
// payee on a payment bill) to a customer id. When customerID is given, it is
// verified to exist and its address/phone are synced same as the by-name
// path below. Otherwise the customer is found (matched by exact name) or
// created, and its address/phone are kept in sync with the latest ones used
// — the address and phone are stored as independent snapshots on the bill
// itself either way.
func (r Repository) resolveCustomerID(tx *gorm.DB, customerID *int, name string, address string, phone string) (int, error) {
	if customerID != nil {
		var customer models.Customer
		if err := tx.First(&customer, *customerID).Error; err != nil {
			return 0, err
		}
		if err := r.syncCustomerAddressPhone(tx, customer.ID, address, phone); err != nil {
			return 0, err
		}
		return customer.ID, nil
	}

	return r.resolveCustomer(tx, name, address, phone)
}

// resolveCustomer finds a customer by exact name, creating the customer and/or
// new address/phone records when needed. The address and phone are stored as
// snapshots on the bill itself; this function keeps the customer_addresses and
// customer_phones tables in sync so future bills can prefill from them.
func (r Repository) resolveCustomer(tx *gorm.DB, name string, address string, phone string) (int, error) {
	var customer models.Customer
	err := tx.Where("name = ?", name).First(&customer).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return 0, err
	}

	// New customer: create with first (default) address and phone.
	if err == gorm.ErrRecordNotFound {
		customer = models.Customer{Name: name}
		if err = tx.Create(&customer).Error; err != nil {
			return 0, err
		}

		if err = tx.Create(&models.CustomerAddress{
			CustomerID: customer.ID,
			Address:    address,
			IsDefault:  true,
		}).Error; err != nil {
			return 0, err
		}

		if phone != "" {
			if err = tx.Create(&models.CustomerPhone{
				CustomerID: customer.ID,
				Phone:      phone,
				IsDefault:  true,
			}).Error; err != nil {
				return 0, err
			}
		}

		return customer.ID, nil
	}

	if err = r.syncCustomerAddressPhone(tx, customer.ID, address, phone); err != nil {
		return 0, err
	}

	return customer.ID, nil
}

// syncCustomerAddressPhone adds address/phone as a new default entry for an
// existing customer when neither is already on file verbatim — it never
// removes or edits an existing entry.
func (r Repository) syncCustomerAddressPhone(tx *gorm.DB, customerID int, address string, phone string) error {
	var existingAddress models.CustomerAddress
	err := tx.Where("customer_id = ? AND address = ?", customerID, address).First(&existingAddress).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	if err == gorm.ErrRecordNotFound {
		if err = tx.Model(&models.CustomerAddress{}).
			Where("customer_id = ?", customerID).
			Update("is_default", false).Error; err != nil {
			return err
		}
		if err = tx.Create(&models.CustomerAddress{
			CustomerID: customerID,
			Address:    address,
			IsDefault:  true,
		}).Error; err != nil {
			return err
		}
	}

	if phone != "" {
		var existingPhone models.CustomerPhone
		err = tx.Where("customer_id = ? AND phone = ?", customerID, phone).First(&existingPhone).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}
		if err == gorm.ErrRecordNotFound {
			if err = tx.Model(&models.CustomerPhone{}).
				Where("customer_id = ?", customerID).
				Update("is_default", false).Error; err != nil {
				return err
			}
			if err = tx.Create(&models.CustomerPhone{
				CustomerID: customerID,
				Phone:      phone,
				IsDefault:  true,
			}).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

func (r Repository) DeleteBill(ctx context.Context, id int) error {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "DeleteBillRepository", trace.WithAttributes(attribute.String("repository", "DeleteBill"), attribute.Int64("id", int64(id))))
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Delete(&models.Bill{}, id).Error; err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			if err == gorm.ErrRecordNotFound {
				err = exception.ErrRecordNotFound
			} else {
				err = e.(error)
				// Logging
				r.logger.Error(err.Error())
				sentry.CaptureException(err)
				exception.SqlErrorMessage = err.Error()
				err = exception.ErrDbQueryStatement
			}
		},
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)

	return err
}

// CountBillSlips returns how many slips a bill has, after checking the bill
// exists and hasn't been soft-deleted.
func (r Repository) CountBillSlips(ctx context.Context, billID int) (int, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "CountBillSlipsRepository", trace.WithAttributes(attribute.String("repository", "CountBillSlips"), attribute.Int64("id", int64(billID))))
		bill         models.Bill
		count        int64
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Select("id").First(&bill, billID).Error; err != nil {
				utils.Throw(err)
			}
			if err = r.db.Model(&models.BillSlip{}).Where("bill_id = ?", billID).Count(&count).Error; err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			if err == gorm.ErrRecordNotFound {
				err = exception.ErrRecordNotFound
			} else {
				err = e.(error)
				r.logger.Error(err.Error())
				sentry.CaptureException(err)
				exception.SqlErrorMessage = err.Error()
				err = exception.ErrDbQueryStatement
			}
		},
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)

	if err != nil {
		return 0, err
	}

	return int(count), nil
}

// AddBillSlip attaches an already-uploaded object to a bill and re-syncs
// the bill's slip mirror. Returns exception.ErrTooManySlips when the bill
// already has models.MaxSlipsPerBill slips — checked under the bill's row
// lock, so concurrent uploads can't both slip past it.
func (r Repository) AddBillSlip(ctx context.Context, billID int, key string, uploadedBy *int) (models.BillSlip, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "AddBillSlipRepository", trace.WithAttributes(attribute.String("repository", "AddBillSlip"), attribute.Int64("id", int64(billID))))
		slip         models.BillSlip
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Transaction(func(tx *gorm.DB) error {
				if txErr := lockBill(tx, billID); txErr != nil {
					return txErr
				}

				var count int64
				if txErr := tx.Model(&models.BillSlip{}).Where("bill_id = ?", billID).Count(&count).Error; txErr != nil {
					return txErr
				}
				if count >= models.MaxSlipsPerBill {
					return exception.ErrTooManySlips
				}

				slip = models.BillSlip{BillID: billID, ObjectKey: key, UploadedBy: uploadedBy}
				if txErr := tx.Create(&slip).Error; txErr != nil {
					return txErr
				}

				return syncBillSlipMirror(tx, billID)
			}); err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			switch e {
			case exception.ErrRecordNotFound:
				err = exception.ErrRecordNotFound
			case exception.ErrTooManySlips:
				err = exception.ErrTooManySlips
			default:
				err = e.(error)
				r.logger.Error(err.Error())
				sentry.CaptureException(err)
				exception.SqlErrorMessage = err.Error()
				err = exception.ErrDbQueryStatement
			}
		},
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)

	return slip, err
}

// DeleteBillSlip removes one of a bill's slips and re-syncs the bill's slip
// mirror, returning the removed slip's object key so the caller can delete
// the file. A slip id that doesn't belong to this bill is ErrRecordNotFound.
func (r Repository) DeleteBillSlip(ctx context.Context, billID int, slipID int) (string, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "DeleteBillSlipRepository", trace.WithAttributes(attribute.String("repository", "DeleteBillSlip"), attribute.Int64("id", int64(billID)), attribute.Int64("slipId", int64(slipID))))
		slip         models.BillSlip
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Transaction(func(tx *gorm.DB) error {
				if txErr := lockBill(tx, billID); txErr != nil {
					return txErr
				}

				if txErr := tx.Where("id = ? AND bill_id = ?", slipID, billID).First(&slip).Error; txErr != nil {
					if txErr == gorm.ErrRecordNotFound {
						return exception.ErrRecordNotFound
					}
					return txErr
				}
				if txErr := tx.Delete(&slip).Error; txErr != nil {
					return txErr
				}

				return syncBillSlipMirror(tx, billID)
			}); err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			switch e {
			case exception.ErrRecordNotFound:
				err = exception.ErrRecordNotFound
			default:
				err = e.(error)
				r.logger.Error(err.Error())
				sentry.CaptureException(err)
				exception.SqlErrorMessage = err.Error()
				err = exception.ErrDbQueryStatement
			}
		},
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)

	if err != nil {
		return "", err
	}

	return slip.ObjectKey, nil
}
