package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"first-app/models"
	"log"
	"strconv"

	"fmt"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"

	"net/http"

	"github.com/gin-gonic/gin"
)

var db *sql.DB

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

type ToDoItem struct {
	gorm.Model
	// Id        int        `json:"id" gorm:"column:id;"`
	Title  string `json:"title" gorm:"column:title;"`
	Status string `json:"status" gorm:"column:status;"`
	// CreatedAt *time.Time `json:"created_at" gorm:"column:created_at;"`
	// UpdatedAt *time.Time `json:"updated_at" gorm:"column:updated_at;"`
}

func (ToDoItem) TableName() string { return "todo_items" }

// AddItemRequest chứa thông tin được yêu cầu khi thêm một mục mới
// swagger:parameters addItem
type AddItemRequest struct {
	// Tiêu đề của mục
	// required: true
	// example: Mua sữa
	Title string `json:"title"`

	// Trạng thái của mục
	// required: true
	// enum: Doing,Done
	// example: Doing
	Status string `json:"status"`
}

// AddItemResponse chứa thông tin về kết quả khi thêm một mục mới
// swagger:response addItemResponse
type AddItemResponse struct {
	// ID của mục mới được thêm
	// example: 1
	ID int `json:"id"`
}

// GetItemResponse chứa thông tin về một mục được trả về từ API
// swagger:response getItemResponse
type GetItemResponse struct {
	// ID của mục
	// example: 1
	ID int `json:"id"`

	// Tiêu đề của mục
	// example: Mua sữa
	Title string `json:"title"`

	// Trạng thái của mục
	// example: Doing
	Status string `json:"status"`
}

// EditItemRequest chứa thông tin được yêu cầu khi chỉnh sửa một mục
// swagger:parameters editItem
type EditItemRequest struct {
	// Tiêu đề của mục
	// required: true
	// example: Mua sữa
	Title string `json:"title"`

	// Trạng thái của mục
	// required: true
	// enum: Doing,Done
	// example: Doing
	Status string `json:"status"`
}

// DeleteItemResponse chứa thông tin về kết quả khi xóa một mục
// swagger:response deleteItemResponse
type DeleteItemResponse struct {
	// Kết quả xóa mục
	// example: true
	Success bool `json:"success"`
}

func createItem(c *gin.Context) {
	var dataItem ToDoItem

	if err := c.ShouldBind(&dataItem); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// preprocess title - trim all spaces
	dataItem.Title = strings.TrimSpace(dataItem.Title)

	if dataItem.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title cannot be blank"})
		return
	}

	// do not allow "finished" status when creating a new task
	dataItem.Status = "Doing" // set to default

	// Thực hiện các thao tác với cơ sở dữ liệu ở đây
	stmt, err := db.Prepare("INSERT INTO todo_items (title, status) VALUES ($1, $2) RETURNING id")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer stmt.Close()

	err = stmt.QueryRow(dataItem.Title, dataItem.Status).Scan(&dataItem.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": dataItem.ID})
}

// func createItem(db *gorm.DB) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		var dataItem ToDoItem

// 		if err := c.ShouldBind(&dataItem); err != nil {
// 			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 			return
// 		}

// 		// preprocess title - trim all spaces
// 		dataItem.Title = strings.TrimSpace(dataItem.Title)

// 		if dataItem.Title == "" {
// 			c.JSON(http.StatusBadRequest, gin.H{"error": "title cannot be blank"})
// 			return
// 		}

// 		// do not allow "finished" status when creating a new task
// 		dataItem.Status = "Doing" // set to default

// 		if err := db.Create(&dataItem).Error; err != nil {
// 			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 			return
// 		}

// 		c.JSON(http.StatusOK, gin.H{"data": dataItem.ID})
// 	}
// }

func readItemById(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var dataItem ToDoItem

		id, err := strconv.Atoi(c.Param("id"))

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := db.Where("id = ?", id).First(&dataItem).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": dataItem})
	}
}

func getListOfItems(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		type DataPaging struct {
			Page  int   `json:"page" form:"page"`
			Limit int   `json:"limit" form:"limit"`
			Total int64 `json:"total" form:"-"`
		}

		var paging DataPaging

		if err := c.ShouldBind(&paging); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if paging.Page <= 0 {
			paging.Page = 1
		}

		if paging.Limit <= 0 {
			paging.Limit = 10
		}

		offset := (paging.Page - 1) * paging.Limit

		var result []ToDoItem

		if err := db.Table(ToDoItem{}.TableName()).
			Count(&paging.Total).
			Offset(offset).
			Order("id desc").
			Find(&result).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": result})
	}
}

func editItemById(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var dataItem ToDoItem

		if err := c.ShouldBind(&dataItem); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := db.Where("id = ?", id).Updates(&dataItem).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": true})
	}
}

func deleteItemById(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := db.Table(ToDoItem{}.TableName()).
			Where("id = ?", id).
			Delete(nil).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": true})
	}
}

