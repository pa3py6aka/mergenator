package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type post struct {
	Id       int    `json:"id"`
	User     string `json:"user"`
	Likes    int    `json:"likes"`
	Comments int    `json:"comments"`
	Image    string `json:"image"`
	Recipe   string `json:"recipe"`
}

var posts = []post{
	{
		Id:       1,
		User:     "Ашхабад Булбас Иванович",
		Likes:    245,
		Comments: 38,
		Image:    "https://avatars.mds.yandex.net/get-altay/10267222/2a000001929ab7ebe8ea196691001ed716f7/XXXL",
		Recipe:   "Секретный маринад из граната",
	},
	{
		Id:       2,
		User:     "Антоха Дураков",
		Likes:    76,
		Comments: 4,
		Image:    "https://avatars.mds.yandex.net/get-altay/14067398/2a000001929ab7e3a1e6b9fa0d14d4acb455/XXXL",
		Recipe:   "Афигенный шашлык из собаки",
	},
	{
		Id:       3,
		User:     "Кхуям Асламбеков",
		Likes:    301,
		Comments: 20,
		Image:    "https://avatars.mds.yandex.net/get-altay/9686455/2a0000018a4892e137bca9e41256a79d2c7a/XXL_height",
		Recipe:   "Лучший рецепт шашлыка на кухне",
	},
	{
		Id:       4,
		User:     "Шату Пердуани",
		Likes:    1,
		Comments: 4309,
		Image:    "https://masterpiecer-images.s3.yandex.net/06fdac1a96b611eeaa922ab2a9c6ab46:upscaled",
		Recipe:   "Просто маринад из уха кота",
	},
	{
		Id:       5,
		User:     "Антон Петрович",
		Likes:    4,
		Comments: 0,
		Image:    "https://img.theepochtimes.com/assets/uploads/2016/05/26/shutterstock_278631779-1200x800.jpg",
		Recipe:   "Шашлык с изюмом",
	},
	{
		Id:       6,
		User:     "Антон Петрович",
		Likes:    101,
		Comments: 48,
		Image:    "https://avatars.mds.yandex.net/i?id=6dd6cc33dc9f5335e8369b969cde6eda_l-4119069-images-thumbs&n=13",
		Recipe:   "Мясной рулет на сале",
	},
	{
		Id:       7,
		User:     "Артак Люмэн",
		Likes:    69,
		Comments: 96,
		Image:    "https://avatars.mds.yandex.net/i?id=c9fa46d02c99b4bbecbed677f7a83fb3_l-5251360-images-thumbs&n=13",
		Recipe:   "Кувыркающийся язык в кляре из масла",
	},
}

func getPosts(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, posts)
}

func getMainPage(c *gin.Context) {
	c.HTML(http.StatusOK, "main.html", gin.H{
		"Posts": posts, // передаём список постов в шаблон
	})
}

func getMergenatorPage(c *gin.Context) {
	c.HTML(http.StatusOK, "mergenator.html", gin.H{})
}

func main() {
	router := gin.Default()
	// Указываем, где искать HTML‑шаблоны
	router.LoadHTMLGlob("templates/*.html")

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"PUT", "PATCH", "GET", "POST"},
		AllowHeaders:     []string{"Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return origin == "https://localhost"
		},
		MaxAge: 12 * time.Hour,
	}))

	router.GET("/posts", getPosts)
	router.GET("/", getMainPage)
	router.GET("/mergenator", getMergenatorPage)

	err := router.Run("localhost:8080")
	if err != nil {
		return
	}
}
