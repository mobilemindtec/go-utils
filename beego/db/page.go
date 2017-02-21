package db

import (
	"fmt"
)

type Page struct {
  Offset int64
  Limit int64
  Search string
  Order string
  Sort string  
  FilterColumns map[string]interface{} 
  AndFilterColumns map[string]interface{}
}

func (this *Page) AddFilter(columnName string, value interface{}) *Page{
	
	if this.FilterColumns == nil {
		this.FilterColumns = make(map[string]interface{} )
	}	

	this.FilterColumns[columnName] = value

	return this
}

func (this *Page) AddFilterDefault(columnName string) *Page{	
	return this.AddFilter(fmt.Sprintf("%v__icontains", columnName), this.Search)
}

func (this *Page) AddFilterAnd(columnName string, value interface{}) *Page{	
	if this.AndFilterColumns == nil {
		this.AndFilterColumns = make(map[string]interface{} )
	}	

	this.AndFilterColumns[columnName] = value

	return this
}

func (this *Page) AddFilterDefaults(columnName ...string) *Page{	

	for _, s := range columnName {
		this.AddFilter(fmt.Sprintf("%v__icontains", s), this.Search)		
	}

	return this
}


func (this *Page) MakeDefaultSort() {
	if this.Order == "asc" {
		this.Order = ""
	} else {
		this.Order = "-"
	}	
}