package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"

	"membership-gym/config"
	"membership-gym/internal/middleware"
	jwthelper "membership-gym/pkg/auth"
	"membership-gym/pkg/datetime"
	apperror "membership-gym/pkg/errors"
	"membership-gym/pkg/hash"
	"membership-gym/pkg/token"
	txtx "membership-gym/pkg/tx"
)

type Service struct {
	cfg   config.Config
	repo  *Repository
	redis *redis.Client
}

func NewService(cfg config.Config, repo *Repository, redisClient *redis.Client) *Service {
	return &Service{cfg: cfg, repo: repo, redis: redisClient}
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (map[string]any, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if email == "" || req.Password == "" {
		return nil, apperror.BadRequest("Email dan password wajib diisi")
	}

	args := []any{email}
	where := "LOWER(u.email)=$1"
	if req.GymPublicID != "" {
		gymID, err := uuid.Parse(req.GymPublicID)
		if err != nil {
			return nil, apperror.BadRequest("gym_public_id tidak valid")
		}
		where += " AND g.public_id=$2"
		args = append(args, gymID)
	}

	rows, err := s.repo.FindLoginUser(ctx, where, args...)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, apperror.Unauthorized("Email atau password salah")
	}

	user := rows[0]
	if ok, _ := user["is_active"].(bool); !ok {
		return nil, apperror.Unauthorized("User nonaktif")
	}
	if status, _ := user["gym_status"].(string); status != "active" {
		return nil, apperror.Unauthorized("Gym nonaktif")
	}
	if !hash.CheckPassword(fmt.Sprint(user["password_hash"]), req.Password) {
		return nil, apperror.Unauthorized("Email atau password salah")
	}

	userID := user["user_id"].(int64)
	gymID := user["gym_id"].(int64)
	claims := jwthelper.Claims{
		UserID:       userID,
		UserPublicID: user["user_public_id"].(uuid.UUID),
		GymID:        gymID,
		GymPublicID:  user["gym_public_id"].(uuid.UUID),
		Role:         fmt.Sprint(user["role"]),
	}
	accessToken, err := jwthelper.GenerateAccessToken(claims, s.cfg.JWTAccessSecret, s.cfg.JWTAccessTTL)
	if err != nil {
		return nil, err
	}
	refreshToken, err := jwthelper.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}
	if err := s.redis.Set(ctx, "refresh:"+refreshToken, fmt.Sprintf("%d:%d", userID, gymID), s.cfg.JWTRefreshTTL).Err(); err != nil {
		return nil, err
	}

	_ = s.repo.TouchLastLogin(ctx, userID)
	return map[string]any{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user": map[string]any{
			"public_id": user["user_public_id"],
			"name":      user["name"],
			"email":     user["email"],
			"role":      user["role"],
		},
		"gym": map[string]any{
			"public_id": user["gym_public_id"],
			"name":      user["gym_name"],
		},
	}, nil
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (map[string]any, error) {
	if refreshToken == "" {
		return nil, apperror.BadRequest("refresh_token wajib diisi")
	}
	key := "refresh:" + refreshToken
	refreshValue, err := s.redis.Get(ctx, key).Result()
	if err != nil {
		return nil, apperror.Unauthorized("Refresh token tidak valid")
	}
	rows, err := s.repo.FindRefreshUser(ctx, refreshValue)
	if err != nil || len(rows) == 0 {
		return nil, apperror.Unauthorized("Refresh token tidak valid")
	}
	row := rows[0]
	claims := jwthelper.Claims{
		UserID:       row["user_id"].(int64),
		UserPublicID: row["user_public_id"].(uuid.UUID),
		GymID:        row["gym_id"].(int64),
		GymPublicID:  row["gym_public_id"].(uuid.UUID),
		Role:         fmt.Sprint(row["role"]),
	}
	accessToken, err := jwthelper.GenerateAccessToken(claims, s.cfg.JWTAccessSecret, s.cfg.JWTAccessTTL)
	if err != nil {
		return nil, err
	}
	return map[string]any{"access_token": accessToken}, nil
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken != "" {
		return s.redis.Del(ctx, "refresh:"+refreshToken).Err()
	}
	return nil
}

func (s *Service) RegisterOwner(ctx context.Context, req RegisterOwnerRequest) (map[string]any, error) {
	email := strings.ToLower(strings.TrimSpace(req.OwnerEmail))
	password := req.OwnerPassword
	if email == "" || password == "" || strings.TrimSpace(req.GymName) == "" || strings.TrimSpace(req.OwnerName) == "" {
		return nil, apperror.BadRequest("Data register owner belum lengkap")
	}
	if len(password) < 6 {
		return nil, apperror.BadRequest("Password minimal 6 karakter")
	}
	existing, err := s.repo.CountUsersByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if existing > 0 {
		return nil, apperror.Conflict("Email owner sudah terdaftar")
	}
	hashed, err := hash.HashPassword(password)
	if err != nil {
		return nil, err
	}
	userID, gymID, err := s.createOwnerGym(ctx, ownerRegistration{
		GymName:      req.GymName,
		GymPhone:     req.GymPhone,
		GymEmail:     req.GymEmail,
		GymAddress:   req.GymAddress,
		OwnerName:    req.OwnerName,
		OwnerEmail:   email,
		PasswordHash: hashed,
		AuthProvider: "password",
	})
	if err != nil {
		return nil, err
	}
	return s.issueTokens(ctx, userID, gymID)
}

