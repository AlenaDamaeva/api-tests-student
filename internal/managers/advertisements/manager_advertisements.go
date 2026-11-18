package advertisements

import (
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	"api-tests-template/internal/client/http/advertisements"
)

// CreateAdvertisement создает объявление через клиент и проверяет, что вернулся ожидаемый статус-код.
// Возвращает тело ответа в виде строки.
func CreateAdvertisement(t *testing.T, token string, title, description string, price, quantity int, photoPath string, expectedStatusCode int) string {
	rawResponse := advertisements.HttpPostAdvertisement(t, token, title, description, price, quantity, photoPath)
	defer rawResponse.Body.Close()

	require.Equalf(t, expectedStatusCode, rawResponse.StatusCode,
		"HTTP status code должен быть %d", expectedStatusCode)

	responseBody, err := io.ReadAll(rawResponse.Body)
	require.NoError(t, err)
	return string(responseBody)
}

// GetAdvertisementByID получает объявление по ID и проверяет статус-код.
// Возвращает тело ответа в виде строки.
func GetAdvertisementByID(t *testing.T, token string, id string, expectedStatusCode int) string {
	rawResponse := advertisements.HttpGetAdvertisementByID(t, token, id)
	defer rawResponse.Body.Close()

	require.Equalf(t, expectedStatusCode, rawResponse.StatusCode,
		"HTTP status code должен быть %d", expectedStatusCode)

	responseBody, err := io.ReadAll(rawResponse.Body)
	require.NoError(t, err)
	return string(responseBody)
}

// SearchAdvertisements ищет объявления по названию и проверяет статус-код.
// Возвращает тело ответа в виде строки.
func SearchAdvertisements(t *testing.T, token string, searchTitle string, expectedStatusCode int) string {
	rawResponse := advertisements.HttpGetAdvertisementsBySearch(t, token, searchTitle)
	defer rawResponse.Body.Close()

	require.Equalf(t, expectedStatusCode, rawResponse.StatusCode,
		"HTTP status code должен быть %d", expectedStatusCode)

	responseBody, err := io.ReadAll(rawResponse.Body)
	require.NoError(t, err)
	return string(responseBody)
}

// GetAdvertisementPhotos получает фотографии объявления по ID и проверяет статус-код.
// Возвращает тело ответа в виде строки.
func GetAdvertisementPhotos(t *testing.T, token string, id string, expectedStatusCode int) string {
	rawResponse := advertisements.HttpGetAdvertisementPhotos(t, token, id)
	defer rawResponse.Body.Close()

	require.Equalf(t, expectedStatusCode, rawResponse.StatusCode,
		"HTTP status code должен быть %d", expectedStatusCode)

	responseBody, err := io.ReadAll(rawResponse.Body)
	require.NoError(t, err)
	return string(responseBody)
}
