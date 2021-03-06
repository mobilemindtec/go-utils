package db

import (
  "github.com/beego/beego/v2/client/orm"
  "reflect"
  "strings"
  "errors"
  "fmt"
)

type SessionState int

const (
  SessionStateOk SessionState = iota + 1
  SessionStateError
)

type Model interface {
  IsPersisted() bool
  TableName() string
}

type Session struct {
  State SessionState
  Tenant interface{}
  IgnoreTenantFilter bool
  Debug bool
  DbName string

  deepSetDefault map[string]int
  deepSaveOrUpdate map[string]int
  deepEager map[string]int
  deepRemove map[string]int  

  openDbError bool

  db *DataBase
  tx bool
}


func NewSession() *Session{
  return &Session{ State: SessionStateOk, Debug: false, db: NewDataBase("default") }
}

func NewSessionWithDbName(dbName string) *Session{
  return &Session{ State: SessionStateOk, Debug: false,  db: NewDataBase(dbName) }
}

func NewSessionWithTenant(tenant interface{}) *Session{
  return &Session{ State: SessionStateOk, Tenant: tenant, Debug: false, db: NewDataBase("default") }
}

func NewSessionWithTenantAndDbName(tenant interface{}, dbName string) *Session{
  return &Session{ State: SessionStateOk, Tenant: tenant, Debug: false, db: NewDataBase(dbName) }
}

func OrmVerbose(verbose bool){
  orm.Debug = verbose
}

func (this *Session) GetDb() *DataBase {
  return this.db
}

func (this *Session) SetTenant(tenant interface{}) *Session {
  this.Tenant = tenant
  return this
}

func (this *Session) OnError() *Session {
  this.SetError()
  return this
}

func (this *Session) SetError() *Session {
  this.State = SessionStateError
  return this
}

func (this *Session) RunWithTenant(tenant interface{}, runner func()) {

  tmp := this.Tenant
  this.Tenant = tenant

  defer func() {
    this.Tenant = tmp
  }()

  runner()
}

func (this *Session) IsOpenDbError() bool{
  return this.openDbError
}

func (this *Session) Open(withTx bool) error {
  if withTx {
    return this.OpenWithTx()
  } 
  return this.OpenWithoutTx()
}

func (this *Session) OpenTx() error {
  return this.OpenWithTx()
}

func (this *Session) OpenWithTx() error {
  this.tx = true
  return this.beginTx()
}

func (this *Session) OpenNoTx() error{  
  return this.OpenWithoutTx()
}

func (this *Session) OpenWithoutTx() error{
  this.tx = false
  this.db.Open()
  return nil
}

func (this *Session) Close() {

  if this.tx {
    if this.State == SessionStateOk {
      this.Commit()
    } else {
      this.Rollback()
    }
  } else {
    this.db = nil
  }
}

func (this *Session) beginTx() error {
  
  if err := this.db.Begin(); err != nil {

    this.openDbError = true

    fmt.Println("****************************************************************")
    fmt.Println("****************************************************************")
    fmt.Println("****************************************************************")
    fmt.Println("************************ db begin error: %v", err.Error())
    fmt.Println("****************************************************************")
    fmt.Println("****************************************************************")
    fmt.Println("****************************************************************")

    return err
    //panic(err)
  }

  //fmt.Println("tx openned = %v", this.tx)

  return nil

}

func (this *Session) Commit() error{

  if this.Debug {
    fmt.Println("## session commit ")
  }

  if this.db != nil {
    if err := this.db.Commit(); err != nil {
      fmt.Println("****************************************************************")
      fmt.Println("****************************************************************")
      fmt.Println("****************************************************************")
      fmt.Println("************************ db commit error: %v", err.Error())
      fmt.Println("****************************************************************")
      fmt.Println("****************************************************************")
      fmt.Println("****************************************************************")
      this.Rollback()
      //panic(err)
      return err
    }  
    this.db = nil
  }

  return nil
}

func (this *Session) Rollback() error{

  if this.Debug {
    fmt.Println("## session rollback ")
  }

  //fmt.Println("## session rollback ")

  fmt.Println("** Session Rollback ")
  
  if this.db != nil {
    if err := this.db.Rollback(); err != nil {
      fmt.Println("****************************************************************")
      fmt.Println("****************************************************************")
      fmt.Println("****************************************************************")
      fmt.Println("************************ db roolback error: %v", err.Error())
      fmt.Println("****************************************************************")
      fmt.Println("****************************************************************")
      fmt.Println("****************************************************************")
      return err
    }
    this.db = nil
  }

  return nil
}

