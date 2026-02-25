package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"wavetime/backend/internal/auth"
	"wavetime/backend/internal/config"
	"wavetime/backend/internal/store"
)

type API struct {
	cfg   config.Config
	store *store.Store
	mux   *http.ServeMux
}

type contextKey string

const claimsKey contextKey = "claims"

func New(cfg config.Config, st *store.Store) *API {
	a := &API{cfg: cfg, store: st, mux: http.NewServeMux()}
	a.routes()
	return a
}

func (a *API) Handler() http.Handler {
	return a.logMiddleware(a.mux)
}

func (a *API) routes() {
	a.mux.HandleFunc("GET /healthz", a.handleHealth)

	a.mux.HandleFunc("POST /auth/register", a.handleRegister)
	a.mux.HandleFunc("POST /auth/login", a.handleLogin)
	a.mux.HandleFunc("POST /auth/refresh", a.authRequired(a.handleRefresh))
	a.mux.HandleFunc("POST /auth/logout", a.authRequired(a.handleLogout))

	a.mux.HandleFunc("GET /me", a.authRequired(a.handleMe))
	a.mux.HandleFunc("PATCH /me", a.authRequired(a.handleMeUpdate))

	a.mux.HandleFunc("GET /pools", a.authRequired(a.handlePools))
	a.mux.HandleFunc("GET /schedule", a.authRequired(a.handleSchedule))

	a.mux.HandleFunc("POST /bookings", a.authRequired(a.handleBookingCreate))
	a.mux.HandleFunc("GET /bookings/my", a.authRequired(a.handleMyBookings))
	a.mux.HandleFunc("DELETE /bookings/", a.authRequired(a.handleBookingDelete))

	a.mux.HandleFunc("POST /payments/init", a.authRequired(a.handlePaymentInit))
	a.mux.HandleFunc("POST /payments/webhook", a.handlePaymentWebhook)

	a.mux.HandleFunc("GET /admin/bookings", a.authRequired(a.adminOnly(a.handleAdminBookings)))
	a.mux.HandleFunc("PATCH /admin/bookings/", a.authRequired(a.adminOnly(a.handleAdminBookingUpdate)))
	a.mux.HandleFunc("POST /admin/slots", a.authRequired(a.adminOnly(a.handleAdminSlotCreate)))
	a.mux.HandleFunc("PATCH /admin/slots/", a.authRequired(a.adminOnly(a.handleAdminSlotUpdate)))
}

func (a *API) handleHealth(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (a *API) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req struct {
		FullName string  `json:"full_name"`
		Phone    *string `json:"phone"`
		Email    *string `json:"email"`
		Password string  `json:"password"`
	}
	if err := decodeJSON(r, &req); err != nil {
		errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	if strings.TrimSpace(req.FullName) == "" {
		errorResponse(w, http.StatusBadRequest, "full_name is required")
		return
	}

	h, err := auth.HashPassword(req.Password)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	u, err := a.store.CreateUser(r.Context(), req.FullName, req.Phone, req.Email, h)
	if errors.Is(err, store.ErrConflict) {
		errorResponse(w, http.StatusConflict, "user with provided phone/email already exists")
		return
	}
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	token, err := auth.IssueToken(a.cfg.AppSecret, u.ID, u.Role, a.cfg.TokenTTL)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "failed to issue token")
		return
	}
	jsonResponse(w, http.StatusCreated, map[string]any{"token": token, "user": u})
}

func (a *API) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	if err := decodeJSON(r, &req); err != nil {
		errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	u, err := a.store.GetUserByLogin(r.Context(), req.Login)
	if err != nil {
		errorResponse(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if !auth.VerifyPassword(u.PasswordHash, req.Password) {
		errorResponse(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token, err := auth.IssueToken(a.cfg.AppSecret, u.ID, u.Role, a.cfg.TokenTTL)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "failed to issue token")
		return
	}
	jsonResponse(w, http.StatusOK, map[string]any{"token": token, "user": u})
}

func (a *API) handleRefresh(w http.ResponseWriter, r *http.Request) {
	cl := claimsFromCtx(r.Context())
	token, err := auth.IssueToken(a.cfg.AppSecret, cl.UserID, cl.Role, a.cfg.TokenTTL)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "failed to issue token")
		return
	}
	jsonResponse(w, http.StatusOK, map[string]string{"token": token})
}

func (a *API) handleLogout(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) handleMe(w http.ResponseWriter, r *http.Request) {
	cl := claimsFromCtx(r.Context())
	u, err := a.store.GetUserByID(r.Context(), cl.UserID)
	if err != nil {
		errorResponse(w, http.StatusNotFound, "user not found")
		return
	}
	jsonResponse(w, http.StatusOK, u)
}

