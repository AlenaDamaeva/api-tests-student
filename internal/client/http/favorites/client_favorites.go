package favorites

import (
	"net/http"
	"testing"

	"api-tests-template/internal/constants/path"
	apiRunner "api-tests-template/internal/helpers/api-runner"
)

// HttpPostFavorite добавляет объявление в избранное (отправляем JSON с advertisement_id)
func HttpPostFavorite(t *testing.T, token string, adID string) *http.Response {
	body := map[string]string{"advertisement_id": adID}
	return apiRunner.GetRunner().Auth(token).Create().Post(path.FavoritesPath).
		JSON(body).
		Expect(t).
		End().Response
}

// HttpGetFavorites получает список избранных объявлений
func HttpGetFavorites(t *testing.T, token string) *http.Response {
	return apiRunner.GetRunner().Auth(token).Create().Get(path.FavoritesPath).
		Expect(t).
		End().Response
}

// HttpDeleteFavorite удаляет объявление из избранного (ID передается в URL)
func HttpDeleteFavorite(t *testing.T, token string, adID string) *http.Response {
	return apiRunner.GetRunner().Auth(token).Create().Delete(path.FavoritesPath + "/" + adID).
		Expect(t).
		End().Response
}