func (this *Session) Save(entity interface{}) error {


  if !this.isTenantNil() && this.isSetTenant(entity) {
    if this.Debug {
      fmt.Println("## Save set tenant")
    }
    this.setTenant(entity)
  }

  _, err := this.GetDb().Insert(entity)

  if this.Debug {
    fmt.Println("## save data: %+v", entity)
  }

  if err != nil {
    fmt.Println("## Session: error on save: %v", err.Error())
    this.SetError()
    return err
  }

  return nil
}

func (this *Session) Update(entity interface{}) error {

  if !this.isTenantNil() && this.isSetTenant(entity) {
    if this.Debug {
      fmt.Println("## Update set tenant")
    }
    this.setTenant(entity)
  }

  _, err := this.GetDb().Update(entity)

  if this.Debug {
    fmt.Println("## update data: %+v", entity)
  }

  if err != nil {
    fmt.Println("## Session: error on update: %v", err.Error())
    this.SetError()
    return err
  }

  return nil
}

func (this *Session) Remove(entity interface{}) error {


  _, err := this.GetDb().Delete(entity)

  if err != nil {
    fmt.Println("## Session: error on remove: %v", err.Error())
    this.SetError()
    return err
  }

  return nil
}

func (this *Session) Load(entity interface{}) (bool, error) {
  return this.Get(entity)
}

func (this *Session) Get(entity interface{}) (bool, error) {

  if model, ok := entity.(Model); ok {

    err := this.GetDb().Read(entity)
    
    if err == orm.ErrNoRows {
      //fmt.Println("## Session: error on load: %v", err.Error())
      //this.SetError()
      return false, nil
    }
    

    if model.IsPersisted() {
      return true, nil
    }

    return false, nil

  }

  this.SetError()
  return false, errors.New("entity does not implements of Model")  
}

func (this *Session) Count(entity interface{}) (int64, error){

  if model, ok := entity.(Model); ok {

    query := this.GetDb().QueryTable(model.TableName())

    query = this.setTenantFilter(entity, query)

    num, err := query.Count()

    if err != nil {
      fmt.Println("## Session: error on count table %v: %v", model.TableName(), err.Error())
      //this.SetError()
    }
    return num, err
  }

  this.SetError()
  return 0, errors.New("entity does not implements of Model")
}

func (this *Session) HasById(entity interface{}, id int64) (bool, error) {

  if model, ok := entity.(Model); ok {
    query := this.GetDb().QueryTable(model.TableName()).Filter("id", id)

    query = this.setTenantFilter(entity, query)

    return query.Exist(), nil
  }

  this.SetError()
  return false, errors.New("entity does not implements of Model")
}

func (this *Session) FindById(entity interface{}, id int64) (interface{}, error) {

  if model, ok := entity.(Model); ok {
    query := this.GetDb().QueryTable(model.TableName()).Filter("id", id)

    query = this.setTenantFilter(entity, query)

    err := query.One(entity)

    if err == orm.ErrNoRows {
      return nil, nil
    }

    if err != nil{
      fmt.Println("## Session: error on find by id table %v: %v", model.TableName(), err.Error())
      //this.SetError()
      return nil, err
    }

    if !model.IsPersisted() {
      return nil, nil
    }

    return entity, nil
  }

  this.SetError()
  return false, errors.New("entity does not implements of Model")
}

func (this *Session) SaveOrUpdate(entity interface{}) error{


  if !this.isTenantNil() && this.isSetTenant(entity) {
    if this.Debug {
      fmt.Println("## Save set tenant")
    }
    this.setTenant(entity)
  }


  if model, ok := entity.(Model); ok {
    if model.IsPersisted() {
      if err := this.Update(entity); err != nil {
        return errors.New(fmt.Sprintf("error on update %v: %v", model.TableName(), err))
      }
      return nil
    }
    if err := this.Save(entity); err != nil {
      return errors.New(fmt.Sprintf("error on save %v: %v", model.TableName(), err))
    }    
    return nil
  }

  this.SetError()
  return errors.New(fmt.Sprintf("entity %v does not implements of Model", entity))
}

