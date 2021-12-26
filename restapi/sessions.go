package restapi

import "github.com/gin-gonic/gin"

func (h *handler) getSessions(c *gin.Context) {

	data := make(map[string]interface{})
	h.ctr.RLock()
	for k, s := range h.ctr.PCEPSessions {
		data[k] = s.CopyToExportableSession()
	}
	h.ctr.RUnlock()

	c.JSON(200, data)
}
