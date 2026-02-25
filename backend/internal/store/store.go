package store

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrConflict = errors.New("conflict")

type Store struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, databaseURL string) (*Store, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}
	return &Store{pool: pool}, nil
}

func (s *Store) Close() { s.pool.Close() }

type User struct {
	ID           int64     `json:"id"`
	Role         string    `json:"role"`
	FullName     string    `json:"full_name"`
	Phone        *string   `json:"phone,omitempty"`
	Email        *string   `json:"email,omitempty"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

type Pool struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Address  string `json:"address"`
	Timezone string `json:"timezone"`
}

type TrainingType struct {
	ID              int64  `json:"id"`
	Name            string `json:"name"`
	DurationMinutes int    `json:"duration_minutes"`
	DefaultCapacity int    `json:"default_capacity"`
	Price           string `json:"price"`
}

type Slot struct {
	ID             int64     `json:"id"`
	PoolID         int64     `json:"pool_id"`
	PoolName       string    `json:"pool_name"`
	TrainingType   string    `json:"training_type"`
	StartsAt       time.Time `json:"starts_at"`
	EndsAt         time.Time `json:"ends_at"`
	Capacity       int       `json:"capacity"`
	BookedCount    int       `json:"booked_count"`
	FreeCount      int       `json:"free_count"`
	Price          string    `json:"price"`
	Status         string    `json:"status"`
	TrainingTypeID int64     `json:"training_type_id"`
}

type Booking struct {
	ID            int64      `json:"id"`
	UserID        int64      `json:"user_id"`
	SlotID        int64      `json:"slot_id"`
	SeatsCount    int        `json:"seats_count"`
	Status        string     `json:"status"`
	ReservedUntil *time.Time `json:"reserved_until,omitempty"`
	Price         string     `json:"price"`
	CreatedAt     time.Time  `json:"created_at"`
	StartsAt      time.Time  `json:"starts_at"`
	PoolName      string     `json:"pool_name"`
	TrainingType  string     `json:"training_type"`
}

type Payment struct {
	ID         int64     `json:"id"`
	BookingID  int64     `json:"booking_id"`
	Status     string    `json:"status"`
	Amount     string    `json:"amount"`
	Currency   string    `json:"currency"`
	Provider   string    `json:"provider"`
	ExternalID string    `json:"external_id"`
	CreatedAt  time.Time `json:"created_at"`
}

func (s *Store) CreateUser(ctx context.Context, fullName string, phone, email *string, passwordHash string) (User, error) {
	var u User
	q := `
INSERT INTO users (full_name, phone, email, password_hash)
VALUES ($1,$2,$3,$4)
RETURNING id, role, full_name, phone, email, password_hash, created_at`
	err := s.pool.QueryRow(ctx, q, fullName, phone, email, passwordHash).
		Scan(&u.ID, &u.Role, &u.FullName, &u.Phone, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if isUnique(err) {
		return User{}, ErrConflict
	}
	return u, err
}

func (s *Store) GetUserByLogin(ctx context.Context, login string) (User, error) {
	var u User
	q := `
SELECT id, role, full_name, phone, email, password_hash, created_at
FROM users
WHERE is_active = true
  AND (email = $1 OR phone = $1)
LIMIT 1`
	err := s.pool.QueryRow(ctx, q, login).
		Scan(&u.ID, &u.Role, &u.FullName, &u.Phone, &u.Email, &u.PasswordHash, &u.CreatedAt)
	return u, err
}

func (s *Store) GetUserByID(ctx context.Context, userID int64) (User, error) {
	var u User
	q := `SELECT id, role, full_name, phone, email, password_hash, created_at FROM users WHERE id=$1`
	err := s.pool.QueryRow(ctx, q, userID).
		Scan(&u.ID, &u.Role, &u.FullName, &u.Phone, &u.Email, &u.PasswordHash, &u.CreatedAt)
	return u, err
}

func (s *Store) UpdateUser(ctx context.Context, userID int64, fullName, phone, email *string) (User, error) {
	var u User
	q := `
UPDATE users
SET full_name = COALESCE($2, full_name),
    phone = COALESCE($3, phone),
    email = COALESCE($4, email)
WHERE id = $1
RETURNING id, role, full_name, phone, email, password_hash, created_at`
	err := s.pool.QueryRow(ctx, q, userID, fullName, phone, email).
		Scan(&u.ID, &u.Role, &u.FullName, &u.Phone, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if isUnique(err) {
		return User{}, ErrConflict
	}
	return u, err
}

func (s *Store) ListPools(ctx context.Context) ([]Pool, error) {
	q := `SELECT id, name, address, timezone FROM pools WHERE is_active = true ORDER BY name`
	rows, err := s.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]Pool, 0)
	for rows.Next() {
		var p Pool
		if err := rows.Scan(&p.ID, &p.Name, &p.Address, &p.Timezone); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *Store) ListSchedule(ctx context.Context, poolID int64, dateFrom, dateTo time.Time) ([]Slot, error) {
	q := `
SELECT s.id, s.pool_id, p.name, t.name, s.starts_at, s.ends_at,
       s.capacity, COALESCE(bc.booked_count, 0) AS booked_count, (s.capacity - COALESCE(bc.booked_count, 0)) as free_count,
       s.price::text, s.status::text, s.training_type_id
FROM slots s
JOIN pools p ON p.id = s.pool_id
JOIN training_types t ON t.id = s.training_type_id
LEFT JOIN (
  SELECT slot_id, COALESCE(SUM(seats_count), 0)::int AS booked_count
  FROM bookings
  WHERE status IN ('pending_payment','reserved','confirmed','attended','no_show')
  GROUP BY slot_id
) bc ON bc.slot_id = s.id
WHERE s.pool_id = $1
  AND s.starts_at >= $2
  AND s.starts_at < $3
ORDER BY s.starts_at`
	rows, err := s.pool.Query(ctx, q, poolID, dateFrom, dateTo)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]Slot, 0)
	for rows.Next() {
		var s Slot
		if err := rows.Scan(&s.ID, &s.PoolID, &s.PoolName, &s.TrainingType,
			&s.StartsAt, &s.EndsAt, &s.Capacity, &s.BookedCount, &s.FreeCount,
			&s.Price, &s.Status, &s.TrainingTypeID); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (s *Store) ListTrainingTypes(ctx context.Context) ([]TrainingType, error) {
	q := `
SELECT id, name, duration_minutes, default_capacity, price::text
FROM training_types
WHERE is_active = true
ORDER BY id`
	rows, err := s.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]TrainingType, 0)
	for rows.Next() {
		var t TrainingType
		if err := rows.Scan(&t.ID, &t.Name, &t.DurationMinutes, &t.DefaultCapacity, &t.Price); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (s *Store) CreateBooking(ctx context.Context, userID, slotID int64, reserveTTLMin int) (Booking, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return Booking{}, err
	}
	defer tx.Rollback(ctx)

	var b Booking
	q := `
INSERT INTO bookings (user_id, slot_id, status, reserved_until, price)
SELECT $1, s.id, 'pending_payment', NOW() + ($3::int || ' minutes')::interval, s.price
FROM slots s
WHERE s.id = $2
ON CONFLICT (user_id, slot_id)
DO UPDATE
SET status = 'pending_payment',
    reserved_until = NOW() + ($3::int || ' minutes')::interval,
    price = EXCLUDED.price,
    cancel_reason = NULL,
    seats_count = CASE
      WHEN bookings.status IN ('cancelled', 'expired') THEN 1
      ELSE bookings.seats_count + 1
    END
WHERE bookings.status IN ('pending_payment','reserved','confirmed','attended','no_show','cancelled','expired')
RETURNING id, user_id, slot_id, seats_count, status::text, reserved_until, price::text, created_at`
	err = tx.QueryRow(ctx, q, userID, slotID, reserveTTLMin).
		Scan(&b.ID, &b.UserID, &b.SlotID, &b.SeatsCount, &b.Status, &b.ReservedUntil, &b.Price, &b.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || isUnique(err) {
			return Booking{}, ErrConflict
		}
		return Booking{}, err
	}

	q2 := `
SELECT s.starts_at, p.name, t.name
FROM slots s
JOIN pools p ON p.id = s.pool_id
JOIN training_types t ON t.id = s.training_type_id
WHERE s.id = $1`
	if err := tx.QueryRow(ctx, q2, slotID).Scan(&b.StartsAt, &b.PoolName, &b.TrainingType); err != nil {
		return Booking{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return Booking{}, err
	}
	return b, nil
}

func (s *Store) ListMyBookings(ctx context.Context, userID int64) ([]Booking, error) {
	q := `
SELECT b.id, b.user_id, b.slot_id, b.seats_count, b.status::text, b.reserved_until, b.price::text, b.created_at,
       s.starts_at, p.name, t.name
FROM bookings b
JOIN slots s ON s.id = b.slot_id
JOIN pools p ON p.id = s.pool_id
JOIN training_types t ON t.id = s.training_type_id
WHERE b.user_id = $1
  AND b.status <> 'cancelled'
ORDER BY s.starts_at DESC`
	rows, err := s.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]Booking, 0)
	for rows.Next() {
		var b Booking
		if err := rows.Scan(&b.ID, &b.UserID, &b.SlotID, &b.SeatsCount, &b.Status, &b.ReservedUntil, &b.Price, &b.CreatedAt, &b.StartsAt, &b.PoolName, &b.TrainingType); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (s *Store) CancelBooking(ctx context.Context, userID, bookingID int64) error {
	q := `
UPDATE bookings
SET status='cancelled', reserved_until=NULL, cancel_reason='cancelled_by_user'
WHERE id = $1 AND user_id = $2
  AND status IN ('pending_payment','reserved','confirmed')`
	ct, err := s.pool.Exec(ctx, q, bookingID, userID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s *Store) AdjustBookingSeats(ctx context.Context, userID, bookingID int64, action string) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var seatsCount int
	var status string
	if err := tx.QueryRow(ctx,
		`SELECT seats_count, status::text FROM bookings WHERE id=$1 AND user_id=$2 FOR UPDATE`,
		bookingID, userID,
	).Scan(&seatsCount, &status); err != nil {
		return err
	}

	if status != "pending_payment" && status != "reserved" && status != "confirmed" {
		return ErrConflict
	}

	switch action {
	case "inc":
		if _, err := tx.Exec(ctx, `UPDATE bookings SET seats_count = seats_count + 1 WHERE id=$1`, bookingID); err != nil {
			return err
		}
	case "dec":
		if seatsCount > 1 {
			if _, err := tx.Exec(ctx, `UPDATE bookings SET seats_count = seats_count - 1 WHERE id=$1`, bookingID); err != nil {
				return err
			}
		} else {
			if _, err := tx.Exec(ctx,
				`UPDATE bookings SET status='cancelled', reserved_until=NULL, cancel_reason='cancelled_by_user' WHERE id=$1`,
				bookingID,
			); err != nil {
				return err
			}
		}
	default:
		return ErrConflict
	}

	return tx.Commit(ctx)
}

func (s *Store) InitPayment(ctx context.Context, bookingID int64, provider, externalID string) (Payment, error) {
	var p Payment
	q := `
INSERT INTO payments (booking_id, provider, amount, currency, status, external_id)
SELECT b.id, $2, b.price, 'RUB', 'pending', $3
FROM bookings b
WHERE b.id = $1
RETURNING id, booking_id, status::text, amount::text, currency, provider, external_id, created_at`
	err := s.pool.QueryRow(ctx, q, bookingID, provider, externalID).
		Scan(&p.ID, &p.BookingID, &p.Status, &p.Amount, &p.Currency, &p.Provider, &p.ExternalID, &p.CreatedAt)
	if isUnique(err) {
		return Payment{}, ErrConflict
	}
	return p, err
}

func (s *Store) ApplyPaymentWebhook(ctx context.Context, externalID, status string) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var bookingID int64
	q := `UPDATE payments SET status=$2::payment_status, paid_at=CASE WHEN $2='paid' THEN NOW() ELSE paid_at END WHERE external_id=$1 RETURNING booking_id`
	if err := tx.QueryRow(ctx, q, externalID, status).Scan(&bookingID); err != nil {
		return err
	}

	if status == "paid" {
		if _, err := tx.Exec(ctx, `UPDATE bookings SET status='confirmed', reserved_until=NULL WHERE id=$1`, bookingID); err != nil {
			return err
		}
	}
	if status == "failed" || status == "cancelled" {
		if _, err := tx.Exec(ctx, `UPDATE bookings SET status='expired' WHERE id=$1 AND status IN ('pending_payment','reserved')`, bookingID); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (s *Store) ListAdminBookings(ctx context.Context, status string, limit int) ([]Booking, error) {
	args := []any{}
	conds := []string{"1=1"}
	if status != "" {
		args = append(args, status)
		conds = append(conds, fmt.Sprintf("b.status = $%d::booking_status", len(args)))
	}
	args = append(args, limit)
	query := fmt.Sprintf(`
SELECT b.id, b.user_id, b.slot_id, b.status::text, b.reserved_until, b.price::text, b.created_at,
       s.starts_at, p.name
FROM bookings b
JOIN slots s ON s.id = b.slot_id
JOIN pools p ON p.id = s.pool_id
WHERE %s
ORDER BY b.created_at DESC
LIMIT $%d`, strings.Join(conds, " AND "), len(args))

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]Booking, 0)
	for rows.Next() {
		var b Booking
		if err := rows.Scan(&b.ID, &b.UserID, &b.SlotID, &b.Status, &b.ReservedUntil, &b.Price, &b.CreatedAt, &b.StartsAt, &b.PoolName); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (s *Store) UpdateBookingStatusByAdmin(ctx context.Context, adminID, bookingID int64, status, reason string) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `UPDATE bookings SET status=$2::booking_status, cancel_reason=$3, reserved_until=CASE WHEN $2 IN ('cancelled','confirmed','expired') THEN NULL ELSE reserved_until END WHERE id=$1`, bookingID, status, reason); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO admin_actions (admin_id, entity, entity_id, action, reason) VALUES ($1,'booking',$2,'status_update',$3)`, adminID, bookingID, reason); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Store) CreateSlot(ctx context.Context, poolID, trainingTypeID int64, startsAt time.Time, capacity int, price string) (int64, error) {
	var id int64
	q := `
INSERT INTO slots (pool_id, training_type_id, starts_at, ends_at, capacity, price, status)
VALUES ($1,$2,$3::timestamptz,$3::timestamptz + interval '60 minutes',$4,$5::numeric,'open')
RETURNING id`
	err := s.pool.QueryRow(ctx, q, poolID, trainingTypeID, startsAt, capacity, price).Scan(&id)
	if isUnique(err) {
		return 0, ErrConflict
	}
	return id, err
}

func (s *Store) UpdateSlot(ctx context.Context, slotID int64, capacity *int, status *string, price *string) error {
	q := `
UPDATE slots
SET capacity = COALESCE($2, capacity),
    status = COALESCE($3::slot_status, status),
    price = COALESCE($4::numeric, price)
WHERE id = $1`
	ct, err := s.pool.Exec(ctx, q, slotID, capacity, status, price)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func isUnique(err error) bool {
	if err == nil {
		return false
	}
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
