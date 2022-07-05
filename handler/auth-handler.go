package handler

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gopkg.in/boj/redistore.v1"
	"gorm.io/gorm"
	"io/ioutil"
	"log"
	"net/http"
	common "test-vehcile-monitoring/common/config"
	"test-vehcile-monitoring/session/sessionstore"
	"time"
)

type AuthServiceHandler struct {
	Logger              *logrus.Entry
	Config              *common.Config
	OauthConfig         *oauth2.Config
	ServiceDB           *gorm.DB
	SessionStore        *redistore.RediStore
	ServiceSessionStore *sessionstore.ServiceSessionStore
}

/*
 AuthCodeURL로 유저를 어떤 경로로 보내야 하는지 지정 (구글 로그인 경로로 보내야 함-> googleConfig)
 http.Redirect로 요청, 응답, 주소, 상태코드(보내는 이유)

AuthCodeURL에 state 코드를 넣어서 보내야 함
state 코드란 : CSRF 공격을 막기 위한 토큰 (URL 변조 해킹), 일회용 키 역할  --> 브라우저 쿠키에 임시 키를 심음
state 객체에 cookie 생성 function을 넣어 전달

이후 callback에 담겨오는 state 객체 값이랑 아래 쿠키에 저장한 state 값을 비교해 일치하면 인증 성공
*/
func (h *AuthServiceHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	state := h.generateStateOauthCookie(w)
	h.OauthConfig = &oauth2.Config{
		ClientID:     h.Config.GoogleOAuth2.ClientID,
		ClientSecret: h.Config.GoogleOAuth2.ClientSecret,
		RedirectURL:  h.Config.GoogleOAuth2.CallbackURL,
		Scopes:       []string{h.Config.GoogleOAuth2.ScopeEmail, h.Config.GoogleOAuth2.ScopeProfile},
		Endpoint:     google.Endpoint,
	}

	url := h.OauthConfig.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (h *AuthServiceHandler) GoogleAuthCallback(w http.ResponseWriter, r *http.Request) {
	oauthstate, _ := r.Cookie("oauthstate") // -- 1

	if r.FormValue("state") != oauthstate.Value { // -- 2
		log.Printf("invalid google oauth state cookie : %s state : %s\n", oauthstate.Value, r.FormValue("state"))
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	data, err := h.getGoogleUserInfo(r.FormValue("code")) // -- 3
	if err != nil {
		log.Println(err.Error())
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	} // -- 3-1

	fmt.Fprint(w, string(data)) // -- 3-2
}

//cookie에 일회용 비밀번호 저장
//쿠키 만료 시간 : 현재로부터 24시간
//16byte 짜리 배열을 랜덤하게 채우고 bytes를 string으로 인코딩 -> 이 값을 state 객체로 저장
//http header에 setcookie 설정
func (h *AuthServiceHandler) generateStateOauthCookie(w http.ResponseWriter) string {
	expiration := time.Now().Add(1 * 24 * time.Hour)

	bytes := make([]byte, 16)
	rand.Read(bytes)
	state := base64.URLEncoding.EncodeToString(bytes)

	cookie := &http.Cookie{
		Name:    "oauthstate",
		Value:   state,
		Expires: expiration,
	}
	http.SetCookie(w, cookie)
	return state
}

//구글에서 유저정보 가져오기
func (h *AuthServiceHandler) getGoogleUserInfo(code string) ([]byte, error) {
	ctx := context.Background()
	token, err := h.OauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("Failed to Exchange %s\n", err.Error())
	}

	resp, err := http.Get(h.Config.GoogleOAuth2.OathUrlAPI + token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("Failed to get userInfo %s\n", err.Error())
	}
	return ioutil.ReadAll(resp.Body)

}
