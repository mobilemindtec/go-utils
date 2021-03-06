package web

import (
  "github.com/mobilemindtec/go-utils/beego/validator"
  "github.com/mobilemindtec/go-utils/beego/filters"
  "github.com/mobilemindtec/go-utils/beego/db"
  "github.com/mobilemindtec/go-utils/support"
  "github.com/beego/beego/v2/core/validation"
  "github.com/leekchan/accounting"  
  beego "github.com/beego/beego/v2/server/web"
  "github.com/beego/beego/v2/core/logs"
  "github.com/beego/i18n"
  "html/template"
  "strings"
  "strconv"
  "time"
  "fmt"
  "runtime/debug"
)

var (
  langTypes []string // Languages that are supported.
  datetimeLayout = "02/01/2006 15:04:05"
  timeLayout = "10:25"
  dateLayout = "02/01/2006"
  jsonDateLayout = "2006-01-02T15:04:05-07:00"
)

type BaseController struct {
  EntityValidator *validator.EntityValidator
  beego.Controller
  Flash *beego.FlashData
  Session *db.Session
  support.JsonParser
  ViewPath string  
  i18n.Locale

  defaultPageLimit int64
}

type RecoverInfo struct {
  Error string
  StackTrace string
}

type NestPreparer interface {
  NestPrepare()
}

type NestFinisher interface {
  NestFinish()
}

type NestRecover interface {
  NextOnRecover(info * RecoverInfo)
}

func init() {
  LoadIl8n()
  LoadFuncs(nil)
}

func LoadFuncs(controller *BaseController) {
  inc := func(i int) int {
      return i + 1
  }

  hasError := func(args map[string]string, key string) string{
    if args[key] != "" {
      return "has-error"
    }
    return ""
  }

  errorMsg := func(args map[string]string, key string) string{
    return args[key]
  }

  currentYaer := func () string {
    return strconv.Itoa(time.Now().Year())
  }

  formatMoney := func(number float64) string{
    ac := accounting.Accounting{Symbol: "R$ ", Precision: 2, Thousand: ",", Decimal: "."}
    return ac.FormatMoney(number)
  }

  isZeroDate := func(date time.Time) bool{
    return time.Time.IsZero(date)
  }

  formatDate := func(date time.Time) string{
    if !time.Time.IsZero(date) {
      return date.Format("02/01/2006")
    }
    return ""
  }

  formatDateTime := func(date time.Time) string{
    if !time.Time.IsZero(date) {
      return date.Format("02/01/2006 15:04")
    }
    return ""
  }

  dateFormat := func(date time.Time, layout string) string{
    if !time.Time.IsZero(date) {
      return date.Format(layout)
    }
    return ""
  }

  getNow := func(layout string) string{
    return time.Now().Format(layout)
  }

  getYear := func() string{
    return time.Now().Format("2006")
  }

  formatBoolean := func(b bool, wrapLabel bool) string{
    var s string
    if b {
      s = "Sim"
    }else{
      s = "Não"
    }
    if wrapLabel {
      var class string
      if b {
        class = "info"
      }else{
        class = "danger"
      }
      val := "<span class='label label-" + class + "'>"+s+"</span>"
      s = val
    }
    return s
  }

  formatDecimal := func(number float64) string{
    ac := accounting.Accounting{Symbol: "", Precision: 2, Thousand: ",", Decimal: "."}
    return ac.FormatMoney(number)
  }

  sum := func(numbers ...float64) float64{
    total := 0.0
    for i, it := range numbers {
      if i == 0 {
        total = it
      } else {
        total += it
      }
    }
    return total
  }

  subtract := func(numbers ...float64) float64{
    total := 0.0
    for i, it := range numbers {
      if i == 0 {
        total = it  
      } else {
        total -= it
      }
    }
    return total
  }

  mult := func(numbers ...float64) float64{
    total := 0.0
    for i, it := range numbers {
      if i == 0 {
        total = it
      } else {
        total *= it
      }
    }
    return total
  }

  numberMask := func(text interface{}, mask string) string {
    return support.NumberMask(fmt.Sprintf("%v", text), mask)
  }

  numberMaskReverse := func(text interface{}, mask string) string {
    return support.NumberMaskReverse(fmt.Sprintf("%v", text), mask)
  }

  beego.AddFuncMap("is_zero_date", isZeroDate)
  
  beego.AddFuncMap("inc", inc)
  beego.AddFuncMap("has_error", hasError)
  beego.AddFuncMap("error_msg", errorMsg)
  beego.AddFuncMap("current_yaer", currentYaer)
  beego.InsertFilter("*", beego.BeforeRouter, filters.FilterMethod) // enable put
  beego.AddFuncMap("format_boolean", formatBoolean)
  beego.AddFuncMap("format_date", formatDate)
  beego.AddFuncMap("date_format", dateFormat)
  beego.AddFuncMap("get_now", getNow)
  beego.AddFuncMap("get_year", getYear)
  beego.AddFuncMap("format_date_time", formatDateTime)
  beego.AddFuncMap("format_money", formatMoney)
  beego.AddFuncMap("format_decimal", formatDecimal)
  beego.AddFuncMap("sum", sum)
  beego.AddFuncMap("subtract", subtract)
  beego.AddFuncMap("mult", mult)

  beego.AddFuncMap("mask", numberMask)
  beego.AddFuncMap("mask_reverse", numberMaskReverse)
}

