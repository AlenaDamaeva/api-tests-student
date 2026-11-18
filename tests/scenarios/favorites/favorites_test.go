package favorites

import (
	"fmt"
	"net/http"
	"os"
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
	testPhotoPath = "test_photo.jpg"
)

type TestSuite struct {
	suite.Suite
	loginData1 models.LoginOkResponse // Для создания объявления
	loginData2 models.LoginOkResponse // Для работы с избранным
}

func TestSuiteRun(t *testing.T) {
	suite.Run(t, &TestSuite{})
}

func (s *TestSuite) SetupSuite() {
	base.SetupSuite()
	// Логинимся под первым пользователем (создатель)
	s.loginData1 = auth.Login(s.T(), os.Getenv("TEST_LOGIN"), os.Getenv("TEST_PASSWORD"))
	// Логинимся под вторым пользователем (для избранного)
	s.loginData2 = auth.Login(s.T(), "i.damaeva@yandex.ru", "rEK-Wie-B2E-jG5")
}

func (s *TestSuite) TearDownSuite() {
	base.TearDownSuite()
}

// Positive: Полный цикл работы с избранным (второй пользователь добавляет чужое объявление)
func (s *TestSuite) TestFavoritesCycle() {
	// 1. Создаем объявление под первым пользователем
	title := fmt.Sprintf("Тестовое объявление %d", time.Now().Unix())
	createResponse := advertisements.CreateAdvertisement(
		s.T(), s.loginData1.Token,
		title, "Описание", 1000, 1, testPhotoPath,
		http.StatusCreated,
	)
	adID := gjson.Get(createResponse, "id").String()
	require.NotEmpty(s.T(), adID, "ID объявления пустой")

	// Очистка: удаляем объявление после теста (под первым пользователем)
	s.T().Cleanup(func() {
		if adID != "" {
			advertisements.DeleteAdvertisement(s.T(), s.loginData1.Token, adID, http.StatusNoContent)
		}
	})

	// 2. Добавляем в избранное под вторым пользователем
	s.Run("Добавляем объявление в избранное", func() {
		favorites.AddToFavorites(s.T(), s.loginData2.Token, adID, http.StatusCreated)
	})

	// 3. Проверяем, что оно в списке избранного (под вторым пользователем)
	s.Run("Проверяем, что объявление есть в избранном", func() {
		favList := favorites.GetFavorites(s.T(), s.loginData2.Token, http.StatusOK)
		require.Contains(s.T(), favList, adID, "Объявление не найдено в избранном")
	})

	// 4. Удаляем из избранного (под вторым пользователем)
	s.Run("Удаляем объявление из избранного", func() {
		favorites.RemoveFromFavorites(s.T(), s.loginData2.Token, adID, http.StatusNoContent)
	})

	// 5. Проверяем, что его больше нет (под вторым пользователем)
	s.Run("Проверяем, что объявления нет в избранном", func() {
		favList := favorites.GetFavorites(s.T(), s.loginData2.Token, http.StatusOK)
		require.NotContains(s.T(), favList, adID, "Объявление все еще в избранном")
	})

	s.T().Logf("Все проверки для цикла избранного пройдены!")
}

// Negative: Попытка добавить в избранное несуществующее объявление
func (s *TestSuite) TestAddNonExistentFavorite() {
	// Используем валидный по формату UUID, но которого нет в базе
	nonExistentID := "11111111-1111-1111-1111-111111111111"
	favorites.AddToFavorites(s.T(), s.loginData2.Token, nonExistentID, http.StatusNotFound)
}

// Negative: Попытка удалить из избранного несуществующее объявление
func (s *TestSuite) TestRemoveNonExistentFavorite() {
	nonExistentID := "11111111-1111-1111-1111-111111111111"
	favorites.RemoveFromFavorites(s.T(), s.loginData2.Token, nonExistentID, http.StatusNotFound)
}

// Negative: Запрос к избранному без авторизации
func (s *TestSuite) TestGetFavoritesWithoutToken() {
	favorites.GetFavorites(s.T(), "", http.StatusUnauthorized)
}

// Negative: Повторное добавление того же объявления в избранное
func (s *TestSuite) TestAddDuplicateFavorite() {
	// Создаем объявление под первым пользователем
	title := fmt.Sprintf("Тестовое объявление %d", time.Now().Unix())
	createResponse := advertisements.CreateAdvertisement(
		s.T(), s.loginData1.Token,
		title, "Описание", 1000, 1, testPhotoPath,
		http.StatusCreated,
	)
	adID := gjson.Get(createResponse, "id").String()
	require.NotEmpty(s.T(), adID, "ID объявления пустой")

	// Очистка
	s.T().Cleanup(func() {
		if adID != "" {
			advertisements.DeleteAdvertisement(s.T(), s.loginData1.Token, adID, http.StatusNoContent)
		}
	})

	// Добавляем первый раз (успешно)
	favorites.AddToFavorites(s.T(), s.loginData2.Token, adID, http.StatusCreated)

	// Добавляем второй раз (ожидаем 409 Conflict)
	favorites.AddToFavorites(s.T(), s.loginData2.Token, adID, http.StatusConflict)
}

// Negative: Попытка добавить в избранное свое собственное объявление
func (s *TestSuite) TestAddOwnAdvertisementToFavorites() {
	// Создаем объявление под вторым пользователем (тем же, кто будет добавлять)
	title := fmt.Sprintf("Свое объявление %d", time.Now().Unix())
	createResponse := advertisements.CreateAdvertisement(
		s.T(), s.loginData2.Token,
		title, "Описание", 1000, 1, testPhotoPath,
		http.StatusCreated,
	)
	adID := gjson.Get(createResponse, "id").String()
	require.NotEmpty(s.T(), adID, "ID объявления пустой")

	// Очистка
	s.T().Cleanup(func() {
		if adID != "" {
			advertisements.DeleteAdvertisement(s.T(), s.loginData2.Token, adID, http.StatusNoContent)
		}
	})

	// Пытаемся добавить свое же объявление в избранное (ожидаем 403 Forbidden)
	favorites.AddToFavorites(s.T(), s.loginData2.Token, adID, http.StatusForbidden)
}
