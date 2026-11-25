package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Claims struct {
	UserID int    `json:"uid"`
	Nama   string `json:"nama"`
	Peran  string `json:"peran"`
	jwt.RegisteredClaims
}

type ctxKey struct{}

type Manager struct{ secret []byte }

func NewManager(secret []byte) *Manager { return &Manager{secret: secret} }

func HashPassword(pw string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	return string(b), err
}

func CheckPassword(hash, pw string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pw)) == nil
}

func (m *Manager) Issue(uid int, nama, peran string) (string, error) {
	c := Claims{UserID: uid, Nama: nama, Peran: peran, RegisteredClaims: jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(10 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(m.secret)
}

func (m *Manager) Parse(tok string) (*Claims, error) {
	c := &Claims{}
	_, err := jwt.ParseWithClaims(tok, c, func(t *jwt.Token) (any, error) { return m.secret, nil },
		jwt.WithValidMethods([]string{"HS256"}))
	if err != nil {
		return nil, err
	}
	return c, nil
}

// Require memastikan request membawa token valid dengan salah satu peran yang diizinkan.
func (m *Manager) Require(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := r.Header.Get("Authorization")
			tok, ok := strings.CutPrefix(h, "Bearer ")
			if !ok {
				http.Error(w, `{"error":"tidak terautentikasi"}`, http.StatusUnauthorized)
				return
			}
			c, err := m.Parse(tok)
			if err != nil {
				http.Error(w, `{"error":"tidak terautentikasi"}`, http.StatusUnauthorized)
				return
			}
			if len(roles) > 0 && !contains(roles, c.Peran) {
				http.Error(w, `{"error":"akses ditolak"}`, http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKey{}, c)))
		})
	}
}

func FromContext(ctx context.Context) (*Claims, error) {
	c, ok := ctx.Value(ctxKey{}).(*Claims)
	if !ok {
		return nil, errors.New("tidak ada sesi")
	}
	return c, nil
}

func contains(xs []string, s string) bool {
	for _, x := range xs {
		if x == s {
			return true
		}
	}
	return false
}