func LoadIl8n() {
  beego.AddFuncMap("i18n", i18n.Tr)
  logs.SetLevel(logs.LevelDebug)

  // Initialize language type list.
  types, _:= beego.AppConfig.String("lang_types")
  langTypes = strings.Split(types, "|")

  logs.Info(" langTypes %v", langTypes)

  // Load locale files according to language types.
  for _, lang := range langTypes {
    if err := i18n.SetMessage(lang, "conf/i18n/"+"locale_" + lang + ".ini"); err != nil {
      logs.Error("Fail to set message file:", err)
      return
    }
  }
}

// Prepare implemented Prepare() method for baseController.
// It's used for language option check and setting.
func (this *BaseController) NestPrepareBase () {

  //this.Log("** web.BaseController.NestPrepareBase")

  // Reset language option.
  this.Lang = "" // This field is from i18n.Locale.

  // 1. Get language information from 'Accept-Language'.
  al := this.Ctx.Request.Header.Get("Accept-Language")
  if len(al) > 4 {
    al = al[:5] // Only compare first 5 letters.
    if i18n.IsExist(al) {
      this.Lang = al
    }
  }


  // 2. Default language is English.
  if len(this.Lang) == 0 {
    this.Lang = "pt-BR"
  }

  //this.Log(" ** use language %v", this.Lang)

  this.Flash = beego.NewFlash()

  // Set template level language option.
  this.Data["Lang"] = this.Lang
  this.Data["xsrfdata"]= template.HTML(this.XSRFFormHTML())
  this.Data["dateLayout"] = dateLayout
  this.Data["datetimeLayout"] = datetimeLayout
  this.Data["timeLayout"] = timeLayout


  this.Session = db.NewSession()
  var err error
  err = this.Session.OpenTx()

  if err != nil {
    this.Log("***************************************************")
    this.Log("***************************************************")
    this.Log("***** erro ao iniciar conexão com banco de dados: %v", err)
    this.Log("***************************************************")
    this.Log("***************************************************")

    this.Abort("505")
    return
  }

  this.FlashRead()

  this.EntityValidator = validator.NewEntityValidator(this.Lang, this.ViewPath)

  //this.Log("use default time location America/Sao_Paulo")
  this.DefaultLocation, _ = time.LoadLocation("America/Sao_Paulo")

  this.defaultPageLimit = 25
}

func (this *BaseController) DisableXSRF(pathList []string) {

  for _, url := range pathList {
    if strings.HasPrefix(this.Ctx.Input.URL(), url) {
      this.EnableXSRF = false
    }
  }

}

func (this *BaseController) FlashRead() {
  Flash := beego.ReadFromRequest(&this.Controller)

  if n, ok := Flash.Data["notice"]; ok {
    this.Flash.Notice(n)
  }

  if n, ok := Flash.Data["error"]; ok {
    this.Flash.Error(n)
  }

  if n, ok := Flash.Data["warning"]; ok {
    this.Flash.Warning(n)
  }

  if n, ok := Flash.Data["success"]; ok {
    this.Flash.Success(n)
  }
}

func (this *BaseController) Finish() {

  this.Log("* Controller.Finish, Commit")

  this.Session.Close()

  if app, ok := this.AppController.(NestFinisher); ok {
    app.NestFinish()
  }
}