func (a *API) handleMeUpdate(w http.ResponseWriter, r *http.Request) {
	cl := claimsFromCtx(r.Context())
	var req struct {
		FullName *string `json:"full_name"`
		Phone    *string `json:"phone"`
		Email    *string `json:"email"`
	}
	if err := decodeJSON(r, &req); err != nil {
		errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	u, err := a.store.UpdateUser(r.Context(), cl.UserID, req.FullName, req.Phone, req.Email)
	if errors.Is(err, store.ErrConflict) {
		errorResponse(w, http.StatusConflict, "phone/email already used")
		return
	}
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "failed to update profile")
		return
	}
	jsonResponse(w, http.StatusOK, u)
}

func (a *API) handlePools(w http.ResponseWriter, r *http.Request) {
	pools, err := a.store.ListPools(r.Context())
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "failed to list pools")
		return
	}
	jsonResponse(w, http.StatusOK, pools)
}

func (a *API) handleSchedule(w http.ResponseWriter, r *http.Request) {
	poolID, err := parseInt64(r.URL.Query().Get("pool_id"))
	if err != nil || poolID <= 0 {
		errorResponse(w, http.StatusBadRequest, "pool_id is required")
		return
	}
	dateFrom, dateTo, err := parseDateRange(r.URL.Query().Get("date_from"), r.URL.Query().Get("date_to"))
	if err != nil {
		errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	slots, err := a.store.ListSchedule(r.Context(), poolID, dateFrom, dateTo)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "failed to list schedule")
		return
	}
	jsonResponse(w, http.StatusOK, slots)
}

func (a *API) handleBookingCreate(w http.ResponseWriter, r *http.Request) {
	cl := claimsFromCtx(r.Context())
	var req struct {
		SlotID int64 `json:"slot_id"`
	}
	if err := decodeJSON(r, &req); err != nil {
		errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.SlotID <= 0 {
		errorResponse(w, http.StatusBadRequest, "slot_id is required")
		return
	}

	b, err := a.store.CreateBooking(r.Context(), cl.UserID, req.SlotID, a.cfg.ReserveTTLMin)
	if errors.Is(err, store.ErrConflict) {
		errorResponse(w, http.StatusConflict, "booking already exists")
		return
	}
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "failed to create booking")
		return
	}
	jsonResponse(w, http.StatusCreated, b)
}

func (a *API) handleMyBookings(w http.ResponseWriter, r *http.Request) {
	cl := claimsFromCtx(r.Context())
	bookings, err := a.store.ListMyBookings(r.Context(), cl.UserID)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "failed to list bookings")
		return
	}
	jsonResponse(w, http.StatusOK, bookings)
}

func (a *API) handleBookingDelete(w http.ResponseWriter, r *http.Request) {
	cl := claimsFromCtx(r.Context())
	bookingID, err := parseIDFromPath(r.URL.Path, "/bookings/")
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "invalid booking id")
		return
	}
	if err := a.store.CancelBooking(r.Context(), cl.UserID, bookingID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			errorResponse(w, http.StatusNotFound, "booking not found")
			return
		}
		errorResponse(w, http.StatusBadRequest, "failed to cancel booking")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) handlePaymentInit(w http.ResponseWriter, r *http.Request) {
	var req struct {
		BookingID int64  `json:"booking_id"`
		Provider  string `json:"provider"`
	}
	if err := decodeJSON(r, &req); err != nil {
		errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.BookingID <= 0 {
		errorResponse(w, http.StatusBadRequest, "booking_id is required")
		return
	}
	if req.Provider == "" {
		req.Provider = "mock"
	}
	externalID := fmt.Sprintf("pay_%d", time.Now().UnixNano())
	p, err := a.store.InitPayment(r.Context(), req.BookingID, req.Provider, externalID)
	if errors.Is(err, store.ErrConflict) {
		errorResponse(w, http.StatusConflict, "payment already exists")
		return
	}
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "failed to init payment")
		return
	}
	jsonResponse(w, http.StatusCreated, p)
}

func (a *API) handlePaymentWebhook(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ExternalID string `json:"external_id"`
		Status     string `json:"status"`
	}
	if err := decodeJSON(r, &req); err != nil {
		errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.ExternalID == "" || req.Status == "" {
		errorResponse(w, http.StatusBadRequest, "external_id and status are required")
		return
	}
	if err := a.store.ApplyPaymentWebhook(r.Context(), req.ExternalID, req.Status); err != nil {
		errorResponse(w, http.StatusBadRequest, "failed to apply webhook")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) handleAdminBookings(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	limit := 50
	if s := r.URL.Query().Get("limit"); s != "" {
		v, err := strconv.Atoi(s)
		if err == nil && v > 0 && v <= 500 {
			limit = v
		}
	}
	bookings, err := a.store.ListAdminBookings(r.Context(), status, limit)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "failed to list admin bookings")
		return
	}
	jsonResponse(w, http.StatusOK, bookings)
}

