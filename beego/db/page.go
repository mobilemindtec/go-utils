package db

import (
	"fmt"
	"strings"
)

type Page struct {
  Offset int64
  Limit int64
  Search	 string
  Order string
  Sort string  
  FilterColumns map[string]interface{} 
  AndFilterColumns map[string]interface{}
  TenantColumnFilter map[string]interface{}
}

func NewPage(offset int64, limit int64, sort string, order string) *Page{
	return &Page{ Offset: offset, Limit: limit, Sort: sort, Order: order }
}

/* deprecated */
func (this *Page) AddFilter(columnName string, value interface{}) *Page{
	
	if this.FilterColumns == nil {
		this.FilterColumns = make(map[string]interface{} )
	}	

	this.FilterColumns[columnName] = value

	return this
}

/* deprecated */
func (this *Page) AddFilterDefault(columnName string) *Page{	
	return this.AddFilter(fmt.Sprintf("%v__icontains", columnName), this.Search)
}

/* deprecated */
func (this *Page) AddFilterAnd(columnName string, value interface{}) *Page{	
	if this.AndFilterColumns == nil {
		this.AndFilterColumns = make(map[string]interface{} )
	}	

	this.AndFilterColumns[columnName] = value

	return this
}

/* deprecated */
func (this *Page) AddFilterDefaults(columnName ...string) *Page{	

	for _, s := range columnName {
		if len(strings.TrimSpace(this.Search)) > 0 {
			this.AddFilter(fmt.Sprintf("%v__icontains", s), this.Search)		
		}
	}

	return this
}

func (this *Page) AddAndInOrConditionFilter(columnName string, value interface{}) *Page{
	if this.TenantColumnFilter == nil {
		this.TenantColumnFilter = make(map[string]interface{} )
	}		
	this.TenantColumnFilter[columnName] = value
	return this
}

/* deprecated */
func (this *Page) MakeDefaultSort() {

}