func (this *BaseController) Finally(){

  this.Log("* Controller.Finally, Rollback")

  if this.Session != nil {
    this.Session.OnError().Close()
  }
}

func (this *BaseController) Recover(info interface{}){
  /*
  this.Log("--------------- Controller.Recover ---------------")
  this.Log("INFO: %v", info)
  this.Log("STACKTRACE: %v", string(debug.Stack()))
  this.Log("--------------- Controller.Recover ---------------")
  */
  if app, ok := this.AppController.(NestRecover); ok {
    info := &RecoverInfo{ Error: fmt.Sprintf("%v", info), StackTrace: string(debug.Stack()) }
    app.NextOnRecover(info)
  }

  
}

func (this *BaseController) Rollback() {
  if this.Session != nil {
    this.Session.OnError()
  }
}

func (this *BaseController) OnEntity(viewName string, entity interface{}) {
  this.Data["entity"] = entity
  this.OnTemplate(viewName)
  this.OnFlash(false)
}

func (this *BaseController) OnEntityError(viewName string, entity interface{}, message string) {
  this.Rollback()
  this.Flash.Error(message)
  this.Data["entity"] = entity
  this.OnTemplate(viewName)
  this.OnFlash(false)
}

func (this *BaseController) OnEntities(viewName string, entities interface{}) {
  this.Data["entities"] = entities
  this.OnTemplate(viewName)
  this.OnFlash(false)
}

func (this *BaseController) OnEntitiesWithTotalCount(viewName string, entities interface{}, totalCount int64) {
  this.Data["entities"] = entities
  this.Data["totalCount"] = totalCount
  this.OnTemplate(viewName)
  this.OnFlash(false)
}

func (this *BaseController) OnResult(viewName string, result interface{}) {
  this.Data["result"] = result
  this.OnTemplate(viewName)
  this.OnFlash(false)
}

func (this *BaseController) OnResults(viewName string, results interface{}) {
  this.Data["results"] = results
  this.OnTemplate(viewName)
  this.OnFlash(false)
}

func (this *BaseController) OnResultsWithTotalCount(viewName string, results interface{}, totalCount int64) {
  this.Data["results"] = results
  this.Data["totalCount"] = totalCount
  this.OnTemplate(viewName)
  this.OnFlash(false)
}

func (this *BaseController) OnJsonResult(result interface{}) {
  this.Data["json"] = support.JsonResult{ Result: result, Error: false, CurrentUnixTime: this.GetCurrentTimeUnix() }
  this.ServeJSON()
}

func (this *BaseController) OnJsonMessage(message string) {
  this.Data["json"] = support.JsonResult{ Message: message, Error: false, CurrentUnixTime: this.GetCurrentTimeUnix() }
  this.ServeJSON()
}

func (this *BaseController) OnJsonResultError(result interface{}, message string) {
  this.Rollback()
  this.Data["json"] = support.JsonResult{ Result: result, Message: message, Error: true, CurrentUnixTime: this.GetCurrentTimeUnix() }
  this.ServeJSON()
}

func (this *BaseController) OnJsonResultWithMessage(result interface{}, message string) {
  this.Data["json"] = support.JsonResult{ Result: result, Error: false, Message: message, CurrentUnixTime: this.GetCurrentTimeUnix() }
  this.ServeJSON()
}

func (this *BaseController) OnJsonResults(results interface{}) {
  this.Data["json"] = support.JsonResult{ Results: results, Error: false, CurrentUnixTime: this.GetCurrentTimeUnix() }
  this.ServeJSON()
}

func (this *BaseController) OnJsonResultAndResults(result interface{}, results interface{}) {
  this.Data["json"] = support.JsonResult{ Result: result, Results: results, Error: false, CurrentUnixTime: this.GetCurrentTimeUnix() }
  this.ServeJSON()
}

func (this *BaseController) OnJsonResultsWithTotalCount(results interface{}, totalCount int64) {
  this.Data["json"] = support.JsonResult{ Results: results, Error: false, CurrentUnixTime: this.GetCurrentTimeUnix(), TotalCount: totalCount }
  this.ServeJSON()
}

func (this *BaseController) OnJsonResultAndResultsWithTotalCount(result interface{}, results interface{}, totalCount int64) {
  this.Data["json"] = support.JsonResult{ Result: result, Results: results, Error: false, CurrentUnixTime: this.GetCurrentTimeUnix(), TotalCount: totalCount }
  this.ServeJSON()
}

