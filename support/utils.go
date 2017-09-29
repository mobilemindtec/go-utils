package support 

import (
	"github.com/astaxie/beego/context"
	"regexp"
	"strings"
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

