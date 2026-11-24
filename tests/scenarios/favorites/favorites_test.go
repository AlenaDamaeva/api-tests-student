package favorites

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/tidwall/gjson"

	"api-tests-template/internal/managers/advertisements"
	"api-tests-template/internal/managers/auth"
	"api-tests-template/internal/managers/auth/models"
	"api-tests-template/internal/managers/favorites"

	base "api-tests-template/tests"
)

const (
	testPhotoPath    = "test_photo.jpg"
	nonExistentID    = "11111111-1111-1111-1111-111111111111"
	invalidTokenText = "invalid_token_12345"
)

type TestSuite struct {
	suite.Suite
	loginData1 models.LoginOkResponse // Создатель объявления
	loginData2 models.LoginOkResponse // Пользователь, работающий с избранным
}

func TestSuiteRun(t *testing.T) {
	suite.Run(t, &TestSuite{})
}

func (s *TestSuite) SetupSuite() {
	base.SetupSuite()
	s.loginData1 = auth.Login(s.T(), os.Getenv("TEST_LOGIN"), os.Getenv("TEST_PASSWORD"))
	s.loginData2 = auth.Login(s.T(), "i.damaeva@yandex.ru", "rEK-Wie-B2E-jG5")
}

func (s *TestSuite) TearDownSuite() {
	base.TearDownSuite()
}

func checkErrorResponse(t *testing.T, body string) {
	require.NotEmpty(t, body, "Тело ответа с ошибкой пустое")

	errorText := gjson.Get(body, "error").String()
	messageText := gjson.Get(body, "message").String()

	require.NotEmpty(t, errorText, "В ответе отсутствует поле error (тело: %s)", body)
	require.NotEmpty(t, messageText, "В ответе отсутствует поле message (тело: %s)", body)
}

func containsAdID(favList string, adID string) bool {
	if adID == "" || favList == "" {
		return false
	}

	var elements []gjson.Result
	if items := gjson.Get(favList, "items"); items.IsArray() {
		elements = items.Array()
	} else if parsed := gjson.Parse(favList); parsed.IsArray() {
		elements = parsed.Array()
	}

	for _, item := range elements {
		if item.Get("id").String() == adID ||
			item.Get("advertisement_id").String() == adID ||
			item.Get("advertisement.id").String() == adID {
			return true
		}
	}

	return strings.Contains(favList, adID)
}

// Positive: Полный цикл работы с избранным:
// POST /favorites -> GET /favorites -> DELETE /favorites/{ad_id} -> GET /favorites.
func (s *TestSuite) TestFavoritesCycle() {
	title := fmt.Sprintf("Тестовое объявление %d", time.Now().UnixNano())
	createResponse := advertisements.CreateAdvertisement(
		s.T(), s.loginData1.Token,
		title, "Описание", 1000, 1, testPhotoPath,
		http.StatusCreated,
	)
	adID := gjson.Get(createResponse, "id").String()
	require.NotEmpty(s.T(), adID, "ID объявления пустой")

	s.T().Cleanup(func() {
		if adID != "" {
			advertisements.DeleteAdvertisement(s.T(), s.loginData1.Token, adID, http.StatusNoContent)
		}
	})

	s.Run("Добавляем объявление в избранное", func() {
		favorites.AddToFavorites(s.T(), s.loginData2.Token, adID, http.StatusCreated)
	})

	s.Run("Проверяем, что объявление появилось в GET /favorites", func() {
		favList := favorites.GetFavorites(s.T(), s.loginData2.Token, http.StatusOK)
		require.True(s.T(), containsAdID(favList, adID), "Объявление %s не найдено в избранном: %s", adID, favList)
	})

	s.Run("Удаляем объявление из избранного", func() {
		favorites.RemoveFromFavorites(s.T(), s.loginData2.Token, adID, http.StatusNoContent)
	})

	s.Run("Проверяем, что объявление пропало из GET /favorites", func() {
		favList := favorites.GetFavorites(s.T(), s.loginData2.Token, http.StatusOK)
		require.False(s.T(), containsAdID(favList, adID), "Объявление %s всё ещё в избранном: %s", adID, favList)
	})
}

// Positive: Удаление самого объявления убирает его и из избранного пользователя.
func (s *TestSuite) TestDeleteAdvertisementRemovesFromFavorites() {
	title := fmt.Sprintf("Объявление для каскада %d", time.Now().UnixNano())
	createResponse := advertisements.CreateAdvertisement(
		s.T(), s.loginData1.Token,
		title, "Описание", 1000, 1, testPhotoPath,
		http.StatusCreated,
	)
	adID := gjson.Get(createResponse, "id").String()
	require.NotEmpty(s.T(), adID)

	favorites.AddToFavorites(s.T(), s.loginData2.Token, adID, http.StatusCreated)

	favBeforeDelete := favorites.GetFavorites(s.T(), s.loginData2.Token, http.StatusOK)
	require.True(s.T(), containsAdID(favBeforeDelete, adID), "Объявление должно быть в избранном до удаления")

	// Удаляем само объявление
	advertisements.DeleteAdvertisement(s.T(), s.loginData1.Token, adID, http.StatusNoContent)

	// Проверяем, что оно исчезло и из избранного
	favAfterDelete := favorites.GetFavorites(s.T(), s.loginData2.Token, http.StatusOK)
	require.False(s.T(), containsAdID(favAfterDelete, adID), "Объявление должно исчезнуть из избранного после DELETE объявления")
}

