package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/j18e/sbanken-client/pkg/models"
	"github.com/j18e/sbanken-client/pkg/storage"
)

func (s *Server) handlerHome() gin.HandlerFunc {
	return func(c *gin.Context) {
		n := time.Now()
		c.Redirect(http.StatusFound, fmt.Sprintf("/spending/%04d/%02d", n.Year(), n.Month()))
	}
}

func (s *Server) handlerSpendingMonth() gin.HandlerFunc {
	type PathArgs struct {
		Year  int        `binding:"required" uri:"year"`
		Month time.Month `binding:"required" uri:"month"`
	}
	return func(c *gin.Context) {
		var path PathArgs
		if err := c.ShouldBindUri(&path); err != nil {
			c.String(http.StatusNotFound, err.Error())
			return
		}

		month := models.Date{Year: path.Year, Month: path.Month, MonthNum: int(path.Month)}
		purchases, err := s.Storage.GetPurchases(month)
		if err != nil {
			c.String(http.StatusInternalServerError, "an error occurred: %v", err)
			return
		}

		var total int
		for _, p := range purchases {
			total += p.NOK
		}

		c.HTML(http.StatusOK, "spending.html", gin.H{
			"title":     fmt.Sprintf("Spending in %s", month),
			"payload":   purchases,
			"month":     month,
			"prevMonth": month.SubMonth(),
			"nextMonth": month.AddMonth(),
			"total":     total,
		})
	}
}

func (s *Server) handlerAPIPurchases() gin.HandlerFunc {
	return func(c *gin.Context) {
		var params struct {
			Year  int `uri:"year" binding:"required"`
			Month int `uri:"month" binding:"required"`
		}
		if err := c.BindUri(&params); err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		p, err := s.Storage.GetPurchases(models.Date{
			Year:     params.Year,
			Month:    time.Month(params.Month),
			MonthNum: params.Month,
		})
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		c.JSON(http.StatusOK, p)
	}
}

func (s *Server) handlerAPIPurchase() gin.HandlerFunc {
	return func(c *gin.Context) {
		p, err := s.Storage.GetPurchase(c.Param("purchase"))
		if err != nil {
			c.AbortWithError(http.StatusNotFound, err)
			return
		}
		c.JSON(http.StatusOK, p)
	}
}

// func (s *Server) handlerAPIPurchaseUpdate() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		p, err := s.Storage.GetPurchase(c.Param("purchase"))
// 		if err != nil {
// 			c.AbortWithError(http.StatusNotFound, err)
// 			return
// 		}
// 		c.JSON(http.StatusOK, p)
// 	}
// }

func (s *Server) handlerAPIPurchaseDelete() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := s.Storage.DeletePurchase(c.Param("purchase")); err != nil {
			if err == storage.ErrNotFound {
				c.String(http.StatusNotFound, "purchase not found")
			} else {
				c.AbortWithError(http.StatusInternalServerError, err)
			}
			return
		}
		c.String(http.StatusOK, "purchase deleted")
	}
}
