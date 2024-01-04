package session

import (
	_ "errors"

	"github.com/mobilemindtec/go-utils/v2/criteria"

	"fmt"
	"reflect"

	"github.com/mobilemindtec/go-utils/app/models"
	"github.com/mobilemindtec/go-utils/beego/db"
	"github.com/mobilemindtec/go-utils/v2/optional"
)

type Action struct {
	entity interface{}
}

func (this *Action) Get() interface{} {
	return this.entity
}

type ActionSave struct {
	Action
}

func NewActionSave(e interface{}) *ActionSave {
	return &ActionSave{Action{entity: e}}
}

type ActionUpdate struct {
	Action
}

func NewActionUpdate(e interface{}) *ActionUpdate {
	return &ActionUpdate{Action{entity: e}}
}

type ActionPersist struct {
	Action
}

func NewActionPersist(e interface{}) *ActionPersist {
	return &ActionPersist{Action{entity: e}}
}

type ActionRemove struct {
	Action
}

func NewActionRemove(e interface{}) *ActionRemove {
	return &ActionRemove{Action{entity: e}}
}

type ActionRemoveCascade struct {
	Action
}

func NewActionRemoveCascade(e interface{}) *ActionRemoveCascade {
	return &ActionRemoveCascade{Action{entity: e}}
}

type RxSession[T any] struct {
	session *db.Session
	actions []interface{}
	where   *criteria.Reactive
}

func WithTxOpt[T any]() *optional.Optional[*RxSession[T]] {
	s := db.NewSession()
	err := s.OpenTx()
	if err != nil {
		return optional.OfFail[*RxSession[T]](err)
	}
	val := &RxSession[T]{session: s, actions: []interface{}{}}
	return optional.Of[*RxSession[T]](val)
}

func WithNoTx[T any]() *RxSession[T] {
	s := db.NewSession()
	err := s.OpenNoTx()
	if err != nil {
		panic(err)
	}
	return New[T](s)
}

func WithTx[T any]() *RxSession[T] {
	s := db.NewSession()
	err := s.OpenTx()
	if err != nil {
		panic(err)
	}
	return New[T](s)
}

func WithNoTxOpt[T any]() *optional.Optional[*RxSession[T]] {
	s := db.NewSession()
	err := s.OpenNoTx()
	if err != nil {
		return optional.OfFail[*RxSession[T]](err)
	}
	val := &RxSession[T]{session: s, actions: []interface{}{}}
	return optional.Of[*RxSession[T]](val)
}

func New[T any](session *db.Session) *RxSession[T] {
	return &RxSession[T]{session: session, actions: []interface{}{}}
}

func (this *RxSession[T]) Close() {
	this.session.Close()
}

func (this *RxSession[T]) Where(c *criteria.Reactive) *RxSession[T] {
	this.where = c
	return this
}

func (this *RxSession[T]) AddAction(ac ...interface{}) *RxSession[T] {
	for _, it := range ac {
		this.actions = append(this.actions, it)
	}
	return this
}

func (this *RxSession[T]) RunWithTenantId(id int64, cb func(*RxSession[T]) T) T {
	var result T
	this.session.RunWithTenant(models.NewTenantWithId(id), func() {
		result = cb(this)
	})
	return result
}

func (this *RxSession[T]) AddPersist(items ...interface{}) *RxSession[T] {
	for _, o := range items {
		this.AddAction(NewActionPersist(o))
	}
	return this
}

func (this *RxSession[T]) AddPersistOf(items ...T) *RxSession[T] {
	for _, o := range items {
		this.AddPersist(o)
	}
	return this
}

func (this *RxSession[T]) AddSave(items ...interface{}) *RxSession[T] {
	for _, o := range items {
		this.AddAction(NewActionSave(o))
	}
	return this
}

func (this *RxSession[T]) AddSaveOf(items ...T) *RxSession[T] {
	for _, o := range items {
		this.AddSave(o)
	}
	return this
}

func (this *RxSession[T]) AddUpdate(items ...interface{}) *RxSession[T] {
	for _, o := range items {
		this.AddAction(NewActionUpdate(o))
	}
	return this
}

func (this *RxSession[T]) AddUpdateOf(items ...T) *RxSession[T] {
	for _, o := range items {
		this.AddUpdate(o)
	}
	return this
}

