package bill_module

import (
	"context"
	"time"

	"github.com/getsentry/sentry-go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
	"maphraohom.app/maphraohom-backoffice/internal/exception"
	"maphraohom.app/maphraohom-backoffice/internal/utils"
	"maphraohom.app/maphraohom-backoffice/pkg/database/paginator"
	"maphraohom.app/maphraohom-backoffice/src/models"
)

// MaxReceiptNoPerBook is how many receipts ("ใบเสร็จ") fit in one physical
// book ("เล่ม") before rolling over to the next book. Numbering resets to
// book 1 / receipt 1 at the start of every calendar year.
const MaxReceiptNoPerBook = 50

// CreateBillItemInput is one line item of a create-bill request.
type CreateBillItemInput struct {
	ProductID int
	Kilogram  float64
}

// CreateBillInput carries the already-validated fields the service resolved
// from the request. Everything derived (book/receipt numbers, price
// snapshots, subtotals, total, customer linkage) is computed inside the
// repository transaction so it stays consistent under concurrent writes.
type CreateBillInput struct {
	StoreID         int
	CustomerName    string
	CustomerAddress string
	Discount        float64
	ShippingFee     float64
	Slip            string
	Items           []CreateBillItemInput
}

// UpdateBillInput carries the already-validated fields for replacing a
// bill's editable fields and its full set of line items (existing items are
// deleted and recreated, not diffed). Book/receipt numbers are untouched.
// Slip is nil to leave the existing slip as-is, or a pointer to the new
// value (possibly "") to replace/clear it.
type UpdateBillInput struct {
	StoreID         int
	CustomerName    string
	CustomerAddress string
	Discount        float64
	ShippingFee     float64
	Slip            *string
	Items           []CreateBillItemInput
}