func (this *Session) List(entity interface{}, entities interface{}) error {
  if model, ok := entity.(Model); ok {

    query := this.GetDb().QueryTable(model.TableName())

    query = this.setTenantFilter(entity, query)

    if _, err := query.All(entities); err != nil {
      fmt.Println("## Session: error on list: %v", err.Error())
      //this.SetError()
      return err
    }
    return nil
  }

  this.SetError()
  return errors.New("entity does not implements of Model 1")
}

func (this *Session) Page(entity interface{}, entities interface{}, page *Page) error {
  return this.PageQuery(nil, entity, entities, page)
}

func (this *Session) PageQuery(query orm.QuerySeter, entity interface{}, entities interface{}, page *Page) error {
  if model, ok := entity.(Model); ok {

    if query == nil {
      query = this.GetDb().QueryTable(model.TableName())
    }

    query = query.Limit(page.Limit).Offset(page.Offset)

    switch page.Sort {
      case "asc":
        query = query.OrderBy(fmt.Sprintf("%v", page.Sort))
      case "desc":
        query = query.OrderBy(fmt.Sprintf("-%v", page.Sort))
    }

    if page.FilterColumns != nil && len(page.FilterColumns) > 0 {

      if len(page.FilterColumns) == 1 {
        for k, v := range page.FilterColumns {
          query = query.Filter(k, v)
        }
      } else {
        cond := orm.NewCondition()
        for k, v := range page.FilterColumns {
          cond = cond.Or(k, v)
        }
        query = query.SetCond(orm.NewCondition().AndCond(cond))
      }

    }

    if page.AndFilterColumns != nil && len(page.AndFilterColumns) > 0 {
      for k, v := range page.AndFilterColumns {
        query = query.Filter(k, v)
      }
    }

    query = this.setTenantFilter(entity, query)

    if _, err := query.All(entities); err != nil {
      fmt.Println("## Session: error on page: %v", err.Error())
      //this.SetError()
      return err
    }


    return nil
  }

  this.SetError()
  return errors.New("entity does not implements of Model")
}


func (this *Session) Query(entity interface{}) (orm.QuerySeter, error) {
  if model, ok := entity.(Model); ok {
    query := this.GetDb().QueryTable(model.TableName())

    query = this.setTenantFilter(entity, query)

    return query, nil
  }

  this.SetError()
  return nil, errors.New("entity does not implements of Model")
}

func (this *Session) ToList(querySeter orm.QuerySeter, entities interface{}) error {
  if _, err := querySeter.All(entities); err != nil {
    fmt.Println("## Session: error on to list: %v", err.Error())
    //this.SetError()
    return err
  }
  return nil
}

func (this *Session) ToOne(querySeter orm.QuerySeter, entity interface{}) error {
  if err := querySeter.One(entity); err != nil {
    fmt.Println("## Session: error on to one: %v", err.Error())
    //this.SetError()
    return err
  }
  return nil
}

func (this *Session) ToPage(querySeter orm.QuerySeter, entities interface{}, page *Page) error {
  querySeter.Limit(page.Limit)
  querySeter.Offset(page.Offset)
  if _, err := querySeter.All(entities); err != nil {
    fmt.Println("## Session: error on to page: %v", err.Error())
    //this.SetError()
    return err
  }
  return nil
}

func (this *Session) ToCount(querySeter orm.QuerySeter) (int64, error) {

  var count int64
  var err error

  if count, err = querySeter.Count(); err != nil {
    fmt.Println("## Session: error on to count: %v", err.Error())
    //this.SetError()
    return count, err
  }
  return count, err
}

func (this *Session) ExecuteDelete(querySeter orm.QuerySeter) (int64, error) {

  var count int64
  var err error
  
  if count, err = querySeter.Delete(); err != nil {
    fmt.Println("## Session: error on to list: %v", err.Error())
    //this.SetError()
    return count, err
  }
  
  return count, err
}

func (this *Session) ExecuteUpdate(querySeter orm.QuerySeter, args map[string]interface{}) (int64, error) {

  var count int64
  var err error
  var params orm.Params

  for k, v := range args {
    params[k] = v
  }
  
  if _, err := querySeter.Update(params); err != nil {
    fmt.Println("## Session: error on to list: %v", err.Error())
    //this.SetError()
    return count, err
  }

  return count, err
}

func (this *Session) Raw(query string, args ...interface{}) orm.RawSeter{  
  return this.GetDb().Raw(query, args...)
}


func (this *Session) Eager(reply interface{}) error{
  this.deepEager = map[string]int{}
  return this.eagerDeep(reply, false)
}

