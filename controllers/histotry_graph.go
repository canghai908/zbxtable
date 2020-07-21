package controllers

type EchartController struct {
	BaseController
}

//URLMapping st
func (c *EchartController) URLMapping() {
	c.Mapping("GetHistory", c.GetHistory)
}

// GetHistory controller
// @Title Get One
// @Description get item by key
// @Param	X-Token		header  string	true		"x-token in header"
// @Param	item_id		query 	string	false		"The key for item"
// @Param	begin 	    query 	string	false		"history type"
// @Param	end	 	    query 	int		false		"The key for limit"
// @Success 200 {object} models.Item
// @Failure 403 :id is empty
// @router /history [get]
func (*EchartController) GetHistory() {
	// var p models.Item
	// p.Itemid
	// p, err := models.GetApplicationByHostid(10084)
	// if err != nil {
	// 	log.Println(err)
	// }
	// fmt.Println(p)
	// data, err := models.GetHistoryByItemIDNewP(23296, 1590390322, 1590476723)
	// if err != nil {
	// 	log.Println(err)
	// }
	// fmt.Println(data)
}
