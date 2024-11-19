package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var googleOAuthConfig = &oauth2.Config{
	RedirectURL: "http://localhost:8000/auth/google/callback",
	ClientID: os.Getenv("GOOGLE_CLIENT_ID"),
	ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRECT"),
	Scopes: []string{"https://www.googleapis.com/auth/userinfor.email"},
	Endpoint: google.Endpoint,
}


const oAuthGoogleUrlAPI = "https://googleapis.com/oauth2/v2/userinfor?access_token="

func oAuthGoogleLogin(w http.ResponseWriter, r *http.Request){

	oauthState := generateStateOauthCookie(w)

	u := googleOAuthConfig.AuthCodeURL(oauthState)
	http.Redirect(w, r, u, http.StatusTemporaryRedirect)

}

func oauthGoogleCallback(w http.ResponseWriter, r *http.Request){
	oauthState, _ := r.Cookie("oauthstate")

	if r.FormValue("state") != oauthState.Value{
		log.Println("invalid oauth google state")
		http.Redirect(w, r , "/", http.StatusTemporaryRedirect)
		return
	}

	data, err := getUserDataFromGoogle(r.FormValue("code"))
	if err != nil{
		log.Println(err.Error())
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	fmt.Fprintf(w, "userInfor: %s\n", data)
}


func generateStateOauthCookie(w http.ResponseWriter) string {
	var expiration = time.Now().Add(20 * time.Minute)

	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)
	cookie := http.Cookie{Name: "oauthstate", Value: state, Expires: expiration}
	http.SetCookie(w, &cookie)

	return state
}


func getUserDataFromGoogle(code string)([]byte, error){
	token, err := googleOAuthConfig.Exchange(context.Background(), code)
	if err != nil{
		return nil, fmt.Errorf("code exchange wrong: %s", err.Error())
	}
	response, err := http.Get(oAuthGoogleUrlAPI + token.AccessToken)
	if err != nil{
		return nil, fmt.Errorf("failed getting user infor: %s", err.Error())
	}
	
	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil{
		return nil, fmt.Errorf("failed to read response")
	}

	return contents, nil
}