func (this *Session) EagerForce(reply interface{}) error{
  this.deepEager = map[string]int{}
  return this.eagerDeep(reply, true)
}

func (this *Session) eagerDeep(reply interface{}, ignoreTag bool) error{

  if reply == nil {
    if this.Debug {
      fmt.Println("## reply is nil")
    }
    return nil
  }

  // value e type of pointer
  refValue := reflect.ValueOf(reply)
  //refType := reflect.TypeOf(reply)


  // value e type of instance
  fullValue := refValue.Elem()
  fullType := fullValue.Type()

  for i := 0; i < fullType.NumField(); i++ {
    field := fullType.Field(i)

    fieldStruct := fullValue.FieldByName(field.Name)
    fieldValue := fieldStruct.Interface()
    //fieldType := fieldStruct.Type()

    tags := this.getTags(field)

    if !ignoreTag{
      if tags == nil || len(tags) == 0 || !this.hasTag(tags, "eager"){
        continue
      }
    }

    if tags != nil && this.hasTag(tags, "ignore_eager"){
      continue
    }

    zero := reflect.Zero(reflect.TypeOf(fieldValue)).Interface() == fieldValue
    model, ok := fieldValue.(Model)

    if zero {

      if this.Debug {
        fmt.Println("## no eager zero field: ", field.Name)
      }

    } else if !ok {

      if this.Debug {
        fmt.Println("## no eager. field does not implemente model: ", field.Name)
      }

    } else {

      if model.IsPersisted() {

        if this.Debug {
          fmt.Println("## eager field: ", field.Name, fieldValue)
        }

        if _, err := this.GetDb().LoadRelated(reply, field.Name); err != nil {
          fmt.Println("********* eager field error ", fullType, field.Name, fieldValue, err.Error())
        } else {
          // reload loaded value of field reference
          refValue = reflect.ValueOf(reply)
          fullValue = refValue.Elem()
          fullType = fullValue.Type()
          field = fullType.Field(i)
          fieldStruct = fullValue.FieldByName(field.Name)
          fieldValue = fieldStruct.Interface()

          key := fmt.Sprintf("%v.%v", strings.Split(fullType.String(), ".")[1], field.Name)
          if count, ok := this.deepEager[key]; ok {

            if count >= 5 {
              continue
            }

            this.deepEager[key] = count + 1
          } else {
            this.deepEager[key] = 1
          }

          if this.Debug {
            fmt.Println("## eager field success: ", field.Name, fieldValue)
          }

        }
      } else {
        if this.Debug {
          fmt.Println("## not eager field not persisted: ", field.Name)
        }
      }

      if tags != nil && this.hasTag(tags, "ignore_eager_child"){
        continue
      }

      if this.Debug {
        fmt.Println("## eager next field: ", field.Name)
      }

      if err := this.eagerDeep(fieldValue, ignoreTag); err != nil {
        fmt.Println("## eager next field %v: %v", field.Name, err.Error())
        return err
      }
    }

  }

  return nil
}

func (this *Session) SaveCascade(reply interface{}) error{
  this.deepSaveOrUpdate = map[string]int{}
  return this.saveOrUpdateCascadeDeep(reply)
}


func (this *Session) SaveOrUpdateCascade(reply interface{}) error{
  this.deepSaveOrUpdate = map[string]int{}
  return this.saveOrUpdateCascadeDeep(reply)
}

func (this *Session) saveOrUpdateCascadeDeep(reply interface{}) error{

  // value e type of pointer
  refValue := reflect.ValueOf(reply)
  //refType := reflect.TypeOf(reply)


  // value e type of instance
  fullValue := refValue.Elem()
  fullType := fullValue.Type()

  for i := 0; i < fullType.NumField(); i++ {
    field := fullType.Field(i)

    fieldStruct := fullValue.FieldByName(field.Name)
    fieldValue := fieldStruct.Interface()
    //fieldType := fieldStruct.Type()

    tags := this.getTags(field)

    if tags == nil || len(tags) == 0 {
      continue
    }

    if tags == nil || !this.hasTag(tags, "save_or_update_cascade"){
      continue
    }

    zero := reflect.Zero(reflect.TypeOf(fieldValue)).Interface() == fieldValue
    _, ok := fieldValue.(Model)

    if zero {

      if this.Debug {
        fmt.Println("## no cascade zero field: ", field.Name)
      }

    } else if !ok {

      if this.Debug {
        fmt.Println("## no cascade. field does not implemente model: ", field.Name)
      }

    } else {

      if this.Debug {
        fmt.Println("## cascade field: ", field.Name)
      }

      key := fmt.Sprintf("%v.%v", strings.Split(fullType.String(), ".")[1], field.Name)
      if count, ok := this.deepSaveOrUpdate[key]; ok {

        if count >= 5 {
          continue
        }

        this.deepSaveOrUpdate[key] = count + 1
      } else {
        this.deepSaveOrUpdate[key] = 1
      }

      if err := this.saveOrUpdateCascadeDeep(fieldValue); err != nil {
        return err
      }
    }

  }

  if this.Debug {
    fmt.Println("## save or update: ", fullType)
  }

  return this.SaveOrUpdate(reply)
}

