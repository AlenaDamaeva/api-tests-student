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
func HttpPostAdvertisement(
	t *testing.T,
	token string,
	title, description string,
	price, quantity int,
	photoPath string,
) *http.Response {
	return HttpPostAdvertisementRaw(
		t,
		token,
		title,
		description,
		strconv.Itoa(price),
		strconv.Itoa(quantity),
		[]string{photoPath},
	)
}

// HttpPostAdvertisementRaw отправляет POST-запрос
// с произвольными значениями цены, количества и фотографиями.
func HttpPostAdvertisementRaw(
	t *testing.T,
	token string,
	title, description string,
	price, quantity string,
	photoPaths []string,
) *http.Response {
	request := apiRunner.GetRunner().Auth(token).Create().Post(path.AdvertisementPath).
		MultipartFormData("title", title).
		MultipartFormData("description", description).
		MultipartFormData("price", price).
		MultipartFormData("quantity", quantity)

	for _, photo := range photoPaths {
		request = request.MultipartFile("photos", photo)
	}

	return request.Expect(t).End().Response
}

// HttpGetAdvertisementByID получает объявление по ID.
func HttpGetAdvertisementByID(t *testing.T, token string, id string) *http.Response {
	return apiRunner.GetRunner().Auth(token).Create().Get(path.AdvertisementPath).
		Query("id", id).
		Expect(t).
		End().Response
}

// HttpGetAdvertisementsBySearch ищет объявления по названию.
func HttpGetAdvertisementsBySearch(t *testing.T, token string, searchTitle string) *http.Response {
	return apiRunner.GetRunner().Auth(token).Create().Get(path.AdvertisementsPath).
		Query("search", searchTitle).
		Expect(t).
		End().Response
}

// HttpGetAdvertisementPhotos получает фотографии объявления по ID.
func HttpGetAdvertisementPhotos(t *testing.T, token string, id string) *http.Response {
	return apiRunner.GetRunner().Auth(token).Create().Get(fmt.Sprintf(path.AdvertisementPhotosPath, id)).
		Expect(t).
		End().Response
}

// HttpDeleteAdvertisement удаляет объявление по ID.
func HttpDeleteAdvertisement(t *testing.T, token string, id string) *http.Response {
	return apiRunner.GetRunner().Auth(token).Create().Delete(path.AdvertisementPath).
		Query("id", id).
		Expect(t).
		End().Response
}
