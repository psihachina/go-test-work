package apiserver

import (
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/psihachina/go-test-work.git/internal/app/model"
	"github.com/psihachina/go-test-work.git/internal/app/store"
	"github.com/sirupsen/logrus"
	"github.com/twinj/uuid"
	"net/http"
	"os"
	"strings"
	"time"
)

type server struct {
	router *mux.Router
	logger *logrus.Logger
	store  store.Store
}

func newServer(store store.Store) *server {
	s := &server{
		router: mux.NewRouter(),
		logger: logrus.New(),
		store:  store,
	}
	s.configureRouter()
	return s
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *server) configureRouter() {
	s.router.HandleFunc("/Login", s.handleSessionsCreate()).Methods("POST")
	s.router.HandleFunc("/Logout", s.handleSessionsDelete()).Methods("POST")
	s.router.HandleFunc("/Refresh", s.HandleSessionsRefresh()).Methods("POST")
	s.router.HandleFunc("/LogoutAll", s.handleAllSessionsDelete()).Methods("POST")
}

func (s *server) HandleSessionsRefresh() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, err := s.VerifyRefreshToken(r)
		if err != nil {
			s.respond(w, r, http.StatusUnauthorized, "Refresh token expired")
			return
		}
		if _, ok := token.Claims.(jwt.Claims); !ok && !token.Valid {
			s.respond(w, r, http.StatusUnauthorized, err)
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if ok && token.Valid {
			refreshUuid, ok := claims["refresh_uuid"].(string)
			if !ok {
				s.respond(w, r, http.StatusUnprocessableEntity, err)
				return
			}
			userId := claims["user_id"].(string)
			if err != nil {
				s.respond(w, r, http.StatusUnprocessableEntity, "Error occurred")
				return
			}

			deleted, delErr := s.store.Token().DeleteAuth(refreshUuid)
			if delErr != nil || deleted == 0 { //if any goes wrong
				s.respond(w, r, http.StatusUnauthorized, "unauthorized")
				return
			}

			ts, createErr := s.Create(userId)
			if createErr != nil {
				s.respond(w, r, http.StatusForbidden, createErr.Error())
				return
			}

			saveErr := s.store.Token().CreateAuth(userId, ts)
			if saveErr != nil {
				s.respond(w, r, http.StatusForbidden, saveErr.Error())
				return
			}
			tokens := map[string]string{
				"access_token":  ts.AccessToken,
				"refresh_token": ts.RefreshToken,
			}

			w.Header().Add("Authorization", ts.AccessToken)

			http.SetCookie(w, &http.Cookie{
				Name:     "refresh_token",
				Value:    ts.RefreshToken,
				Expires:  time.Now().Add(120 * time.Minute),
				HttpOnly: true,
			})

			s.respond(w, r, http.StatusCreated, tokens)
		} else {
			s.respond(w, r, http.StatusUnauthorized, "refresh expired")
		}
	}

}