func (this *Session) RemoveCascade(reply interface{}) error{
  this.deepRemove = map[string]int{}
  return this.RemoveCascadeDeep(reply)
}

func (this *Session) RemoveCascadeDeep(reply interface{}) error{

  // value e type of pointer
  refValue := reflect.ValueOf(reply)
  //refType := reflect.TypeOf(reply)


  // value e type of instance
  fullValue := refValue.Elem()
  fullType := fullValue.Type()

  var itensToRemove []interface{}

  for i := 0; i < fullType.NumField(); i++ {
    field := fullType.Field(i)

    fieldStruct := fullValue.FieldByName(field.Name)
    fieldValue := fieldStruct.Interface()
    //fieldType := fieldStruct.Type()

    tags := this.getTags(field)

    if tags == nil || len(tags) == 0 {
      continue
    }

    if tags == nil || !this.hasTag(tags, "remove_cascade"){
      continue
    }

    zero := reflect.Zero(reflect.TypeOf(fieldValue)).Interface() == fieldValue
    _, ok := fieldValue.(Model)

    if zero {

      if this.Debug {
        fmt.Println("## no cascade remove zero field: ", field.Name)
      }

    } else if !ok {

      if this.Debug {
        fmt.Println("## no cascade remove. field does not implemente model: ", field.Name)
      }

    } else {

      //fieldFullType := reflect.TypeOf(fieldValue).Elem()

      key := fmt.Sprintf("%v.%v", strings.Split(fullType.String(), ".")[1], field.Name)

      if count, ok := this.deepRemove[key]; ok {

        if count >= 5 {
          continue
        }

        this.deepRemove[key] = count + 1
      } else {
        this.deepRemove[key] = 1
      }

      if this.Debug {
        fmt.Println("## cascade remove field: ", field.Name)
      }

      itensToRemove = append(itensToRemove, fieldValue)
    }

  }

  if this.Debug {
    fmt.Println("## remove: ", fullType)
  }

  if err := this.Remove(reply); err != nil {
    return err
  }

  for _, it := range itensToRemove {
    if err := this.RemoveCascadeDeep(it); err != nil {
      return err
    }    
  }

  return nil
}


func (this *Session) SetDefaults(reply interface{}) error{

  this.deepSetDefault = map[string]int{}

  return this.setDefaultsDeep(reply)
}

func (this *Session) setDefaultsDeep(reply interface{}) error{

  if reply == nil {
    return nil
  }

  // value e type of pointer
  refValue := reflect.ValueOf(reply)
  //refType := reflect.TypeOf(reply)


  // value e type of instance
  fullValue := refValue.Elem()
  fullType := fullValue.Type()

  for i := 0; i < fullType.NumField(); i++ {
    field := fullType.Field(i)

    fieldStruct := fullValue.FieldByName(field.Name)
    fieldValue := fieldStruct.Interface()
    //fieldType := fieldStruct.Type()


    tags := this.getTags(field)

    if tags != nil && this.hasTag(tags, "ignore_set_default"){
      continue
    }


    zero := reflect.Zero(reflect.TypeOf(fieldValue)).Interface() == fieldValue
    _, ok := fieldValue.(Model)

    if ok {

      if this.Debug {
        fmt.Println("set defaults to ", fullValue, field.Name)
      }

      if zero {
        fieldFullType := reflect.TypeOf(fieldValue).Elem()
        newRefValue := reflect.New(fieldFullType)
        fieldStruct.Set(newRefValue)
        fieldValue = fieldStruct.Interface()
      }

      key := fmt.Sprintf("%v.%v", strings.Split(fullType.String(), ".")[1], field.Name)

      if count, ok := this.deepSetDefault[key]; ok {

        if count >= 5 {
          continue
        }

        this.deepSetDefault[key] = count + 1
      } else {
        this.deepSetDefault[key] = 1
      }

      if tags != nil && this.hasTag(tags, "ignore_set_default_child"){
        continue
      }

      this.setDefaultsDeep(fieldValue)
    }

  }

  return nil
}

