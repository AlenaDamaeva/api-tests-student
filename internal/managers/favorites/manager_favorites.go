package favorites

import (
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	"api-tests-template/internal/client/http/favorites"
)

// AddToFavorites добавляет объявление в избранное и проверяет статус-код.
func AddToFavorites(t *testing.T, token string, adID string, expectedStatusCode int) string {
	rawResponse := favorites.HttpPostFavorite(t, token, adID)
	defer rawResponse.Body.Close()

	require.Equalf(t, expectedStatusCode, rawResponse.StatusCode,
		"HTTP status code должен быть %d", expectedStatusCode)

	responseBody, err := io.ReadAll(rawResponse.Body)
	require.NoError(t, err)
	return string(responseBody)
}

// GetFavorites получает список избранного и проверяет статус-код.
func GetFavorites(t *testing.T, token string, expectedStatusCode int) string {
	rawResponse := favorites.HttpGetFavorites(t, token)
	defer rawResponse.Body.Close()

	require.Equalf(t, expectedStatusCode, rawResponse.StatusCode,
		"HTTP status code должен быть %d", expectedStatusCode)

	responseBody, err := io.ReadAll(rawResponse.Body)
	require.NoError(t, err)
	return string(responseBody)
}

// RemoveFromFavorites удаляет объявление из избранного и проверяет статус-код.
func RemoveFromFavorites(t *testing.T, token string, adID string, expectedStatusCode int) {
	_, _ = RemoveFromFavoritesRaw(t, token, adID, expectedStatusCode)
}

// RemoveFromFavoritesRaw удаляет объявление из избранного, проверяет статус-код
// и возвращает тело ответа.
func RemoveFromFavoritesRaw(t *testing.T, token string, adID string, expectedStatusCode int) (string, int) {
	rawResponse := favorites.HttpDeleteFavorite(t, token, adID)
	defer rawResponse.Body.Close()

	require.Equalf(t, expectedStatusCode, rawResponse.StatusCode,
		"HTTP status code должен быть %d", expectedStatusCode)

	responseBody, err := io.ReadAll(rawResponse.Body)
	require.NoError(t, err)
	return string(responseBody), rawResponse.StatusCode
}