// Negative: Добавление несуществующего advertisement_id в избранное.
func (s *TestSuite) TestAddNonExistentFavorite() {
	body := favorites.AddToFavorites(s.T(), s.loginData2.Token, nonExistentID, http.StatusNotFound)
	checkErrorResponse(s.T(), body)
}

// Negative: Удаление несуществующего advertisement_id из избранного.
func (s *TestSuite) TestRemoveNonExistentFavorite() {
	body, _ := favorites.RemoveFromFavoritesRaw(s.T(), s.loginData2.Token, nonExistentID, http.StatusNotFound)
	checkErrorResponse(s.T(), body)
}

// Negative: 401 Unauthorized для POST /favorites (без токена и с неверным токеном).
func (s *TestSuite) TestPostFavoritesUnauthorized() {
	bodyNoToken := favorites.AddToFavorites(s.T(), "", nonExistentID, http.StatusUnauthorized)
	checkErrorResponse(s.T(), bodyNoToken)

	bodyInvalidToken := favorites.AddToFavorites(s.T(), invalidTokenText, nonExistentID, http.StatusUnauthorized)
	checkErrorResponse(s.T(), bodyInvalidToken)
}

// Negative: 401 Unauthorized для GET /favorites (без токена и с неверным токеном).
func (s *TestSuite) TestGetFavoritesUnauthorized() {
	bodyNoToken := favorites.GetFavorites(s.T(), "", http.StatusUnauthorized)
	checkErrorResponse(s.T(), bodyNoToken)

	bodyInvalidToken := favorites.GetFavorites(s.T(), invalidTokenText, http.StatusUnauthorized)
	checkErrorResponse(s.T(), bodyInvalidToken)
}

// Negative: 401 Unauthorized для DELETE /favorites/{ad_id} (без токена и с неверным токеном).
func (s *TestSuite) TestDeleteFavoritesUnauthorized() {
	bodyNoToken, _ := favorites.RemoveFromFavoritesRaw(s.T(), "", nonExistentID, http.StatusUnauthorized)
	checkErrorResponse(s.T(), bodyNoToken)

	bodyInvalidToken, _ := favorites.RemoveFromFavoritesRaw(s.T(), invalidTokenText, nonExistentID, http.StatusUnauthorized)
	checkErrorResponse(s.T(), bodyInvalidToken)
}

// Negative: Повторное добавление того же объявления в избранное.
func (s *TestSuite) TestAddDuplicateFavorite() {
	title := fmt.Sprintf("Тестовое объявление %d", time.Now().UnixNano())
	createResponse := advertisements.CreateAdvertisement(
		s.T(), s.loginData1.Token,
		title, "Описание", 1000, 1, testPhotoPath,
		http.StatusCreated,
	)
	adID := gjson.Get(createResponse, "id").String()
	require.NotEmpty(s.T(), adID)

	s.T().Cleanup(func() {
		if adID != "" {
			advertisements.DeleteAdvertisement(s.T(), s.loginData1.Token, adID, http.StatusNoContent)
		}
	})

	favorites.AddToFavorites(s.T(), s.loginData2.Token, adID, http.StatusCreated)

	dupBody := favorites.AddToFavorites(s.T(), s.loginData2.Token, adID, http.StatusConflict)
	checkErrorResponse(s.T(), dupBody)
}

// Negative: Попытка добавить в избранное свое собственное объявление.
func (s *TestSuite) TestAddOwnAdvertisementToFavorites() {
	title := fmt.Sprintf("Свое объявление %d", time.Now().UnixNano())
	createResponse := advertisements.CreateAdvertisement(
		s.T(), s.loginData2.Token,
		title, "Описание", 1000, 1, testPhotoPath,
		http.StatusCreated,
	)
	adID := gjson.Get(createResponse, "id").String()
	require.NotEmpty(s.T(), adID)

	s.T().Cleanup(func() {
		if adID != "" {
			advertisements.DeleteAdvertisement(s.T(), s.loginData2.Token, adID, http.StatusNoContent)
		}
	})

	ownBody := favorites.AddToFavorites(s.T(), s.loginData2.Token, adID, http.StatusForbidden)
	checkErrorResponse(s.T(), ownBody)
}