func (this *Session) hasTag(tags []string, tagName string) bool{

  for _, tag := range tags {
    if tag == tagName {
      return true
    }
  }

  return false
}

func (this *Session) setTenant(reply interface{}){

  // value e type of pointer
  refValue := reflect.ValueOf(reply)
  //refType := reflect.TypeOf(reply)


  // value e type of instance
  fullValue := refValue.Elem()
  fullType := fullValue.Type()

  if this.Debug {
    fmt.Println("## set Tenant to ", fullType)
  }


  for i := 0; i < fullType.NumField(); i++ {
    field := fullType.Field(i)

    fieldStruct := fullValue.FieldByName(field.Name)
    //fieldValue := fieldStruct.Interface()
    //fieldType := fieldStruct.Type()


    tags := this.getTags(field)

    if this.hasTag(tags, "tenant") {


      value := reflect.ValueOf(this.Tenant)

      if this.Debug {
        fmt.Println("## field %v is tenant set tenant %v", field.Name, value)
      }

      // check is null
      //fmt.Println("## set Tenant to ", fullType)
      elem := fieldStruct.Elem()
      //fmt.Println("elem = %v", elem)
      //valueOf := reflect.ValueOf(elem)
      //fmt.Println("valueOf = %v, IsValid = %v", valueOf, elem.IsValid())
      if !elem.IsValid() {
        //fmt.Println("==== auto set tenant")
        fieldStruct.Set(value)        
      }else{
        //fmt.Println("==== auto set tenant: tenant already")
      }

    } else {
      if this.Debug {
        fmt.Println("## field %v not is tenant ", field.Name)
      }
    }

  }

}

func (this *Session) getTags(field reflect.StructField) []string{

  tag := field.Tag.Get("goutils")
  var tags []string

  if len(strings.TrimSpace(tag)) > 0 {
    tags = strings.Split(tag, ";")
  }

  return tags
}

func (this *Session) setTenantFilter(entity interface{}, query orm.QuerySeter) orm.QuerySeter {


  if !this.isTenantNil() && this.HasFilterTenant(entity) {
    if model, ok := this.Tenant.(Model); ok {
      if model.IsPersisted() {
        query = query.Filter("Tenant", this.Tenant)
      }
    }
  }

  return query
}

func (this *Session) isTenantNil() bool{
  if this.Tenant != nil {
    value := reflect.ValueOf(this.Tenant)

    return value.IsNil()
  }

  return true
}

func (this *Session) isSetTenant(reply interface{}) bool{

  // value e type of pointer
  refValue := reflect.ValueOf(reply)
  //refType := reflect.TypeOf(reply)


  // value e type of instance
  fullValue := refValue.Elem()
  fullType := fullValue.Type()

  if this.Debug {
    fmt.Println("## set Tenant to ", fullType)
  }


  for i := 0; i < fullType.NumField(); i++ {
    field := fullType.Field(i)

		if field.Name == "Tenant" {

			tags := this.getTags(field)

		  if tags == nil || len(tags) == 0 {
        //fmt.Println("## set tenant")
		    return true
		  }

		  set := !this.hasTag(tags, "no_set_tenant")

      //fmt.Println("## set tenant = %v", set)

      return set

		}

	}

  //fmt.Println("## not set tenant")
	return false
}

func (this *Session) HasFilterTenant(reply interface{}) bool{

  // value e type of pointer
  refValue := reflect.ValueOf(reply)
  //refType := reflect.TypeOf(reply)


  // value e type of instance
  fullValue := refValue.Elem()
  fullType := fullValue.Type()

  for i := 0; i < fullType.NumField(); i++ {
    field := fullType.Field(i)

		if field.Name == "Tenant" {

			tags := this.getTags(field)

		  if !this.hasTag(tags, "tenant") {
        //fmt.Println("## filter tenant")
		    return false
		  }

      if this.IgnoreTenantFilter {
        return false
      }

		  noFilter := this.hasTag(tags, "no_filter_tenant")

      //fmt.Println("## filter tenant = %v", filter)

      if noFilter {
        return false
      }

      return true

		}

	}
	//fmt.Println("## filter tenant")
	return false
}