func (this *BaseController) OnJsonResultsError(results interface{}, message string) {
  this.Rollback()
  this.Data["json"] = support.JsonResult{ Results: results, Message: message, Error: true, CurrentUnixTime: this.GetCurrentTimeUnix() }
  this.ServeJSON()
}

func (this *BaseController) OnJson(json support.JsonResult) {
  this.Data["json"] = json
  this.ServeJSON()
}

func (this *BaseController) OnJsonMap(jsonMap map[string]interface{}) {
  this.Data["json"] = jsonMap
  this.ServeJSON()
}

func (this *BaseController) OnJsonError(message string) {
  this.Rollback()
  this.OnJson(support.JsonResult{ Message: message, Error: true, CurrentUnixTime: this.GetCurrentTimeUnix() })
}

func (this *BaseController) OnJsonErrorNotRollback(message string) {  
  this.OnJson(support.JsonResult{ Message: message, Error: true, CurrentUnixTime: this.GetCurrentTimeUnix() })
}

func (this *BaseController) OnJsonOk(message string) {
  this.OnJson(support.JsonResult{ Message: message, Error: false, CurrentUnixTime: this.GetCurrentTimeUnix() })
}

func (this *BaseController) OnJson200() {
  this.OnJson(support.JsonResult{ CurrentUnixTime: this.GetCurrentTimeUnix() })
}

func (this *BaseController) OkAsJson(message string) {
  this.OnJson(support.JsonResult{ CurrentUnixTime: this.GetCurrentTimeUnix(), Message: message })
}

func (this *BaseController) OkAsHtml(message string) {
  this.Ctx.Output.Body([]byte(message))
}

func (this *BaseController) OnJsonValidationError() {
  this.Rollback()
  errors := this.Data["errors"].(map[string]string)
  this.OnJson(support.JsonResult{  Message: this.GetMessage("cadastros.validacao"), Error: true, Errors: errors, CurrentUnixTime: this.GetCurrentTimeUnix() })
}

func (this *BaseController) OnJsonValidationWithErrors(errors map[string]string) {
  this.Rollback()  
  this.OnJson(support.JsonResult{  Message: this.GetMessage("cadastros.validacao"), Error: true, Errors: errors, CurrentUnixTime: this.GetCurrentTimeUnix() })
}

func (this *BaseController) OnTemplate(viewName string) {
  this.TplName = fmt.Sprintf("%s/%s.tpl", this.ViewPath, viewName)
  this.OnFlash(false)
}

func (this *BaseController) OnPureTemplate(templateName string) {
  this.TplName = templateName
  this.OnFlash(false)
}

func (this *BaseController) OnRedirect(action string) {
  this.OnFlash(true)
  if this.Ctx.Input.URL() == "action" {
    this.Abort("500")
  } else {
    this.Redirect(action, 302)
  }
}

func (this *BaseController) OnRedirectError(action string, message string) {
  this.Rollback()
  this.Flash.Error(message)
  this.OnFlash(true)
  if this.Ctx.Input.URL() == "action" {
    this.Abort("500")
  } else {
    this.Redirect(action, 302)
  }}

func (this *BaseController) OnRedirectSuccess(action string, message string) {
  this.Flash.Success(message)
  this.OnFlash(true)
  if this.Ctx.Input.URL() == "action" {
    this.Abort("500")
  } else {
    this.Redirect(action, 302)
  }
}

// executes redirect or OnJsonError
func (this *BaseController) OnErrorAny(path string, message string) {

  //this.Log("** this.IsJson() %v", this.IsJson() )

  if this.IsJson() {
    this.OnJsonError(message)
  } else {
    this.OnRedirectError(path, message)
  }
}

// executes redirect or OnJsonOk
func (this *BaseController) OnOkAny(path string, message string) {

  if this.IsJson() {
    this.OnJsonOk(message)
  } else {
    this.Flash.Success(message)
    this.OnRedirect(path)
  }

}

// executes OnEntity or OnJsonValidationError
func (this *BaseController) OnValidationErrorAny(view string, entity interface{}) {

  if this.IsJson() {
    this.OnJsonValidationError()
  } else {
    this.Rollback()
    this.OnEntity(view, entity)
  }

}

