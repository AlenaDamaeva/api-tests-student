package advertisements

import (
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	"api-tests-template/internal/client/http/advertisements"
)

// CreateAdvertisement создает объявление и проверяет ожидаемый статус-код.
func CreateAdvertisement(
	t *testing.T,
	token string,
	title, description string,
	price, quantity int,
	photoPath string,
	expectedStatusCode int,
) string {
	rawResponse := advertisements.HttpPostAdvertisement(
		t, token, title, description, price, quantity, photoPath,
	)
	defer rawResponse.Body.Close()

	require.Equalf(
		t,
		expectedStatusCode,
		rawResponse.StatusCode,
		"HTTP status code должен быть %d",
		expectedStatusCode,
	)

	responseBody, err := io.ReadAll(rawResponse.Body)
	require.NoError(t, err)

	return string(responseBody)
}

// CreateAdvertisementRaw отправляет запрос и возвращает тело ответа
// и статус-код без предварительной проверки статуса.
func CreateAdvertisementRaw(
	t *testing.T,
	token string,
	title, description string,
	price, quantity string,
	photoPaths []string,
) (string, int) {
	rawResponse := advertisements.HttpPostAdvertisementRaw(
		t, token, title, description, price, quantity, photoPaths,
	)
	defer rawResponse.Body.Close()

	responseBody, err := io.ReadAll(rawResponse.Body)
	require.NoError(t, err)

	return string(responseBody), rawResponse.StatusCode
}

// GetAdvertisementByID получает объявление по ID и проверяет статус-код.
func GetAdvertisementByID(
	t *testing.T,
	token string,
	id string,
	expectedStatusCode int,
) string {
	rawResponse := advertisements.HttpGetAdvertisementByID(t, token, id)
	defer rawResponse.Body.Close()

	require.Equalf(
		t,
		expectedStatusCode,
		rawResponse.StatusCode,
		"HTTP status code должен быть %d",
		expectedStatusCode,
	)

	responseBody, err := io.ReadAll(rawResponse.Body)
	require.NoError(t, err)

	return string(responseBody)
}

// SearchAdvertisements ищет объявления по названию.
func SearchAdvertisements(
	t *testing.T,
	token string,
	searchTitle string,
	expectedStatusCode int,
) string {
	rawResponse := advertisements.HttpGetAdvertisementsBySearch(t, token, searchTitle)
	defer rawResponse.Body.Close()

	require.Equalf(
		t,
		expectedStatusCode,
		rawResponse.StatusCode,
		"HTTP status code должен быть %d",
		expectedStatusCode,
	)

	responseBody, err := io.ReadAll(rawResponse.Body)
	require.NoError(t, err)

	return string(responseBody)
}

// GetAdvertisementPhotos получает фотографии объявления по ID.
func GetAdvertisementPhotos(
	t *testing.T,
	token string,
	id string,
	expectedStatusCode int,
) string {
	rawResponse := advertisements.HttpGetAdvertisementPhotos(t, token, id)
	defer rawResponse.Body.Close()

	require.Equalf(
		t,
		expectedStatusCode,
		rawResponse.StatusCode,
		"HTTP status code должен быть %d",
		expectedStatusCode,
	)

	responseBody, err := io.ReadAll(rawResponse.Body)
	require.NoError(t, err)

	return string(responseBody)
}

// DeleteAdvertisement удаляет объявление по ID.
func DeleteAdvertisement(
	t *testing.T,
	token string,
	id string,
	expectedStatusCode int,
) {
	rawResponse := advertisements.HttpDeleteAdvertisement(t, token, id)
	defer rawResponse.Body.Close()

	require.Equalf(
		t,
		expectedStatusCode,
		rawResponse.StatusCode,
		"HTTP status code должен быть %d",
		expectedStatusCode,
	)
}
