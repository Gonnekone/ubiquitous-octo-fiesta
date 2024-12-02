package tests

import (
	"github.com/Gonnekone/ubiquitous-octo-fiesta/internal/http-server/handlers/refresh"
	"github.com/Gonnekone/ubiquitous-octo-fiesta/internal/lib/api/response"
	"github.com/gavv/httpexpect/v2"
	"github.com/google/uuid"
	"net/http"
	"net/url"
	"testing"
	"time"
)

const host = "localhost:8082"

func TestGetRefresh(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		u := url.URL{
			Scheme: "http",
			Host:   host,
		}

		e := httpexpect.Default(t, u.String())

		// get tokens

		guid := uuid.New().String()

		resp := e.GET("/get-tokens").
			WithQuery("guid", guid).
			Expect().
			Status(http.StatusOK).
			JSON().Object()

		accessToken := resp.Value("access_token").String().Raw()
		refreshToken := resp.Value("refresh_token").String().Raw()

		// refresh tokens

		e.POST("/refresh").
			WithJSON(refresh.Request{
				AccessToken:  accessToken,
				RefreshToken: refreshToken,
			}).
			Expect().
			Status(http.StatusOK).
			JSON().Object().
			ContainsKey("access_token").
			ContainsKey("refresh_token")
	})
}

func TestSecondRefresh(t *testing.T) {
	t.Run("Second refresh with the same refresh token", func(t *testing.T) {
		u := url.URL{
			Scheme: "http",
			Host:   host,
		}

		e := httpexpect.Default(t, u.String())

		// get tokens

		guid := uuid.New().String()

		resp := e.GET("/get-tokens").
			WithQuery("guid", guid).
			Expect().
			Status(http.StatusOK).
			JSON().Object()

		accessToken := resp.Value("access_token").String().Raw()
		refreshToken := resp.Value("refresh_token").String().Raw()

		// first refresh

		time.Sleep(time.Second)

		e.POST("/refresh").
			WithJSON(refresh.Request{
				AccessToken:  accessToken,
				RefreshToken: refreshToken,
			}).
			Expect().
			Status(http.StatusOK).
			JSON().Object().
			ContainsKey("access_token").
			ContainsKey("refresh_token")

		// second refresh

		time.Sleep(time.Second)

		e.POST("/refresh").
			WithJSON(refresh.Request{
				AccessToken:  accessToken,
				RefreshToken: refreshToken,
			}).
			Expect().
			Status(http.StatusBadRequest).
			JSON().Object().
			IsEqual(response.Error("invalid request"))
	})
}

// ?????
//func Test(t *testing.T) {
//	u := url.URL{
//		Scheme: "http",
//		Host:   host,
//	}
//
//	e := httpexpect.Default(t, u.String())
//	var accessToken, refreshToken string
//
//	t.Run("get tokens", func(t *testing.T) {
//		guid := uuid.New().String()
//
//		resp := e.GET("/get-tokens").
//			WithQuery("guid", guid).
//			Expect().
//			Status(http.StatusOK).
//			JSON().Object()
//
//		accessToken = resp.Value("access_token").String().Raw()
//		refreshToken = resp.Value("refresh_token").String().Raw()
//	})
//
//	t.Run("refresh tokens", func(t *testing.T) {
//		e.POST("/refresh").
//			WithJSON(refresh.Request{
//				AccessToken:  accessToken,
//				RefreshToken: refreshToken,
//			}).
//			Expect().
//			Status(http.StatusOK).
//			JSON().Object().
//			ContainsKey("access_token").
//			ContainsKey("refresh_token")
//	})
//
//	t.Run("second refresh tokens", func(t *testing.T) {
//		e.POST("/refresh").
//			WithJSON(refresh.Request{
//				AccessToken:  accessToken,
//				RefreshToken: refreshToken,
//			}).
//			Expect().
//			Status(http.StatusBadRequest).
//			JSON().Object().
//			IsEqual(response.Error("invalid request"))
//	})
//}
