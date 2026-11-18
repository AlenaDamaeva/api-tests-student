package advertisements

import (
	"fmt"
	"net/http"
	"strconv"
	"testing"

	"api-tests-template/internal/constants/path"
	apiRunner "api-tests-template/internal/helpers/api-runner"
)

// HttpPostAdvertisement отправляет POST-запрос на создание объявления.
// Передает текстовые поля и файл картинки через multipart/form-data.
func HttpPostAdvertisement(t *testing.T, token string, title, description string, price, quantity int, photoPath string) *http.Response {
	return apiRunner.GetRunner().Auth(token).Create().Post(path.AdvertisementPath).
		MultipartFormData("title", title).
		MultipartFormData("description", description).
		MultipartFormData("price", strconv.Itoa(price)).
		MultipartFormData("quantity", strconv.Itoa(quantity)).
		MultipartFile("photos", photoPath).
		Expect(t).
		End().Response
}

// HttpGetAdvertisementByID запрашивает одно объявление по его ID.
// ID передается как параметр запроса (query).
func HttpGetAdvertisementByID(t *testing.T, token string, id string) *http.Response {
	return apiRunner.GetRunner().Auth(token).Create().Get(path.AdvertisementPath).
		Query("id", id).
		Expect(t).
		End().Response
}

// HttpGetAdvertisementsBySearch ищет объявления по названию.
// Название передается в параметре search.
func HttpGetAdvertisementsBySearch(t *testing.T, token string, searchTitle string) *http.Response {
	return apiRunner.GetRunner().Auth(token).Create().Get(path.AdvertisementsPath).
		Query("search", searchTitle).
		Expect(t).
		End().Response
}

// HttpGetAdvertisementPhotos запрашивает список фотографий для объявления по ID.
func HttpGetAdvertisementPhotos(t *testing.T, token string, id string) *http.Response {
	return apiRunner.GetRunner().Auth(token).Create().Get(fmt.Sprintf(path.AdvertisementPhotosPath, id)).
		Expect(t).
		End().Response
}