func (a *API) handleAdminBookingUpdate(w http.ResponseWriter, r *http.Request) {
	cl := claimsFromCtx(r.Context())
	bookingID, err := parseIDFromPath(r.URL.Path, "/admin/bookings/")
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "invalid booking id")
		return
	}
	var req struct {
		Status string `json:"status"`
		Reason string `json:"reason"`
	}
	if err := decodeJSON(r, &req); err != nil {
		errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.Status == "" {
		errorResponse(w, http.StatusBadRequest, "status is required")
		return
	}
	if err := a.store.UpdateBookingStatusByAdmin(r.Context(), cl.UserID, bookingID, req.Status, req.Reason); err != nil {
		errorResponse(w, http.StatusBadRequest, "failed to update booking")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) handleAdminSlotCreate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PoolID         int64  `json:"pool_id"`
		TrainingTypeID int64  `json:"training_type_id"`
		StartsAt       string `json:"starts_at"`
		Capacity       int    `json:"capacity"`
		Price          string `json:"price"`
	}
	if err := decodeJSON(r, &req); err != nil {
		errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	startsAt, err := time.Parse(time.RFC3339, req.StartsAt)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "starts_at must be RFC3339")
		return
	}
	id, err := a.store.CreateSlot(r.Context(), req.PoolID, req.TrainingTypeID, startsAt, req.Capacity, req.Price)
	if errors.Is(err, store.ErrConflict) {
		errorResponse(w, http.StatusConflict, "slot already exists")
		return
	}
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "failed to create slot")
		return
	}
	jsonResponse(w, http.StatusCreated, map[string]int64{"id": id})
}

func (a *API) handleAdminSlotUpdate(w http.ResponseWriter, r *http.Request) {
	slotID, err := parseIDFromPath(r.URL.Path, "/admin/slots/")
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "invalid slot id")
		return
	}
	var req struct {
		Capacity *int    `json:"capacity"`
		Status   *string `json:"status"`
		Price    *string `json:"price"`
	}
	if err := decodeJSON(r, &req); err != nil {
		errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := a.store.UpdateSlot(r.Context(), slotID, req.Capacity, req.Status, req.Price); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			errorResponse(w, http.StatusNotFound, "slot not found")
			return
		}
		errorResponse(w, http.StatusBadRequest, "failed to update slot")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) authRequired(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			errorResponse(w, http.StatusUnauthorized, "missing bearer token")
			return
		}
		token := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := auth.ParseToken(a.cfg.AppSecret, token)
		if err != nil {
			errorResponse(w, http.StatusUnauthorized, "invalid token")
			return
		}
		ctx := context.WithValue(r.Context(), claimsKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func (a *API) adminOnly(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cl := claimsFromCtx(r.Context())
		if cl.Role != "admin" {
			errorResponse(w, http.StatusForbidden, "admin only")
			return
		}
		next.ServeHTTP(w, r)
	}
}

func (a *API) logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		_ = start
	})
}

func claimsFromCtx(ctx context.Context) auth.Claims {
	v := ctx.Value(claimsKey)
	if cl, ok := v.(auth.Claims); ok {
		return cl
	}
	return auth.Claims{}
}

func decodeJSON(r *http.Request, out any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(out); err != nil {
		return fmt.Errorf("invalid json: %w", err)
	}
	return nil
}

func jsonResponse(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func errorResponse(w http.ResponseWriter, status int, message string) {
	jsonResponse(w, status, map[string]string{"error": message})
}

func parseDateRange(fromStr, toStr string) (time.Time, time.Time, error) {
	if fromStr == "" || toStr == "" {
		return time.Time{}, time.Time{}, fmt.Errorf("date_from and date_to are required (YYYY-MM-DD)")
	}
	from, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid date_from")
	}
	to, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid date_to")
	}
	if !to.After(from) {
		return time.Time{}, time.Time{}, fmt.Errorf("date_to must be greater than date_from")
	}
	if to.Sub(from) > 31*24*time.Hour {
		return time.Time{}, time.Time{}, fmt.Errorf("max range is 31 days")
	}
	return from, to, nil
}

func parseInt64(v string) (int64, error) {
	return strconv.ParseInt(v, 10, 64)
}

func parseIDFromPath(path, prefix string) (int64, error) {
	if !strings.HasPrefix(path, prefix) {
		return 0, fmt.Errorf("bad path")
	}
	raw := strings.TrimPrefix(path, prefix)
	if strings.Contains(raw, "/") {
		return 0, fmt.Errorf("bad path")
	}
	return strconv.ParseInt(raw, 10, 64)
}
