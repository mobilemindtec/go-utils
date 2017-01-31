package validator

import (  
  "github.com/astaxie/beego/validation"
  "github.com/beego/i18n"
  "fmt" 
)

type EntityValidatorResult struct {
  Errors map[string]string
  ErrorsFields map[string]string
  HasError bool
}

type EntityValidator struct {
  Lang string
  ViewPath string
}

func NewEntityValidator(lang string, viewPath string) *EntityValidator{
  return &EntityValidator{ Lang: lang, ViewPath: viewPath }
}

func (this *EntityValidator) IsValid(entity interface{}, action func(validator *validation.Validation)) (*EntityValidatorResult, error) {
  
  result := new(EntityValidatorResult)

  localValid := validation.Validation{}
  callerValid := validation.Validation{}
  result.Errors = make(map[string]string)
  result.ErrorsFields = make(map[string]string)

  ok, err := localValid.Valid(entity)

  if  err != nil {    
    fmt.Println("## error on run validation %v", err.Error())
    return nil, err
  }

  if action != nil {
    action(&callerValid)    
  }

  if !ok {
    for _, err := range localValid.Errors {    

      label := this.GetMessage(fmt.Sprintf("%s.%s", this.ViewPath, err.Field))
      
      if label != "" {
        result.Errors[label] = err.Message      
      }else{
        result.Errors[err.Field] = err.Message      
      }

      result.ErrorsFields[err.Field] = err.Message

      fmt.Println("## validator error field %v error %v", err.Field, err)
    }

    result.HasError = true
  }

  if callerValid.HasErrors() {    
    for _, err := range callerValid.Errors {

      label := this.GetMessage(fmt.Sprintf("%s.%s", this.ViewPath, err.Field))
      
      if label != "" {
        result.Errors[label] = err.Message      
      }else{
        result.Errors[err.Field] = err.Message      
      }

      result.ErrorsFields[err.Field] = err.Message

      fmt.Println("## validator error field %v error %v", err.Field, err)  
    }

    result.HasError = true
  }

  return result, nil
}

func (this *EntityValidator) CopyErrorsToView(result *EntityValidatorResult, data map[interface{}]interface{}) {
  if len(result.Errors) > 0 {
    data["errors"] = result.Errors

    if data["errorsFields"] == nil {
      data["errorsFields"] = result.ErrorsFields      
    } else {
      mapItem := data["errorsFields"].(map[string]string)      
      for k, v := range result.ErrorsFields       {
        mapItem[k] = v
      }
    }
  }  
}

func (this *EntityValidator) GetMessage(key string, args ...interface{}) string{
  return i18n.Tr(this.Lang, key, args)
}