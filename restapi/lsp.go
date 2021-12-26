package restapi

import (
	"gopcep/pcep"

	"github.com/gin-gonic/gin"
)

func (h *handler) getNetLSPs(c *gin.Context) {
	c.JSON(200, h.ctr.GetSRLSPs())
}

func (h *handler) getLSPs(c *gin.Context) {
	c.JSON(200, h.ctr.GetLSPs())
}

func (h *handler) createUpdLSP(c *gin.Context) {
	var lsp pcep.SRLSP

	err := c.BindJSON(&lsp)
	if err != nil {
		c.AbortWithStatusJSON(500, map[string]string{
			"msg": err.Error(),
		})
		return
	}

	err = h.ctr.CreateUpdSRLSP(&lsp)
	if err != nil {
		c.AbortWithStatusJSON(500, map[string]string{
			"msg": err.Error(),
		})
		return
	}
	c.JSON(200, lsp)
}

func (h *handler) delLSP(c *gin.Context) {

	err := h.ctr.DelSRLSP(c.Param("name"))
	if err != nil {
		c.AbortWithStatusJSON(500, err)
		return
	}

	c.JSON(200, c.Param("name"))
}
