package support 

import (
  "github.com/astaxie/beego/context"
  "github.com/leekchan/accounting"
  "strings"
  "regexp"
  "math"
  "sort"
  "time"
)


func FilterNumber(text string) string{
    re := regexp.MustCompile("[0-9]+")
    result := re.FindAllString(text, -1)
    number := ""
    for _, s := range result {
        number += s
    }

    return number
}

func IsEmpty(text string) bool{
  return len(strings.TrimSpace(text)) == 0
}

func MakeRange(min, max int) []int {
    a := make([]int, max-min+1)
    for i := range a {
        a[i] = min + i
    }
    return a
}

func SliceCopyAndSortOfStrings(arr []string) []string {
    tmpArr := make([]string, len(arr))
    copy(tmpArr, arr)
    sort.Strings(tmpArr)
    return tmpArr
}

func SliceIndex(limit int, predicate func(i int) bool) int {
    for i := 0; i < limit; i++ {
        if predicate(i) {
            return i
        }
    }
    return -1
}

// troca , por .(ponto), posi alterei o js maskMoney pra #.###,##
func NormalizeSemicolon(key string, ctx *context.Context) {
    if _, ok := ctx.Request.Form[key]; ok {
        ctx.Request.Form[key][0] = strings.Replace(ctx.Request.Form[key][0], ",", "", -1)
    }
}

func RemoveAllSemicolon(key string, ctx *context.Context) {
    if _, ok := ctx.Request.Form[key]; ok {
        ctx.Request.Form[key][0] = strings.Replace(ctx.Request.Form[key][0], ",", "", -1)
    }
}

func SetFormDefault(key string, defVal string, ctx *context.Context) {
    if _, ok := ctx.Request.Form[key]; ok {

      val := ctx.Request.Form[key][0]

      if len(strings.TrimSpace(val)) == 0 {
        ctx.Request.Form[key][0] = defVal
      }
      
    }
}

func FormatMoney(number float64) string{
  ac := accounting.Accounting{Symbol: "R$ ", Precision: 2, Thousand: ",", Decimal: "."}
  return ac.FormatMoney(number)
}


func ToFixed(num float64, precision int) float64 {
  output := math.Pow(10, float64(precision))
  return float64(Round(num * output)) / output
}

func Round(num float64) int {
  return int(num + math.Copysign(0.5, num))
}

func Reverse(s string) string {
  runes := []rune(s)
  for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
      runes[i], runes[j] = runes[j], runes[i]
  }
  return string(runes)
}

func NumberMask(text string, maskApply string) string{

  re := regexp.MustCompile("[0-9]+")
  results := re.FindAllString(text, -1)
  text = strings.Join(results[:],",")

  var newText string
  var j int 

  for i:= 0; i < len(maskApply); i++ {
    
    m := maskApply[i]

    if j >= len(text) {
      newText += string(m)
      continue
    }

    c := text[j]

    if re.MatchString(string(c)) {
      if re.MatchString(string(m)) {
        newText += string(c)
        j++
      } else {
        newText += string(m)
      }
    }
  }

  return newText
}

func DateToTheEndOfDay(timeArg time.Time) time.Time {
	returnTime := timeArg.Local().Add(time.Hour * time.Duration(23) +
	                                 time.Minute * time.Duration(59) +
	                                 time.Second * time.Duration(59))
	return returnTime
}

func NumberMaskReverse(text string, maskApply string) string{

  re := regexp.MustCompile("[0-9]+")
  results := re.FindAllString(text, -1)
  text = strings.Join(results[:],",")
  text = Reverse(text)

  var newText string
  var j int 

  for i:= len(maskApply)-1; i >= 0; i-- {
    
    m := maskApply[i]

    if j >= len(text) {
      newText += string(m)
      continue
    }

    c := text[j]

    if re.MatchString(string(c)) {
      if re.MatchString(string(m)) {
        newText += string(c)
        j++
      } else {
        newText += string(m)
      }
    }
  }

  return Reverse(newText)
}