func (s *Service) GoogleAuthURL(ctx context.Context) (map[string]any, error) {
	if s.cfg.GoogleClientID == "" || s.cfg.GoogleClientSecret == "" {
		return nil, apperror.BadRequest("Google OAuth belum dikonfigurasi")
	}
	state, err := randomState()
	if err != nil {
		return nil, err
	}
	if err := s.redis.Set(ctx, "google_oauth_state:"+state, "1", 10*time.Minute).Err(); err != nil {
		return nil, err
	}
	values := url.Values{}
	values.Set("client_id", s.cfg.GoogleClientID)
	values.Set("redirect_uri", s.cfg.GoogleRedirectURL)
	values.Set("response_type", "code")
	values.Set("scope", "openid email profile")
	values.Set("state", state)
	values.Set("access_type", "offline")
	values.Set("prompt", "select_account")
	return map[string]any{"auth_url": "https://accounts.google.com/o/oauth2/v2/auth?" + values.Encode(), "state": state}, nil
}

func (s *Service) GoogleCallback(ctx context.Context, code, state string) (map[string]any, error) {
	if code == "" || state == "" {
		return nil, apperror.BadRequest("code dan state wajib diisi")
	}
	stateKey := "google_oauth_state:" + state
	if err := s.redis.Get(ctx, stateKey).Err(); err != nil {
		return nil, apperror.Unauthorized("State Google OAuth tidak valid")
	}
	_ = s.redis.Del(ctx, stateKey).Err()
	profile, err := s.fetchGoogleProfile(ctx, code)
	if err != nil {
		return nil, err
	}
	if !profile.EmailVerified {
		return nil, apperror.Unauthorized("Email Google belum terverifikasi")
	}
	email := strings.ToLower(strings.TrimSpace(profile.Email))
	rows, err := s.repo.FindGoogleUser(ctx, profile.Sub, email)
	if err != nil {
		return nil, err
	}
	if len(rows) > 0 {
		userID := rows[0]["user_id"].(int64)
		gymID := rows[0]["gym_id"].(int64)
		if rows[0]["google_sub"] == nil || fmt.Sprint(rows[0]["google_sub"]) == "" {
			_ = s.repo.LinkGoogleUser(ctx, userID, profile.Sub, profile.Picture)
		}
		return s.issueTokens(ctx, userID, gymID)
	}
	passwordHash, err := hash.HashPassword("google:" + profile.Sub)
	if err != nil {
		return nil, err
	}
	gymName := strings.TrimSpace(profile.Name)
	if gymName == "" {
		gymName = strings.Split(email, "@")[0]
	}
	userID, gymID, err := s.createOwnerGym(ctx, ownerRegistration{
		GymName:      gymName + " Gym",
		GymEmail:     email,
		OwnerName:    gymName,
		OwnerEmail:   email,
		PasswordHash: passwordHash,
		AuthProvider: "google",
		GoogleSub:    profile.Sub,
		AvatarURL:    profile.Picture,
	})
	if err != nil {
		return nil, err
	}
	return s.issueTokens(ctx, userID, gymID)
}

func (s *Service) Profile(ctx context.Context, auth middleware.AuthUser) (map[string]any, error) {
	rows, err := s.repo.Profile(ctx, auth.UserID, auth.GymID)
	if err != nil || len(rows) == 0 {
		return nil, apperror.NotFound("Profile tidak ditemukan")
	}
	return rows[0], nil
}

func (s *Service) issueTokens(ctx context.Context, userID, gymID int64) (map[string]any, error) {
	rows, err := s.repo.FindIssueTokenUser(ctx, userID, gymID)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, apperror.Unauthorized("User tidak valid")
	}
	user := rows[0]
	claims := jwthelper.Claims{UserID: userID, UserPublicID: user["user_public_id"].(uuid.UUID), GymID: gymID, GymPublicID: user["gym_public_id"].(uuid.UUID), Role: fmt.Sprint(user["role"])}
	accessToken, err := jwthelper.GenerateAccessToken(claims, s.cfg.JWTAccessSecret, s.cfg.JWTAccessTTL)
	if err != nil {
		return nil, err
	}
	refreshToken, err := jwthelper.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}
	if err := s.redis.Set(ctx, "refresh:"+refreshToken, fmt.Sprintf("%d:%d", userID, gymID), s.cfg.JWTRefreshTTL).Err(); err != nil {
		return nil, err
	}
	_ = s.repo.TouchLastLogin(ctx, userID)
	return map[string]any{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user":          map[string]any{"public_id": user["user_public_id"], "name": user["name"], "email": user["email"], "role": user["role"]},
		"gym":           map[string]any{"public_id": user["gym_public_id"], "name": user["gym_name"]},
	}, nil
}

