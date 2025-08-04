package middleware

import (
	"log"
	"net/http"
	"stream-demo/backend/dto/response"
	"stream-demo/backend/utils"

	"github.com/gin-gonic/gin"
)

// ErrorHandler 錯誤處理中間件
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 檢查是否有錯誤
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			log.Printf("錯誤: %v", err)

			// 根據錯誤類型返回適當的響應
			switch e := err.Err.(type) {
			case *utils.AppError:
				c.JSON(e.StatusCode, response.NewErrorResponse(e.StatusCode, e.Message))
			default:
				c.JSON(http.StatusInternalServerError, response.NewErrorResponse(http.StatusInternalServerError, "內部伺服器錯誤"))
			}
		}
	}
}

// Recovery 恢復中間件
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				utils.LogError("Panic recovered: %v", err)
				c.JSON(http.StatusInternalServerError, response.NewErrorResponse(500, "伺服器內部錯誤"))
				c.Abort()
			}
		}()
		c.Next()
	}
}
