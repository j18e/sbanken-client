package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/j18e/sbanken-client/pkg/models"
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

func (s *Server) handlerPurchase() gin.HandlerFunc {
	return func(c *gin.Context) {
		p, err := s.Storage.GetPurchase(c.Param("purchase"))
		if err != nil {
			c.AbortWithError(http.StatusNotFound, err)
			return
		}
		c.HTML(http.StatusOK, "purchase.html", gin.H{
			"title":    fmt.Sprintf("Purchase on %s", p.Date.Stamp()),
			"purchase": p,
		})
	}
}