func (this *RxSession[T]) AddRemove(items ...interface{}) *RxSession[T] {
	for _, o := range items {
		this.AddAction(NewActionRemove(o))
	}
	return this
}

func (this *RxSession[T]) AddRemoveOf(items ...T) *RxSession[T] {
	for _, o := range items {
		this.AddRemove(o)
	}
	return this
}

func (this *RxSession[T]) AddRemoveCascade(items ...interface{}) *RxSession[T] {
	for _, o := range items {
		this.AddAction(NewActionRemoveCascade(o))
	}
	return this
}

func (this *RxSession[T]) AddRemoveCascadeOf(items ...T) *RxSession[T] {
	for _, o := range items {
		this.AddRemoveCascade(o)
	}
	return this
}

func (this *RxSession[T]) Exec() *optional.Optional[bool] {
	r := this.Run()
	switch r.Val().(type) {
	case *optional.Some:
		return optional.Of[bool](true)
	}
	return optional.Of[bool](r.Val())
}

func (this *RxSession[T]) ExecWhere(c *criteria.Reactive) *optional.Optional[bool] {
	this.Where(c)
	return this.Exec()
}

func (this *RxSession[T]) Run() *optional.Optional[T] {

	if this.where != nil {
		first := this.where.Any()
		r := optional.Of[bool](first.Get())

		if r.IsFail() {
			return optional.OfFail[T](r.GetFail())
		}

		if r.UnWrap() {
			return optional.OfOk[T]()
		}
	}

	for _, ac := range this.actions {

		var err error

		switch ac.(type) {
		case *ActionSave:
			err = this.session.Save(ac.(*ActionSave).Get())
			break
		case *ActionUpdate:
			err = this.session.Update(ac.(*ActionUpdate).Get())
			break
		case *ActionPersist:
			err = this.session.SaveOrUpdate(ac.(*ActionPersist).Get())
			break
		case *ActionRemove:
			err = this.session.Remove(ac.(*ActionRemove).Get())
			break
		case *ActionRemoveCascade:
			err = this.session.RemoveCascade(ac.(*ActionRemoveCascade).Get())
			break
		default:
			err = fmt.Errorf("invalid action: %v", reflect.TypeOf(ac))
		}

		if err != nil {
			return optional.OfFail[T](err)
		}
	}

	return optional.OfOk[T]()
}

func (this *RxSession[T]) Save(entity T) *optional.Optional[T] {

	if err := this.session.Save(entity); err != nil {
		return optional.OfFail[T](err)
	}
	return optional.OfSome[T](entity)
}

func (this *RxSession[T]) SaveCascade(entity T) *optional.Optional[T] {

	if err := this.session.SaveCascade(entity); err != nil {
		return optional.OfFail[T](err)
	}
	return optional.OfSome[T](entity)
}

func (this *RxSession[T]) Update(entity T) *optional.Optional[T] {

	if err := this.session.Update(entity); err != nil {
		return optional.OfFail[T](err)
	}
	return optional.OfSome[T](entity)
}

func (this *RxSession[T]) UpdateCascade(entity T) *optional.Optional[T] {

	if err := this.session.Update(entity); err != nil {
		return optional.OfFail[T](err)
	}
	return optional.OfSome[T](entity)
}

func (this *RxSession[T]) Remove(entity T) *optional.Optional[bool] {

	if err := this.session.Remove(entity); err != nil {
		return optional.OfFail[bool](err)
	}
	return optional.OfSome[bool](true)
}

func (this *RxSession[T]) RemoveCascade(entity T) *optional.Optional[bool] {

	if err := this.session.RemoveCascade(entity); err != nil {
		return optional.OfFail[bool](err)
	}
	return optional.OfSome[bool](true)
}

func (this *RxSession[T]) Persist(entity T) *optional.Optional[T] {

	if err := this.session.SaveOrUpdate(entity); err != nil {
		return optional.OfFail[T](err)
	}
	return optional.OfSome[T](entity)
}

func (this *RxSession[T]) SaveWhere(entity T, c *criteria.Reactive) *optional.Optional[T] {

	first := c.First()
	r := optional.Of[T](first.Get())

	if r.IsFail() {
		return r
	}

	if r.IsNone() {
		return this.Save(entity)
	}

	return r
}
