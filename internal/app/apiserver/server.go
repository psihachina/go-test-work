package apiserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/psihachina/go-test-work.git/internal/app/model"
	"github.com/psihachina/go-test-work.git/internal/app/store"
	"github.com/sirupsen/logrus"
	"github.com/twinj/uuid"
)

type server struct {
	router *gin.Engine
	logger *logrus.Logger
	store  store.Store
}

func newServer(store store.Store) *server {
	s := &server{
		router: gin.Default(),
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
	s.router.GET("/", s.HandleServerWork)
	s.router.POST("/Login", s.HandleSessionsCreate)
	s.router.POST("/Logout", s.HandleSessionsDelete)
	s.router.POST("/Refresh", s.HandleSessionsRefresh)
	s.router.POST("/LogoutAll", s.HandleAllSessionsDelete)
}

func (s *server) HandleServerWork(c *gin.Context) {
	c.String(200, "work")
}

func (s *server) HandleSessionsRefresh(c *gin.Context) {
	token, err := s.VerifyRefreshToken(c.Request)
	if err != nil {
		s.respond(c.Writer, c.Request, http.StatusUnauthorized, "Refresh token expired")
		return
	}
	if _, ok := token.Claims.(jwt.Claims); !ok && !token.Valid {
		s.respond(c.Writer, c.Request, http.StatusUnauthorized, err)
		return
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		refreshUUID, ok := claims["refresh_uuid"].(string)
		if !ok {
			s.respond(c.Writer, c.Request, http.StatusUnprocessableEntity, err)
			return
		}
		userID := claims["user_id"].(string)
		if err != nil {
			s.respond(c.Writer, c.Request, http.StatusUnprocessableEntity, "Error occurred")
			return
		}

		deleted, delErr := s.store.Token().DeleteAuth(refreshUUID)
		if delErr != nil || deleted == 0 { //if any goes wrong
			s.respond(c.Writer, c.Request, http.StatusUnauthorized, "unauthorized")
			return
		}

		ts, createErr := s.Create(userID)
		if createErr != nil {
			s.respond(c.Writer, c.Request, http.StatusForbidden, createErr.Error())
			return
		}

		saveErr := s.store.Token().CreateAuth(userID, ts)
		if saveErr != nil {
			s.respond(c.Writer, c.Request, http.StatusForbidden, saveErr.Error())
			return
		}
		tokens := map[string]string{
			"access_token":  ts.AccessToken,
			"refresh_token": ts.RefreshToken,
		}

		c.Writer.Header().Add("Authorization", ts.AccessToken)

		http.SetCookie(c.Writer, &http.Cookie{
			Name:     "refresh_token",
			Value:    ts.RefreshToken,
			Expires:  time.Now().Add(120 * time.Minute),
			HttpOnly: true,
		})

		s.respond(c.Writer, c.Request, http.StatusCreated, tokens)
	} else {
		s.respond(c.Writer, c.Request, http.StatusUnauthorized, "refresh expired")
	}
}

func (s *server) HandleSessionsCreate(c *gin.Context) {
	type request struct {
		UserID string `json:"id"`
	}
	req := &request{}
	if err := json.NewDecoder(c.Request.Body).Decode(req); err != nil {
		s.error(c.Writer, c.Request, http.StatusBadRequest, err)
		return
	}

	ts, err := s.Create(req.UserID)
	if err != nil {
		s.respond(c.Writer, c.Request, http.StatusUnprocessableEntity, err.Error())
		return
	}

	err = s.store.Token().CreateAuth(req.UserID, ts)
	if err != nil {
		s.respond(c.Writer, c.Request, http.StatusUnprocessableEntity, err.Error())
	}

	tokens := map[string]string{
		"access_token":  ts.AccessToken,
		"refresh_token": ts.RefreshToken,
	}

	c.Writer.Header().Set("Authorization", ts.AccessToken)

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "refresh_token",
		Value:    ts.RefreshToken,
		Expires:  time.Now().Add(120 * time.Minute),
		HttpOnly: true,
	})

	s.respond(c.Writer, c.Request, http.StatusOK, tokens)
}

func (s *server) HandleSessionsDelete(c *gin.Context) {
	metadata, err := s.ExtractTokenMetadata(c.Request)
	if err != nil {
		s.respond(c.Writer, c.Request, http.StatusUnauthorized, "unauthorized")
		return
	}

	_, delErr := s.store.Token().DeleteAuth(metadata.RefreshUUID)
	if delErr != nil {
		s.respond(c.Writer, c.Request, http.StatusUnauthorized, delErr.Error())
		return
	}
	s.respond(c.Writer, c.Request, http.StatusOK, "Successfully logged out")
}

func (s *server) HandleAllSessionsDelete(c *gin.Context) {
	metadata, err := s.ExtractTokenMetadata(c.Request)
	if err != nil {
		s.respond(c.Writer, c.Request, http.StatusUnauthorized, "unauthorized")
		return
	}

	delErr := s.store.Token().DeleteTokens(metadata)
	if delErr != nil {
		s.respond(c.Writer, c.Request, http.StatusUnauthorized, delErr.Error())
		return
	}
	s.respond(c.Writer, c.Request, http.StatusOK, "Successfully logged out")
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
		accessUUID, ok := claims["access_uuid"].(string)
		if !ok {
			return nil, err
		}
		userID := claims["user_id"].(string)
		if err != nil {
			return nil, err
		}

		fmt.Println(accessUUID, userID)

		claims, ok := refreshToken.Claims.(jwt.MapClaims)
		if ok && refreshToken.Valid {
			refreshUUID, ok := claims["refresh_uuid"].(string)
			if !ok {
				return nil, err
			}
			fmt.Println(refreshUUID)

			return &model.AccessDetails{
				AccessUUID:  accessUUID,
				UserID:      userID,
				RefreshUUID: refreshUUID,
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
