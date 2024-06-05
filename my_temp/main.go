package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"first-app/models"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"

	"net/http"

	"github.com/gin-gonic/gin"
)

var db *sql.DB
var cache *Cache

func main() {
	env, err := loadEnv(".env")
	if err != nil {
		fmt.Println("Không thể đọc file .env:", err)
		return
	}
	cache = NewCache()
	initDB(env)
	defer db.Close()
	fmt.Println("See you againt!")

	router := gin.Default()
	// router.MaxMultipartMemory = 8 << 20 // 8 MiB
	router.Static("/", "./public")
	v1 := router.Group("/v1")
	{
		v1.GET("/page/:name", readTemplateHtml) // get page by name
	}

	v2 := router.Group("/v2")
	{
		v2.GET("/tet")
	}

	if err := router.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}

func combineHTMLWithData(html string, data string) (string, error) {
	if data != "" {
		var jsonData map[string]interface{}
		err := json.Unmarshal([]byte(data), &jsonData)
		if err != nil {
			return "", err
		}

		for key, value := range jsonData {
			placeholder := "{{ ." + key + " }}"
			html = replacePlaceholder(html, placeholder, fmt.Sprintf("%v", value))
		}
	}

	return html, nil
}

func replacePlaceholder(html string, placeholder string, value string) string {
	return strings.ReplaceAll(html, placeholder, value)
}

func initDB(env map[string]string) {
	connectionString, ok := env["POSTGRE_CONNECTIONSTRING"]
	if !ok {
		fmt.Println("POSTGRE_CONNECTIONSTRING không được tìm thấy trong file .env")
		os.Exit(1)
	}

	// connect DB
	var err error
	db, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	}

	// config pool
	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(5)

	// check connect
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
}

func loadEnv(filename string) (map[string]string, error) {
	env := make(map[string]string)

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := parts[0]
			value := parts[1]
			env[key] = value
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return env, nil
}

func readTemplateHtml(c *gin.Context) {
	templateName := c.Param("name")

	if cachedHTML, found := cache.Get(templateName); found {
		// Nếu đã tồn tại trong cache, trả về HTML từ cache
		c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		c.Writer.WriteHeader(http.StatusOK)
		fmt.Fprintf(c.Writer, cachedHTML.(string))
		return
	}

	var template models.Templates
	if err := db.QueryRow("SELECT id, name, html, data FROM templates WHERE name = $1 AND status = 'active'", templateName).
		Scan(&template.ID, &template.Name, &template.HTML, &template.Data); err != nil {
		log.Printf("Error fetching template: %s\n", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
		return
	}

	combinedHTML, err := combineHTMLWithData(template.HTML, template.Data.String)
	if err != nil {
		log.Printf("Error combining HTML with data: %s\n", err)
		c.JSON(500, gin.H{"error": "Failed to combine HTML with data"})
		return
	}

	cache.Set(templateName, combinedHTML)

	c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	c.Writer.WriteHeader(http.StatusOK)
	fmt.Fprintf(c.Writer, combinedHTML)
}

type Cache struct {
	cache map[string]interface{}
	mutex sync.RWMutex
}

func NewCache() *Cache {
	return &Cache{
		cache: make(map[string]interface{}),
	}
}

func (c *Cache) Set(key string, value interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.cache[key] = value
}

func (c *Cache) Get(key string) (interface{}, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	value, ok := c.cache[key]
	return value, ok
}