type ownerRegistration struct {
	GymName      string
	GymPhone     string
	GymEmail     string
	GymAddress   string
	OwnerName    string
	OwnerEmail   string
	PasswordHash string
	AuthProvider string
	GoogleSub    string
	AvatarURL    string
}

func (s *Service) createOwnerGym(ctx context.Context, data ownerRegistration) (int64, int64, error) {
	var userID, gymID int64
	err := txtx.WithTx(ctx, s.repo.DB(), func(tx pgx.Tx) error {
		if err := s.repo.InsertGymReturningID(ctx, tx, token.GeneratePublicID(), data.GymName, nullString(data.GymPhone), nullString(data.GymEmail), nullString(data.GymAddress), s.cfg.DefaultTimezone, s.cfg.DefaultCurrency, &gymID); err != nil {
			return err
		}
		roleID, err := s.ensureDefaultRoles(ctx, tx, gymID)
		if err != nil {
			return err
		}
		if err := s.repo.InsertOwnerUserReturningID(ctx, tx, token.GeneratePublicID(), gymID, data.OwnerName, data.OwnerEmail, data.PasswordHash, roleID, data.AuthProvider, nullString(data.GoogleSub), nullString(data.AvatarURL), &userID); err != nil {
			return err
		}
		if err := s.createBasicTrial(ctx, tx, gymID, userID); err != nil {
			return err
		}
		payload := map[string]any{"gym_id": gymID, "user_id": userID, "email": data.OwnerEmail, "auth_provider": data.AuthProvider}
		return s.repo.LogAudit(ctx, tx, gymID, &userID, "owner.registered", "users", &userID, payload)
	})
	return userID, gymID, err
}

func (s *Service) ensureDefaultRoles(ctx context.Context, tx pgx.Tx, gymID int64) (int64, error) {
	roles := []struct {
		Code string
		Name string
		Desc string
	}{
		{"owner", "Owner", "Full owner access"},
		{"admin", "Admin", "Operational admin access"},
		{"cashier", "Cashier", "Cashier access"},
		{"trainer", "Trainer", "Trainer access"},
	}
	for _, role := range roles {
		if err := s.repo.InsertDefaultRole(ctx, tx, token.GeneratePublicID(), gymID, role.Name, role.Code, role.Desc); err != nil {
			return 0, err
		}
	}
	return s.repo.OwnerRoleID(ctx, tx, gymID)
}

func (s *Service) createBasicTrial(ctx context.Context, tx pgx.Tx, gymID, userID int64) error {
	planID, duration, err := s.repo.BasicSaasPlan(ctx, tx)
	if err != nil {
		return err
	}
	startDate := datetime.TodayInTimezone(s.cfg.DefaultTimezone)
	endDate := startDate.AddDate(0, 0, duration-1)
	var subID int64
	if err := s.repo.InsertGymSubscriptionReturningID(ctx, tx, token.GeneratePublicID(), gymID, planID, startDate, endDate, userID, &subID); err != nil {
		return err
	}
	invoiceNo := fmt.Sprintf("SAAS-%s-0001", datetime.NowInTimezone(s.cfg.DefaultTimezone).Format("20060102"))
	return s.repo.InsertGymSubscriptionFreePayment(ctx, tx, token.GeneratePublicID(), gymID, subID, invoiceNo, s.cfg.DefaultCurrency, userID)
}

func randomState() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

type googleProfile struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

func (s *Service) fetchGoogleProfile(ctx context.Context, code string) (googleProfile, error) {
	form := url.Values{}
	form.Set("code", code)
	form.Set("client_id", s.cfg.GoogleClientID)
	form.Set("client_secret", s.cfg.GoogleClientSecret)
	form.Set("redirect_uri", s.cfg.GoogleRedirectURL)
	form.Set("grant_type", "authorization_code")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://oauth2.googleapis.com/token", strings.NewReader(form.Encode()))
	if err != nil {
		return googleProfile{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return googleProfile{}, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return googleProfile{}, apperror.Unauthorized("Google OAuth token exchange gagal")
	}
	var tokenResp struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return googleProfile{}, err
	}
	userReq, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.googleapis.com/oauth2/v3/userinfo", nil)
	if err != nil {
		return googleProfile{}, err
	}
	userReq.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)
	userResp, err := http.DefaultClient.Do(userReq)
	if err != nil {
		return googleProfile{}, err
	}
	defer userResp.Body.Close()
	userBody, _ := io.ReadAll(userResp.Body)
	if userResp.StatusCode >= 300 {
		return googleProfile{}, apperror.Unauthorized("Gagal mengambil profil Google")
	}
	var profile googleProfile
	if err := json.Unmarshal(userBody, &profile); err != nil {
		return googleProfile{}, err
	}
	if profile.Sub == "" || profile.Email == "" {
		return googleProfile{}, apperror.Unauthorized("Profil Google tidak valid")
	}
	return profile, nil
}
