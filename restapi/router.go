package restapi

import (
	"gopcep/controller"

	"github.com/gin-gonic/gin"
)

func (h *handler) createUpdRouter(c *gin.Context) {
	var r controller.Router
	err := c.BindJSON(&r)
	if err != nil {
		c.AbortWithStatusJSON(500, err)
		return
	}
	err = h.ctr.CreateUpdRouter(&r)
	if err != nil {

		c.AbortWithStatusJSON(500, err)
		return
	}
	c.JSON(200, r)
}

func (h *handler) deleteRouter(c *gin.Context) {
	err := h.ctr.DeleteRouter(c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(500, err)
		return
	}
	c.JSON(200, c.Param("id"))
}

func (h *handler) listRouters(c *gin.Context) {
	routers, err := h.ctr.GetRouters()
	if err != nil {
		c.AbortWithStatusJSON(500, err)
		return
	}
	c.JSON(200, routers)
}
