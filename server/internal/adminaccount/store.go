package adminaccount

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Alievs-corp/hypedomino/server/internal/rbac"
)

var (
	ErrInvalid  = errors.New("adminaccount: invalid account")
	ErrTaken    = errors.New("adminaccount: email already has an account")
	ErrNotFound = errors.New("adminaccount: no such account")
	ErrDisabled = errors.New("adminaccount: account is disabled")
)

type State string

const (
	Active   State = "active"
	Disabled State = "disabled"
)

type Account struct {
	ID          string
	Email       string
	DisplayName string
	Roles       []rbac.Role
	State       State
	CreatedAt   time.Time
	LastLoginAt time.Time
	DisabledAt  time.Time
}

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func validRoles(roles []rbac.Role) error {
	if len(roles) == 0 {
		return fmt.Errorf("%w: at least one role is required", ErrInvalid)
	}
	if slices.ContainsFunc(roles, func(role rbac.Role) bool { return !rbac.IsRole(role) }) {
		return fmt.Errorf("%w: unknown role in %v", ErrInvalid, roles)
	}
	return nil
}

func uniqueRoles(roles []rbac.Role) []rbac.Role {
	sorted := slices.Clone(roles)
	slices.Sort(sorted)
	return slices.Compact(sorted)
}

const columns = `id, email, display_name, roles, state, created_at, last_login_at, disabled_at`

func scan(row pgx.Row) (Account, error) {
	var a Account
	var lastLogin, disabled *time.Time
	if err := row.Scan(&a.ID, &a.Email, &a.DisplayName, &a.Roles, &a.State, &a.CreatedAt, &lastLogin, &disabled); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Account{}, ErrNotFound
		}
		return Account{}, err
	}
	a.CreatedAt = a.CreatedAt.UTC()
	a.LastLoginAt = utcOrZero(lastLogin)
	a.DisabledAt = utcOrZero(disabled)
	return a, nil
}

func utcOrZero(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return t.UTC()
}

func validEmail(email string) bool {
	local, domain, found := strings.Cut(email, "@")
	return found && local != "" && domain != "" && !strings.Contains(domain, "@")
}

func (s *Store) Create(ctx context.Context, email, displayName string, roles []rbac.Role, now time.Time) (Account, error) {
	email = NormalizeEmail(email)
	displayName = strings.TrimSpace(displayName)
	if !validEmail(email) || displayName == "" {
		return Account{}, fmt.Errorf("%w: email and display name are required", ErrInvalid)
	}
	if err := validRoles(roles); err != nil {
		return Account{}, err
	}
	account, err := scan(s.pool.QueryRow(ctx, `insert into admin_accounts (email, display_name, roles, created_at)
		values ($1, $2, $3, $4) returning `+columns, email, displayName, uniqueRoles(roles), now.UTC()))
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return Account{}, ErrTaken
	}
	return account, err
}

func (s *Store) ByEmail(ctx context.Context, email string) (Account, error) {
	return scan(s.pool.QueryRow(ctx, `select `+columns+` from admin_accounts where email = $1`, NormalizeEmail(email)))
}

func (s *Store) SignIn(ctx context.Context, email string, now time.Time) (Account, error) {
	account, err := scan(s.pool.QueryRow(ctx, `update admin_accounts set last_login_at = $2
		where email = $1 and state = 'active' returning `+columns, NormalizeEmail(email), now.UTC()))
	if !errors.Is(err, ErrNotFound) {
		return account, err
	}
	if _, lookupErr := s.ByEmail(ctx, email); lookupErr == nil {
		return Account{}, ErrDisabled
	}
	return Account{}, ErrNotFound
}

func (s *Store) SetRoles(ctx context.Context, id string, roles []rbac.Role) (Account, error) {
	if err := validRoles(roles); err != nil {
		return Account{}, err
	}
	return scan(s.pool.QueryRow(ctx, `update admin_accounts set roles = $2 where id = $1 returning `+columns, id, uniqueRoles(roles)))
}

func (s *Store) Disable(ctx context.Context, id string, now time.Time) (Account, error) {
	return scan(s.pool.QueryRow(ctx, `update admin_accounts set state = 'disabled', disabled_at = coalesce(disabled_at, $2)
		where id = $1 returning `+columns, id, now.UTC()))
}

func (s *Store) DisableInactive(ctx context.Context, now time.Time, inactiveFor time.Duration) ([]Account, error) {
	rows, err := s.pool.Query(ctx, `update admin_accounts set state = 'disabled', disabled_at = $1
		where state = 'active' and coalesce(last_login_at, created_at) < $2
		returning `+columns, now.UTC(), now.Add(-inactiveFor).UTC())
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (Account, error) { return scan(row) })
}