func readTemplateHtml(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// id, err := strconv.Atoi(c.Param("id"))
		templateName := c.Param("name")
		log.Printf("templateName: %s\n", templateName)

		var template models.Templates
		// if err := db.Where("Name = ? AND Status = 'active'", templateName).First(&template).Error; err != nil {
		if err := db.Where("name = ?", templateName).First(&template).Error; err != nil {
			log.Printf("Error fetching template: %s\n", err)
			c.JSON(404, gin.H{"error": "Template not found"})
			return
		}

		combinedHTML, err := combineHTMLWithData(template.HTML, template.Data.String)
		if err != nil {
			log.Printf("Error combining HTML with data: %s\n", err)
			c.JSON(500, gin.H{"error": "Failed to combine HTML with data"})
			return
		}

		c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		c.Writer.WriteHeader(http.StatusOK)
		fmt.Fprintf(c.Writer, combinedHTML)
	}
}

func main() {
	// env, err := loadEnv(".env")
	// if err != nil {
	// 	fmt.Println("Không thể đọc file .env:", err)
	// 	return
	// }

	// connectionString, ok := env["POSTGRE_CONNECTIONSTRING"]
	// if !ok {
	// 	fmt.Println("POSTGRE_CONNECTIONSTRING không được tìm thấy trong file .env")
	// 	return
	// }

	// // connect db PostgreSQL
	// db, err := gorm.Open("postgres", connectionString)
	// if err != nil {
	// 	panic("Không thể kết nối đến cơ sở dữ liệu")
	// }
	// defer db.Close()

	// create table from entity
	// db.AutoMigrate(&models.User{}, ToDoItem{}, models.Templates{})

	initDB()
	defer db.Close()
	fmt.Println("See you againt!")

	router := gin.Default()

	router.LoadHTMLFiles("htmltemp/template.html")

	// Đường dẫn đến file JSON Swagger
	//go:generate swagger generate spec -o ./docs/swagger.json
	// router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// v1 := router.Group("/v1")
	// {
	// 	v1.POST("/items", createItem(db))           // create item
	// 	v1.GET("/items", getListOfItems(db))        // list items
	// 	v1.GET("/items/:id", readItemById(db))      // get an item by ID
	// 	v1.PUT("/items/:id", editItemById(db))      // edit an item by ID
	// 	v1.DELETE("/items/:id", deleteItemById(db)) // delete an item by ID
	// 	v1.GET("/temp", func(c *gin.Context) {
	// 		c.HTML(http.StatusOK, "template.html", gin.H{
	// 			"Title":   "title",
	// 			"Message": "mess",
	// 		})
	// 	})
	// 	v1.GET("/template/:name", readTemplateHtml(db)) // get page by name
	// 	// render html by templates config
	// 	// v1.GET("/template/:name", func(c *gin.Context) {
	// 	// 	name := c.Param("name")

	// 	// 	// connect db MySQL
	// 	// 	connectionMysqlString, ok := env["MYSQL_CONNECTIONSTRING"]
	// 	// 	if !ok {
	// 	// 		fmt.Println("MYSQL_CONNECTIONSTRING không được tìm thấy trong file .env")
	// 	// 		return
	// 	// 	}
	// 	// 	db, err := sql.Open("mysql", connectionMysqlString) //"root:mydb@tcp(127.0.0.1:3306)/mydb")
	// 	// 	if err != nil {
	// 	// 		log.Fatal(err)
	// 	// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "connect db error"})
	// 	// 		return
	// 	// 	}
	// 	// 	defer db.Close()

	// 	// 	var t models.Templates
	// 	// 	err = db.QueryRow("SELECT id, html, data FROM templates WHERE name = ? AND status = 'active'", name).Scan(&t.ID, &t.HTML, &t.Data)
	// 	// 	if err != nil {
	// 	// 		if err == sql.ErrNoRows {
	// 	// 			c.JSON(http.StatusNotFound, gin.H{"error": "template not found"})
	// 	// 			return
	// 	// 		}
	// 	// 		log.Fatal(err)
	// 	// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	// 	// 		return
	// 	// 	}

	// 	// 	renderedHTML, err := combineHTMLWithData(t.HTML, t.Data.String)

	// 	// 	if err != nil {
	// 	// 		log.Fatal(err)
	// 	// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	// 	// 		return
	// 	// 	}

	// 	// 	c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	// 	// 	c.Writer.WriteHeader(http.StatusOK)
	// 	// 	fmt.Fprintf(c.Writer, renderedHTML)
	// 	// })
	// }

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

// Hàm này thay thế các placeholder trong HTML với dữ liệu tương ứng
func replacePlaceholder(html string, placeholder string, value string) string {
	return strings.ReplaceAll(html, placeholder, value)
}

func initDB() {
	connectionString, ok := os.LookupEnv("POSTGRE_CONNECTIONSTRING")
	if !ok {
		fmt.Println("POSTGRE_CONNECTIONSTRING không được tìm thấy trong file .env")
		os.Exit(1)
	}

	// Mở kết nối đến cơ sở dữ liệu
	var err error
	db, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	}

	// Thiết lập số lượng kết nối tối đa trong pool
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	// Kiểm tra kết nối
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	db.AutoMigrate(&models.User{}, ToDoItem{}, models.Templates{})
}