func (s *server) handleSessionsCreate() http.HandlerFunc {
	type request struct {
		UserId string `json:"id"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		ts, err := s.Create(req.UserId)
		if err != nil {
			s.respond(w, r, http.StatusUnprocessableEntity, err.Error())
			return
		}

		err = s.store.Token().CreateAuth(req.UserId, ts)
		if err != nil {
			s.respond(w, r, http.StatusUnprocessableEntity, err.Error())
		}

		tokens := map[string]string{
			"access_token":  ts.AccessToken,
			"refresh_token": ts.RefreshToken,
		}

		w.Header().Add("Authorization", ts.AccessToken)

		http.SetCookie(w, &http.Cookie{
			Name:     "refresh_token",
			Value:    ts.RefreshToken,
			Expires:  time.Now().Add(120 * time.Minute),
			HttpOnly: true,
		})

		s.respond(w, r, http.StatusOK, tokens)
	}
}

func (s *server) handleSessionsDelete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metadata, err := s.ExtractTokenMetadata(r)
		if err != nil {
			s.respond(w, r, http.StatusUnauthorized, "unauthorized")
			return
		}

		_, delErr := s.store.Token().DeleteAuth(metadata.RefreshUuid)
		if delErr != nil {
			s.respond(w, r, http.StatusUnauthorized, delErr.Error())
			return
		}
		s.respond(w, r, http.StatusOK, "Successfully logged out")
	}
}

func (s *server) handleAllSessionsDelete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metadata, err := s.ExtractTokenMetadata(r)
		if err != nil {
			s.respond(w, r, http.StatusUnauthorized, "unauthorized")
			return
		}

		delErr := s.store.Token().DeleteTokens(metadata)
		if delErr != nil {
			s.respond(w, r, http.StatusUnauthorized, delErr.Error())
			return
		}
		s.respond(w, r, http.StatusOK, "Successfully logged out")
	}
}

func (s *server) ExtractAccessToken(r *http.Request) string {
	bearToken := r.Header.Get("Authorization")
	strArr := strings.Split(bearToken, " ")
	if len(strArr) == 2 {
		return strArr[1]
	}
	return ""
}

func (s *server) ExtractRefreshToken(r *http.Request) string {
	refreshToken, err := r.Cookie("refresh_token")
	if err != nil {
		return ""
	}
	return refreshToken.Value
}

func (s *server) VerifyRefreshToken(r *http.Request) (*jwt.Token, error) {
	tokenString := s.ExtractRefreshToken(r)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("REFRESH_SECRET")), nil
	})
	if err != nil {
		return nil, err
	}

	return token, nil
}

func (s *server) VerifyToken(r *http.Request) (*jwt.Token, error) {
	tokenString := s.ExtractAccessToken(r)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("ACCESS_SECRET")), nil
	})
	if err != nil {
		return nil, err
	}

	return token, nil
}

func (s *server) TokenValid(r *http.Request) error {
	token, err := s.VerifyToken(r)
	if err != nil {
		return err
	}
	if _, ok := token.Claims.(jwt.Claims); !ok || !token.Valid {
		return err
	}
	return nil
}

func (s *server) ExtractTokenMetadata(r *http.Request) (*model.AccessDetails, error) {

	accessToken, err := s.VerifyToken(r)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.VerifyRefreshToken(r)
	if err != nil {
		return nil, err
	}

	claims, ok := accessToken.Claims.(jwt.MapClaims)
	if ok && accessToken.Valid {
		accessUuid, ok := claims["access_uuid"].(string)
		if !ok {
			return nil, err
		}
		userId := claims["user_id"].(string)
		if err != nil {
			return nil, err
		}

		fmt.Println(accessUuid, userId)

		claims, ok := refreshToken.Claims.(jwt.MapClaims)
		if ok && refreshToken.Valid {
			refreshUuid, ok := claims["refresh_uuid"].(string)
			if !ok {
				return nil, err
			}
			fmt.Println(refreshUuid)

			return &model.AccessDetails{
				AccessUuid:  accessUuid,
				UserId:      userId,
				RefreshUuid: refreshUuid,
			}, nil

		}
		return nil, err
	}
	return nil, err
}

func (s *server) Create(userid string) (*model.TokenDetails, error) {
	td := &model.TokenDetails{}
	td.AtExpires = time.Now().Add(time.Minute * 15).Unix()
	td.AccessUuid = uuid.NewV4().String()
	td.RtExpires = time.Now().Add(time.Hour * 24 * 7).Unix()
	td.RefreshUuid = uuid.NewV4().String()

	var err error
	_ = os.Setenv("ACCESS_SECRET", "jdnfksdmfksd") //this should be in an env file
	atClaims := jwt.MapClaims{}
	atClaims["authorized"] = true
	atClaims["access_uuid"] = td.AccessUuid
	atClaims["user_id"] = userid
	atClaims["exp"] = td.AtExpires
	at := jwt.NewWithClaims(jwt.SigningMethodHS512, atClaims)
	td.AccessToken, err = at.SignedString([]byte(os.Getenv("ACCESS_SECRET")))
	if err != nil {
		return nil, err
	}
	_ = os.Setenv("REFRESH_SECRET", "mcmvmkmsdnfsdmfdsjf") //this should be in an env file
	rtClaims := jwt.MapClaims{}
	rtClaims["refresh_uuid"] = td.RefreshUuid
	rtClaims["user_id"] = userid
	rtClaims["exp"] = td.RtExpires
	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims)
	td.RefreshToken, err = rt.SignedString([]byte(os.Getenv("REFRESH_SECRET")))
	if err != nil {
		return nil, err
	}
	return td, nil
}

func (s *server) error(w http.ResponseWriter, r *http.Request, code int, err error) {
	s.respond(w, r, code, map[string]string{"error": err.Error()})
}

func (s *server) respond(w http.ResponseWriter, r *http.Request, code int, data interface{}) {
	w.WriteHeader(code)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}