// executes OnEntity or OnJsonError
func (this *BaseController) OnEntityErrorAny(view string, entity interface{}, message string) {

  if this.IsJson() {
    this.OnJsonError(message)
  } else {
    this.Rollback()
    this.Flash.Error(message)
    this.OnEntity(view, entity)
  }

}

// executes OnEntity or OnJsonResultWithMessage
func (this *BaseController) OnEntityAny(view string, entity interface{}, message string) {

  if this.IsJson() {
    this.OnJsonResultWithMessage(entity, message)
  } else {
    this.Flash.Success(message)
    this.OnEntity(view, entity)
  }

}

// executes OnResults or OnJsonResults
func (this *BaseController) OnResultsAny(viewName string, results interface{}) {

  if this.IsJson() {
    this.OnJsonResults(results)
  } else {
    this.OnResults(viewName, results)
  }

}

// executes  OnResultsWithTotalCount or OnJsonResultsWithTotalCount
func (this *BaseController) OnResultsWithTotalCountAny(viewName string, results interface{}, totalCount int64) {

  if this.IsJson() {
    this.OnJsonResultsWithTotalCount(results, totalCount)
  } else {
    this.OnResultsWithTotalCount(viewName, results, totalCount)
  }

}

func (this *BaseController) OnFlash(store bool) {
  if store {
    this.Flash.Store(&this.Controller)
  } else {
    this.Data["Flash"] = this.Flash.Data
    this.Data["flash"] = this.Flash.Data
  }
}

func (this *BaseController) GetMessage(key string, args ...interface{}) string{
  return i18n.Tr(this.Lang, key, args)
}

func (this *BaseController) OnValidate(entity interface{}, custonValidation func(validator *validation.Validation)) bool {

  result, _ := this.EntityValidator.IsValid(entity, custonValidation)

  if result.HasError {
    this.Flash.Error(this.GetMessage("cadastros.validacao"))
    this.EntityValidator.CopyErrorsToView(result, this.Data)
  }

  return result.HasError == false
}

func (this *BaseController) OnParseForm(entity interface{}) {
  if err := this.ParseForm(entity); err != nil {
    this.Log("*******************************************")
    this.Log("***** ERROR on parse form ", err.Error())
    this.Log("*******************************************")
    this.Abort("500")
  }
}

func (this *BaseController) OnJsonParseForm(entity interface{}) {
  this.OnJsonParseFormWithFieldsConfigs(entity, nil)
}

func (this *BaseController) OnJsonParseFormWithFieldsConfigs(entity interface{}, configs map[string]string) {
  if err := this.FormToModelWithFieldsConfigs(this.Ctx, entity, configs)  ; err != nil {
    this.Log("*******************************************")
    this.Log("***** ERROR on parse form ", err.Error())
    this.Log("*******************************************")
    this.Abort("500")
  }
}

func (this *BaseController) ParamParseMoney(s string) float64{
  return this.ParamParseFloat(s)
}

// remove ,(virgula) do valor em params que vem como val de input com jquery money
// exemplo 45,000.00 vira 45000.00
func (this *BaseController) ParamParseFloat(s string) float64{
  var semic string = ","
  replaced := strings.Replace(s, semic, "", -1) // troca , por espaço
  precoFloat, err := strconv.ParseFloat(replaced, 64)
  var returnValue float64
  if err == nil {
    returnValue = precoFloat
  }else{
    this.Log("*******************************************")
    this.Log("****** ERROR parse string to float64 for stringv", s)
    this.Log("*******************************************")
    this.Abort("500")
  }

  return returnValue
}

func (this *BaseController) OnParseJson(entity interface{}) {
  if err := this.JsonToModel(this.Ctx, entity); err != nil {
    this.Log("*******************************************")
    this.Log("***** ERROR on parse json ", err.Error())
    this.Log("*******************************************")
    this.Abort("500")
  }
}

func (this *BaseController) HasPath(paths ...string) bool{
  for _, it := range paths {
    if strings.HasPrefix(this.Ctx.Input.URL(), it){
      return true
    }
  }
  return false
}

func (this *BaseController) IsJson() bool{
  return  this.Ctx.Input.AcceptsJSON()
}

func (this *BaseController) IsAjax() bool{
  return  this.Ctx.Input.IsAjax()
}

func (this *BaseController) GetToken() string{
  return this.GetHeaderByName("X-Auth-Token")
}