// billPeriodRange resolves a "day"/"week"/"month"/"year" period (relative to
// refDate, or today if refDate is empty/unparseable) into a [from, to) range
// for filtering bills.created_at. Week starts on Monday. ok is false when
// period is empty or unrecognized, meaning no date filter should be applied.
func billPeriodRange(period string, refDate string) (from time.Time, to time.Time, ok bool) {
	if period == "" {
		return time.Time{}, time.Time{}, false
	}

	ref := time.Now()
	if refDate != "" {
		if parsed, err := time.ParseInLocation("2006-01-02", refDate, time.Local); err == nil {
			ref = parsed
		}
	}
	startOfDay := time.Date(ref.Year(), ref.Month(), ref.Day(), 0, 0, 0, 0, time.Local)

	switch period {
	case "day":
		return startOfDay, startOfDay.AddDate(0, 0, 1), true
	case "week":
		// ISO week starting Monday.
		weekday := int(startOfDay.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		from = startOfDay.AddDate(0, 0, -(weekday - 1))
		return from, from.AddDate(0, 0, 7), true
	case "month":
		from = time.Date(ref.Year(), ref.Month(), 1, 0, 0, 0, 0, time.Local)
		return from, from.AddDate(0, 1, 0), true
	case "year":
		from = time.Date(ref.Year(), time.January, 1, 0, 0, 0, 0, time.Local)
		return from, from.AddDate(1, 0, 0), true
	default:
		return time.Time{}, time.Time{}, false
	}
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
	periodAttribute, _ := pagination.GetStringAttribute("period")
	dateAttribute, _ := pagination.GetStringAttribute("date")

	// Set tracing attributes
	r.tracer.SetAttributes(childSpan, attribute.String("search", searchAttribute))
	r.tracer.SetAttributes(childSpan, attribute.String("search_by", searchByAttribute))
	r.tracer.SetAttributes(childSpan, attribute.String("period", periodAttribute))

	tx := r.db
	if from, to, ok := billPeriodRange(periodAttribute, dateAttribute); ok {
		tx = tx.Where("created_at >= ? AND created_at < ?", from, to)
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
			if err = r.db.Preload("Store").Preload("Items.Product").First(&bill, id).Error; err != nil {
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

func (r Repository) CreateBill(ctx context.Context, input CreateBillInput) (models.Bill, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "CreateBillRepository", trace.WithAttributes(attribute.String("repository", "CreateBill")))
		bill         models.Bill
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Transaction(func(tx *gorm.DB) error {
				// Serialize bill creation per store so concurrent requests can't
				// compute the same book/receipt number (Postgres advisory lock,
				// released automatically at transaction end).
				if txErr := tx.Exec("SELECT pg_advisory_xact_lock(?)", input.StoreID).Error; txErr != nil {
					return txErr
				}

				// Price each line item from the store's currently effective price.
				now := time.Now()
				billItems := make([]models.BillItem, 0, len(input.Items))
				var total float64
				for _, item := range input.Items {
					var price models.StoreProductPrice
					if txErr := tx.
						Where("store_id = ? AND product_id = ? AND effective_from <= ?", input.StoreID, item.ProductID, now).
						Order("effective_from DESC").
						First(&price).Error; txErr != nil {
						if txErr == gorm.ErrRecordNotFound {
							return exception.ErrPriceNotConfigured
						}
						return txErr
					}

					subtotal := item.Kilogram * price.Price
					billItems = append(billItems, models.BillItem{
						ProductID: item.ProductID,
						Kilogram:  item.Kilogram,
						Price:     price.Price,
						Subtotal:  subtotal,
					})
					total += subtotal
				}
				total = total - input.Discount + input.ShippingFee

				// Determine next book/receipt number, resetting every calendar year.
				startOfYear := time.Date(now.Year(), time.January, 1, 0, 0, 0, 0, time.Local)
				var lastBill models.Bill
				bookNo, receiptNo := 1, 1
				txErr := tx.
					Where("store_id = ? AND created_at >= ?", input.StoreID, startOfYear).
					Order("id DESC").
					First(&lastBill).Error
				if txErr == nil {
					if lastBill.ReceiptNo >= MaxReceiptNoPerBook {
						bookNo = lastBill.BookNo + 1
						receiptNo = 1
					} else {
						bookNo = lastBill.BookNo
						receiptNo = lastBill.ReceiptNo + 1
					}
				} else if txErr != gorm.ErrRecordNotFound {
					return txErr
				}

				// Resolve (or create) the customer, keeping its default address
				// up to date with the latest one used.
				customerID, txErr := r.resolveCustomer(tx, input.CustomerName, input.CustomerAddress)
				if txErr != nil {
					return txErr
				}

				bill = models.Bill{
					StoreID:         input.StoreID,
					CustomerID:      &customerID,
					BookNo:          bookNo,
					ReceiptNo:       receiptNo,
					CustomerName:    input.CustomerName,
					CustomerAddress: input.CustomerAddress,
					Discount:        input.Discount,
					ShippingFee:     input.ShippingFee,
					Total:           total,
					Slip:            input.Slip,
				}

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
			case exception.ErrPriceNotConfigured:
				err = exception.ErrPriceNotConfigured
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

	return bill, nil
}

// UpdateBill replaces a bill's editable fields and line items in one
// transaction. Book/receipt numbers are never touched. Line items are
// deleted and recreated wholesale rather than diffed against the request.
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

				// Price each line item from the store's currently effective price.
				now := time.Now()
				billItems := make([]models.BillItem, 0, len(input.Items))
				var total float64
				for _, item := range input.Items {
					var price models.StoreProductPrice
					if txErr := tx.
						Where("store_id = ? AND product_id = ? AND effective_from <= ?", input.StoreID, item.ProductID, now).
						Order("effective_from DESC").
						First(&price).Error; txErr != nil {
						if txErr == gorm.ErrRecordNotFound {
							return exception.ErrPriceNotConfigured
						}
						return txErr
					}

					subtotal := item.Kilogram * price.Price
					billItems = append(billItems, models.BillItem{
						BillID:    bill.ID,
						ProductID: item.ProductID,
						Kilogram:  item.Kilogram,
						Price:     price.Price,
						Subtotal:  subtotal,
					})
					total += subtotal
				}
				total = total - input.Discount + input.ShippingFee

				// Resolve (or create) the customer, same as create.
				customerID, txErr := r.resolveCustomer(tx, input.CustomerName, input.CustomerAddress)
				if txErr != nil {
					return txErr
				}

				// Replace line items wholesale rather than diffing.
				if txErr := tx.Where("bill_id = ?", bill.ID).Delete(&models.BillItem{}).Error; txErr != nil {
					return txErr
				}
				if txErr := tx.Create(&billItems).Error; txErr != nil {
					return txErr
				}

				bill.StoreID = input.StoreID
				bill.CustomerID = &customerID
				bill.CustomerName = input.CustomerName
				bill.CustomerAddress = input.CustomerAddress
				bill.Discount = input.Discount
				bill.ShippingFee = input.ShippingFee
				bill.Total = total
				if input.Slip != nil {
					bill.Slip = *input.Slip
				}

				if txErr := tx.Save(&bill).Error; txErr != nil {
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
			case exception.ErrPriceNotConfigured:
				err = exception.ErrPriceNotConfigured
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
	if err = r.db.Preload("Store").Preload("Items.Product").First(&bill, bill.ID).Error; err != nil {
		return bill, err
	}

	return bill, nil
}

// resolveCustomer finds a customer by exact name, creating the customer and/or
// a new address when needed. If the name already exists but this address is
// new, a new CustomerAddress row is added and marked as the default (the one
// used to prefill future bills); an exact name+address match is left as-is.
func (r Repository) resolveCustomer(tx *gorm.DB, name string, address string) (int, error) {
	var customer models.Customer
	err := tx.Where("name = ?", name).First(&customer).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return 0, err
	}

	// New customer: create it plus its first (default) address.
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

		return customer.ID, nil
	}

	// Existing customer: check whether this address is already on file.
	var existingAddress models.CustomerAddress
	err = tx.Where("customer_id = ? AND address = ?", customer.ID, address).First(&existingAddress).Error
	if err == nil {
		// Address already exists for this customer — nothing to do.
		return customer.ID, nil
	}
	if err != gorm.ErrRecordNotFound {
		return 0, err
	}

	// New address for an existing customer — add it and make it the default.
	if err = tx.Model(&models.CustomerAddress{}).
		Where("customer_id = ?", customer.ID).
		Update("is_default", false).Error; err != nil {
		return 0, err
	}

	if err = tx.Create(&models.CustomerAddress{
		CustomerID: customer.ID,
		Address:    address,
		IsDefault:  true,
	}).Error; err != nil {
		return 0, err
	}

	return customer.ID, nil
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
