package models

import (
  "github.com/mobilemindtec/go-utils/beego/db"
  "github.com/satori/go.uuid"
	"time"
)

type Tenant struct{

  Id int64 `form:"-" json:",string,omitempty"`
  CreatedAt time.Time `orm:"auto_now_add;type(datetime)" json:"-"`
  UpdatedAt time.Time `orm:"auto_now;type(datetime)" json:"-"`

  Name string `orm:"size(100)"  valid:"Required;MaxSize(100)" form:""`
  Documento string `orm:"size(20)"  valid:"Required;MaxSize(14);MinSize(11)" form:""`

  Enabled bool `orm:""  form:"" json:""`
  Uuid string `orm:"size(100);unique"  valid:"MaxSize(100)" form:"-" json:""`

  Cidade *Cidade `orm:"rel(fk);on_delete(do_nothing)" valid:"RequiredRel" form:""`

  Session *db.Session `orm:"-"`
}

func (this *Tenant) TableName() string{
  return "tenants"
}

func NewTenant(session *db.Session) *Tenant{
  return &Tenant{ Session: session }
}

func (this *Tenant) IsPersisted() bool{
  return this.Id > 0
}

func (this *Tenant) GenereteUuid() string{

  for true {
    uuid := uuid.NewV4().String()
    if !db.NewCriteria(this.Session, new(Tenant), nil).Eq("Uuid", uuid).Exists() {
      return uuid
    }
  }

  return ""
}

func (this *Tenant) List() (*[]*Tenant , error) {
  var results []*Tenant
  err := this.Session.List(this, &results)
  return &results, err
}

func (this *Tenant) Page(page *db.Page) (*[]*Tenant , error) {
  var results []*Tenant

  page.AddFilterDefault("Name").MakeDefaultSort()

  err := this.Session.Page(this, &results, page)
  return &results, err
}

func (this *Tenant) GetByUuid(uuid string) (*Tenant , error) {

  entity := new(Tenant)
  criteria := db.NewCriteria(this.Session, entity, nil).Eq("Uuid", uuid).One()

  return entity, criteria.Error
}

func (this *Tenant) GetByUuidAndEnabled(uuid string) (*Tenant , error) {

  entity := new(Tenant)
  criteria := db.NewCriteria(this.Session, entity, nil).Eq("Uuid", uuid).Eq("Enabled", true).One()

  return entity, criteria.Error

}

func (this *Tenant) FindByDocumento(documento string) (*Tenant , error) {

  entity := new(Tenant)
  criteria := db.NewCriteria(this.Session, entity, nil).Eq("Documento", documento).One()

  return entity, criteria.Error

}

func (this *Tenant) LoadRelated(entity *Tenant) {

  this.Session.Load(entity.Cidade)
  this.Session.Load(entity.Cidade.Estado)

}