func (this *BaseController) GetHeaderByName(name string) string{
  return this.Ctx.Request.Header.Get(name)
}

func (this *BaseController) GetHeaderByNames(names ...string) string{

  for _, name := range names {
    val := this.Ctx.Request.Header.Get(name)

    if len(val) > 0 {
      return val
    }
  }

  return ""
}

func (this *BaseController) Log(format string, v ...interface{}) {
 logs.Debug(fmt.Sprintf(format, v...))
}

func (this *BaseController) GetCurrentTimeUnix() int64 {
  return this.GetCurrentTime().Unix()
}

func (this *BaseController) GetCurrentTime() time.Time {
  return time.Now().In(this.DefaultLocation)
}

func (this *BaseController) GetPage() *db.Page{
  page := new(db.Page)

  if this.IsJson() {

    if this.Ctx.Input.IsPost() {
      jsonMap, _ := this.JsonToMap(this.Ctx)

      //this.Log(" page jsonMap = %v", jsonMap)

      if _, ok := jsonMap["limit"]; ok {
        page.Limit = this.GetJsonInt64(jsonMap, "limit")
        page.Offset = this.GetJsonInt64(jsonMap, "offset")
        page.Sort = this.GetJsonString(jsonMap, "order_column")
        page.Order = this.GetJsonString(jsonMap, "order_sort")
        page.Search = this.GetJsonString(jsonMap, "search")
      }
    } else {

        page.Limit = this.GetIntByKey("limit")
        page.Offset = this.GetIntByKey("offset")
        page.Sort = this.GetStringByKey("order_column")
        page.Order = this.GetStringByKey("order_sort")
        page.Search = this.GetStringByKey("search")

    }

  } else {

    page.Limit = this.GetIntByKey("limit")
    page.Offset = this.GetIntByKey("offset")
    page.Search = this.GetStringByKey("search")
    page.Order = this.GetStringByKey("order_sort")
    page.Sort = this.GetStringByKey("order_column")

  }

  if page.Limit <= 0 {
    page.Limit = this.defaultPageLimit
  }

  return page
}

func (this *BaseController) StringToInt(text string) int {
  val, _ := strconv.Atoi(text)
  return val
}

func (this *BaseController) StringToInt64(text string) int64 {
  val, _ := strconv.ParseInt(text, 10, 64)
  return val
}

func (this *BaseController) IntToString(val int) string {
  return fmt.Sprintf("%v", val)
}

func (this *BaseController) Int64ToString(val int64) string {
  return fmt.Sprintf("%v", val)
}


func (this *BaseController) GetId() int64 {
  return this.GetIntParam(":id")
}

func (this *BaseController) GetIntParam(key string) int64 {
  id := this.Ctx.Input.Param(key)
  intid, _ := strconv.ParseInt(id, 10, 64)
  return intid
}

func (this *BaseController) GetParam(key string) string {
  return this.Ctx.Input.Param(key)
}

func (this *BaseController) GetIntByKey(key string) int64{
  val := this.Ctx.Input.Query(key)
  intid, _ := strconv.ParseInt(val, 10, 64)
  return intid
}

func (this *BaseController) GetBoolByKey(key string) bool{
  val := this.Ctx.Input.Query(key)
  boolean, _ := strconv.ParseBool(val)
  return boolean
}

func (this *BaseController) GetStringByKey(key string) string{
  return this.Ctx.Input.Query(key)
}

func (this *BaseController) GetDateByKey(key string) (time.Time, error){
  date := this.Ctx.Input.Query(key)
  return this.ParseDate(date)
}

func (this *BaseController) ParseDateByKey(key string, layout string) (time.Time, error){
  date := this.Ctx.Input.Query(key)
  return time.ParseInLocation(layout, date, this.DefaultLocation)
}

// deprecated
func (this *BaseController) ParseDate(date string) (time.Time, error){
  return time.ParseInLocation(dateLayout, date, this.DefaultLocation)
}

// deprecated
func (this *BaseController) ParseDateTime(date string) (time.Time, error){
  return time.ParseInLocation(datetimeLayout, date, this.DefaultLocation)
}

// deprecated
func (this *BaseController) ParseJsonDate(date string) (time.Time, error){
  return time.ParseInLocation(jsonDateLayout, date, this.DefaultLocation)
}
