package optional

import (
	"reflect"
)

type Optional[T any] struct {
	some *Some
	none *None
	fail *Fail
} 

func WithSome[T any](v interface{}) *Optional[T] {

	var s *Some

	switch v.(type) {
	case *Some:
		s = v.(*Some)
		break
	default:
		s = NewSome(v) 
	}

	return &Optional[T]{ some: s }
}

func WithFail[T any](v interface{}) *Optional[T] {

	var s *Fail

	switch v.(type) {
	case *Fail:
		s = v.(*Fail)
		break
	case error:	
		s = NewFail(v.(error)) 
	}

	return &Optional[T]{ fail: s }
}

func WithNone[T any]() *Optional[T] {
	return &Optional[T]{ none: NewNone() }
}

func New[T any](val interface{}) *Optional[T]{

	opt := Optional[T]{}

	if IsNilFixed(val) {
		opt.none = NewNone()
		return &opt
	}

	switch val.(type) {
		case *Some:
			opt.some = val.(*Some)
			break
		case *None:
			opt.none = val.(*None)
			break
		case *Fail:
			opt.fail = val.(*Fail)
			break
		case error:
			opt.fail = NewFail(val.(error))
			break
		default:
			opt.some = NewSome(val)
			break
	}
	return &opt
}

func (this *Optional[T]) OrElse(v T) T {
	return GetOrDefault[T](this.some, v)
} 

func (this *Optional[T]) Get(v T) T {
	return Get[T](this.some)
} 

func (this *Optional[T]) Any() bool {
	return this.some != nil
}

func (this *Optional[T]) Fail() bool {
	return this.fail != nil
}

func (this *Optional[T]) Empty() bool {
	return this.none != nil
}

func (this *Optional[T]) NonEmpty() bool {
	return this.none == nil && this.fail == nil
}


func (this *Optional[T]) Val() interface{} {
	if this.some != nil {
		return this.some
	} else if this.none != nil {
		return this.none
	}else if this.fail != nil {
		return this.fail
	}else{
		return NewEmpty()
	}
}

func (this *Optional[T]) IfFail(cb func(error)) *Optional[T]{
	if this.fail != nil {
		cb(this.fail.Error)
	}
	return this
} 

func (this *Optional[T]) IfSome(cb func(T)) *Optional[T]{
	if this.some != nil {
		cb(GetItem[T](this.some))
	}
	return this
} 

func (this *Optional[T]) IfNone(cb func()) *Optional[T]{
	if this.none != nil {
		cb()
	}
	return this
} 

func (this *Optional[T]) IfEmpty(cb func()) *Optional[T]{
	this.IfNone(cb)
	return this
} 

func (this *Optional[T]) IfNonEmpty(cb func(T)) *Optional[T]{
	this.IfNonEmpty(cb)
	return this
} 

func (this *Optional[T]) Else(cb func()) *Optional[T]{
	cb()
	return this
} 

func Map[F any, T any](opt *Optional[F], fn func(*Optional[F]) T) T {
	return fn(opt)
}

type Empty struct {

}

func NewEmpty() *Empty {
	return &Empty{}
}

type None struct {
	
}

func NewNone() *None {
	return &None{}
}

type Some struct {	
	Item interface{}
}

func NewSome(item interface{}) *Some {
	return &Some{Item: item}
}

func NewSomeEmpty() *Some {
	return &Some{}
}

type Try struct {
	
}

type Fail struct {
	Error error
}

func (this *Fail) ErrorString() string {
	return this.Error.Error()
}


func NewFail(err error) *Fail {
	return &Fail{ Error: err }
}

type Success struct {
	Item interface{}
}

func (this *Success) WithItem(item interface{}) *Success{
	this.Item = item
	return this
}

func NewSuccess() *Success {
	return &Success{ }
}

type Either struct {
	
}

type Left struct {
		Item interface{}
}

func (this *Left) WithItem(item interface{}) *Left{
	this.Item = item
	return this
}

func NewLeft() *Left{
	return &Left{}
}

type Rigth struct {
	Item interface{}
}

func (this *Rigth) WithItem(item interface{}) *Rigth{
	this.Item = item
	return this
}

func NewRigth() *Rigth{
	return &Rigth{}
}

func Get[R any](val interface{}) R{
	return val.(R)
}

func GetOrDefault[R any](val interface{}, r R) R{
	if x, ok := val.(R); ok {
		return x
	}
	return r
}

func GetItem[R any](val interface{}) R{
	switch val.(type) {
		case *Some:
			return GetSome(val).Item.(R)
		case *Success:
			return GetSuccess(val).Item.(R)
		case *Left:
			return GetLeft(val).Item.(R)
		case *Rigth:
			return GetRigth(val).Item.(R)
		default: 
			var x R
			return x	
	}
}

func GetFail(val interface{}) *Fail{
	return val.(*Fail)
}


func GetSuccess(val interface{}) *Success{
	return val.(*Success)
}

func GetSome(val interface{}) *Some{
	return val.(*Some)
}

func GetLeft(val interface{}) *Left{
	return val.(*Left)
}

func GetRigth(val interface{}) *Rigth{
	return val.(*Rigth)
}

func GetFailError(val interface{}) error{
	return val.(*Fail).Error
}

func MakeSlice(val interface{}, err error) interface {} {

	if err != nil {
		return NewFail(err)
	}

	if val == nil || IsNilFixed(val) {
		return NewNone()
	}

	ss := reflect.ValueOf(val)    
  s := reflect.Indirect(ss)
  if s.Len() > 0 {
  	return NewSome(val)
  }
  return NewNone()
}

func Make(val interface{}, err error) interface{} {
	
	if err != nil {
		return NewFail(err)
	}

	if val == nil || IsNilFixed(val) {
		return NewNone()
	}

	switch val.(type) {
		case bool:
			if val.(bool) { 
				return NewSome(val) 
			} 
			return NewNone()
		case string:
			if val.(string) != "" { 
				return NewSome(val) 
			} 
			return NewNone()
		case int:
			if val.(int) != 0 { 
				return NewSome(val) 
			} 
			return NewNone()
		case int64:
			if val.(int64) != 0 { 
				return NewSome(val) 
			} 
			return NewNone()
		case float32:
			if val.(float32) != 0 { 
				return NewSome(val) 
			} 
			return NewNone()
		case float64:
			if val.(float64) != 0 { 
				return NewSome(val) 
			} 
			return NewNone()
		default:
			return NewSome(val)
	}

}

func IsSlice(v interface{}) bool {
    return reflect.TypeOf(v).Kind() == reflect.Slice || reflect.TypeOf(v).Kind() == reflect.Array
}

func IsNilFixed(i interface{}) bool {
   if i == nil {
      return true
   }
   switch reflect.TypeOf(i).Kind() {
   case reflect.Ptr, reflect.Map, reflect.Array, reflect.Chan, reflect.Slice:
      return reflect.ValueOf(i).IsNil()
   }
   return false
}

func IsSimple(v interface{}) bool {
	switch v.(type) {
		case int, int64, float32, float64, bool, string:
			return true		
	}
	return false
}