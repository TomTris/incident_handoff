package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	Users  UserStore
	Secret []byte
	TTL    time.Duration
	Now    func() time.Time
}

func NewAuthHandler(users UserStore, secret []byte, ttl time.Duration) *AuthHandler {
	return &AuthHandler{
		Users:  users,
		Secret: secret,
		TTL:    ttl,
		Now:    time.Now}
}

func (h *AuthHandler) RegistrationHandler(r *http.Request) (*AppResponse, *AppError) {
	u := UserRegistration{}
	// body, err := io.ReadAll(r.Body)
	// fmt.Println(string(body))

	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		return nil, BadRequest(MalformedRequestBody)
	}

	if err := u.Validate(); err != nil {
		return nil, UnprocessableEntity(err)
	}

	hashedPassword, err := HashPassword(u.Password)
	if err != nil {
		return nil, InternalServerError(err)
	}

	_, err = h.Users.Create(r.Context(), User{
		Username: u.Username,
		Password: hashedPassword,
		Role:     u.Role,
	})
	if err != nil {
		if errors.Is(err, ErrUserAlreadyExist) {
			return nil, Conflict(err)
		}
		return nil, InternalServerError(err)
	}
	return newAppResponse(http.StatusNoContent, nil), nil
}

func (h *AuthHandler) LoginHandler(w http.ResponseWriter, r *http.Request) (*AppResponse, *AppError) {
	u := UserLogin{}
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		return nil, BadRequest(MalformedRequestBody)
	}

	user, err := h.Users.GetByUsername(r.Context(), u.Username)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, Unauthorized(errors.New("Username or Password not correct"))
		}
		return nil, Unauthorized(err)
	}
	if err := VerifyPassword(user.Password, u.Password); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return nil, Unauthorized(errors.New("Username or Password not correct"))
		}
		return nil, InternalServerError(err)
	}

	token, err := IssueToken(user, h.Secret, h.TTL, h.Now())
	if err != nil {
		return nil, InternalServerError(errors.New("Token signing failed"))
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    token,
		Path:     "/",
		MaxAge:   int(h.TTL.Seconds()),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	return newAppResponse(http.StatusOK, map[string]string{"status": "ok"}), nil
}

func (h *AuthHandler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	requestID, _ := r.Context().Value(requestIDKey).(string)
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
	writeJSON(w, http.StatusOK, requestID, map[string]string{"status": "ok"})
}
func (h *AuthHandler) WhoAmI(r *http.Request) (*AppResponse, *AppError) {
	claims := r.Context().Value(userContextKey).(UserContext)
	return newAppResponse(http.StatusOK, claims), nil
}

func (h *AuthHandler) IsAuthenticated(r *http.Request) (*AppResponse, *AppError) {
	return newAppResponse(http.StatusNoContent, nil), nil
}
