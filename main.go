package main

import (
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type TodoItem struct {
	Id          int        `json:"id" gorm:"column:id;"`
	Title       string     `json:"title" gorm:"column:title;"`
	Description string     `json:"description" gorm:"column:description;"`
	Status      string     `json:"status" gorm:"column:status;"`
	CreatedAt   *time.Time `json:"created_at" gorm:"column:created_at;"`
	UpdatedAt   *time.Time `json:"updated_at" gorm:"column:updated_at;"`
	//UpdatedAt *time.Time `json:"updated_at,omitempty"`
	//Omiempty dùng cho trường hợp bỏ luôn biến đó nếu biến đó là kiểu int float và có giá trị là zero - kiểu string có giá trị là chuỗi rỗng ,
	//kiểu bool có giá trị là false - biến con trỏ có giá trị là null
}

func (TodoItem) TableName() string { return "todo_items" }

type TodoItemCreation struct {
	Id          int    `json:"-" gorm:"column:id;"`
	Title       string `json:"title" gorm:"column:title;"`
	Description string `json:"description" gorm:"column:description;"`
	//Status      string `json:"status" gorm:"column:status;"`
}

func (TodoItemCreation) TableName() string { return TodoItem{}.TableName() }

type TodoItemUpdate struct {
	Title       string `json:"title" gorm:"column:title;"`
	Description string `json:"description" gorm:"column:description;"`
	Status      string `json:"status" gorm:"column:status;"`
	/*
		Mặc định của gorm khi update nếu giá trị truyền lên là rỗng thì giá trị đó sẽ giữ nguyên giá trị cũ,
		Trường hợp nếu muốn truyền lên giá trị rỗng mà không bị bỏ qua thì có thể thay đổi thành con trỏ string thay vì string
		Title       *string `json:"title" gorm:"column:title;"`
		Description *string `json:"description" gorm:"column:description;"`
		Status      *string `json:"status" gorm:"column:status;"`
	*/
}

func (TodoItemUpdate) TableName() string { return TodoItem{}.TableName() }

/*Tạo 1 struct để phân trang - sử dụng query string trên url*/

type Paging struct {
	//Tham số Page thể hiện cho số trang
	Page int `json:"page" form:"page"`
	//Tham số Limit giới hạn số record cho mỗi trang
	Limit int `json:"limit" form:"limit"`
	//Tham số Total thể hiện tổng số dòng dữ liệu đáp ứng câu query
	Total int64 `json:"total" form:"-"`
}

func (p *Paging) Process() {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.Limit <= 0 || p.Limit >= 100 {
		p.Limit = 10
	}
}

func main() {
	//fmt.Println("Hello")
	dsn := os.Getenv("DB_CONN_STR")
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatalln(err)
	}

	/* Demo structure
	fmt.Println(db)

	now := time.Now().UTC()

	item := TodoItem{
		Id:          1,
		Title:       "This is item 1",
		Description: "This is item 1",
		Status:      "Doing",
		CreatedAt:   &now,
		UpdatedAt:   nil,
	}*/
	/*Từ Structure sang json string*/
	//Hàm Marshal có thể truyền giá trị là bất cứ kiểu dữ liệu nào - sau đó trả về 1 là data - 2 là lỗi
	//jsonData, err := json.Marshal(item)
	//// Nếu trả về lỗi thì in lỗi ra
	//if err != nil {
	//	fmt.Println(err)
	//	// Sau đó return để kết thúc hàm
	//	return
	//}
	//// Nếu trả về data thì in ra - jsonData là 1 mảng byte nên phải ép về kiểu string để in ra
	//fmt.Println(string(jsonData))
	//
	///*Từ json string về Structure*/
	////Ví dụ về hàm Unmarshall - lấy kiểu json trả về mảng byte
	//jsonStr := "{\"id\":1,\"title\":\"This is item 1\",\"description\":\"This is item 1\",\"status\":\"Doing\",\"createdAt\":\"2023-09-02T09:03:53.9963182Z\",\"updatedAt\":null}"
	//var item2 TodoItem
	////Unmarshall cần 2 tham số : 1. Là mảng byte của json đó là gì - biến jsonStr là kiểu string nên ép về kiểu byte
	////							 2. truyền con trỏ của Structure mà chúng ta muốn parse - ở đây là item2
	//// Hàm Unmarshall trả về error
	//if err := json.Unmarshal([]byte(jsonStr), &item2); err != nil {
	//	fmt.Println(err)
	//	// Dùng return hoặc os Exit
	//	//os.Exit(1)
	//	return
	//}
	//fmt.Println(item2)

	r := gin.Default()
	/*
		CRUD: Create , Read , Update , Delete
		POST: /v1/items (tạo mới 1 item)
		GET:  /v1/items (lấy list items) /v1/items?page=1
		GET:  /vi/items/:id (lấy 1 item từ id của item cần lấy)
		PUT || PATCH:	/v1/items/:id (cập nhật 1 item từ id)
		DELETE:  /v1/items/:id (xóa 1 item từ id của item cần xóa)
	*/

	v1 := r.Group("/v1")
	{
		items := v1.Group("/items")
		{
			items.POST("", CreateItem(db))
			items.GET("", ListItem(db))
			items.GET("/:id", GetItem(db))
			items.PATCH("/:id", UpdateItem(db))
			items.DELETE("/:id", DeleteItem(db))
		}
	}

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.Run(":3000")
}

func CreateItem(db *gorm.DB) func(*gin.Context) {
	return func(c *gin.Context) {
		var data TodoItem

		if err := c.ShouldBind(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})

			return
		}

		if err := db.Create(&data).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})

			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": data.Id,
		})
	}
}

func GetItem(db *gorm.DB) func(*gin.Context) {
	return func(c *gin.Context) {
		var data TodoItem
		// /v1/items/1 : c.Param("id") -> kết quả sẽ ra là string "1" , phải convert string sang int
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		//Cách 1 :
		data.Id = id
		if err := db.First(&data).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"data": data,
		})
		/* Cách 2:
		if err := db.Where("id = ?", id).First(&data).Error; err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"data": data,
		})
		*/
	}
}
func UpdateItem(db *gorm.DB) func(*gin.Context) {
	return func(c *gin.Context) {
		var data TodoItemUpdate
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		if err := c.ShouldBind(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})

			return
		}
		//Cách 1 :
		if err := db.Where("id = ?", id).Updates(&data).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"data": true,
		})
		/* Cách 2:
		if err := db.Where("id = ?", id).Updates(&data).Error; err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"data": true,
		})
		*/
	}
}
func DeleteItem(db *gorm.DB) func(*gin.Context) {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		/*Hard Delete*/
		if err := db.Table(TodoItem{}.TableName()).Where("id = ?", id).Delete(nil).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"data": true,
		})
		/*Soft Delete - Update lại record của column status thành Deleted
		if err := db.Table(TodoItem{}.TableName()).Where("id = ?", id).Updates(map[string]interface{}{
			"status: "Deleted",
		}).Error;err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"data": true,
		})
		*/
	}
}

func ListItem(db *gorm.DB) func(*gin.Context) {
	return func(c *gin.Context) {
		var paging Paging

		if err := c.ShouldBind(&paging); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})

			return
		}
		paging.Process()

		var result []TodoItem

		/*Dùng câu query khi Soft Delete để không hiển thị các column status có giá trị Deleted hiển thị ra*/
		db = db.Where("status <> ?", "Deleted")

		if err := db.Table(TodoItem{}.TableName()).Count(&paging.Total).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})

			return
		}
		/*sắp xếp list theo id giảm dần DESC - tăng dần ASC (mặc định là ASC)
		if err := db.Order("id desc").Find(&result).Error;
		*/
		if err := db.Order("id desc").
			Offset((paging.Page - 1) * paging.Limit).
			Limit(paging.Limit).
			Find(&result).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})

			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": result,
			/*hiển thị total*/
			"paging": paging,
		})
	}
}
