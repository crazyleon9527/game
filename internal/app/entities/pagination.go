package entities

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type Paginator struct {
	Page     int         `json:"page" form:"page,default=1"`
	PageSize int         `json:"pageSize" form:"pageSize,default=30"`
	Count    int64       `json:"count" `
	List     interface{} `json:"list" `
}

func NewPaginator(page, pageSize int, list interface{}) *Paginator {
	return &Paginator{
		Page:     page,
		PageSize: pageSize,
		List:     list,
	}
}

func (p *Paginator) Paginate(tx *gorm.DB) error {
	var err error
	err = tx.Model(p.List).Count(&p.Count).Error
	if err != nil {
		return err
	}

	if p.PageSize == 0 {
		err := tx.Find(&p.List).Error
		return err
	}

	if p.Page <= 0 {
		p.Page = 1
	}

	offset := (p.Page - 1) * p.PageSize

	err = tx.Offset(offset).Limit(p.PageSize).Find(&p.List).Error
	if err != nil {
		return err
	}
	return nil
}

type QueryPaginatorList struct {
	Paginator
	OrderFields  []*OrderField     `form:"orders"`
	SelectFields []string          `form:"fields"`
	Conditions   map[string]string `form:"conditions"`
}

func (p *QueryPaginatorList) Paginate(tx *gorm.DB) error {
	if p.SelectFields != nil && len(p.SelectFields) > 0 {
		tx = tx.Select(p.SelectFields)
	}

	if p.OrderFields != nil && len(p.OrderFields) > 0 {
		tx = tx.Order(p.ParseOrder(p.OrderFields))
	}
	if len(p.Conditions) != 0 {
		for k, v := range p.Conditions {
			if len(v) > 0 { //temp
				tx = tx.Where(k, v)
			}
		}
	}

	if p.PageSize == 0 {
		err := tx.Find(&p.List).Error
		return err
	}
	tx = tx.Debug()

	if p.Page <= 0 {
		p.Page = 1
	}
	offset := (p.Page - 1) * p.PageSize
	var err error

	err = tx.Model(p.List).Count(&p.Count).Error
	if err != nil {
		return err
	}
	err = tx.Offset(offset).Limit(p.PageSize).Find(&p.List).Error
	if err != nil {
		return err
	}
	// log.Println("list", p.List)
	return nil
}

func (q *QueryPaginatorList) ParseOrder(items []*OrderField, handle ...OrderFieldFunc) string {
	orders := make([]string, len(items))

	for i, item := range items {
		key := item.Key
		if len(handle) > 0 {
			key = handle[0](key)
		}

		direction := "ASC"
		if item.Direction == OrderByDESC {
			direction = "DESC"
		}
		orders[i] = fmt.Sprintf("%s %s", key, direction)
	}

	return strings.Join(orders, ",")
}

func (q *QueryPaginatorList) AddOrderField(key string, d OrderDirection) {
	if q.OrderFields == nil {
		q.OrderFields = make([]*OrderField, 0)
	}
	q.OrderFields = append(q.OrderFields, NewOrderField(key, d))
}

func (q *QueryPaginatorList) AddSelects(fieldNames ...string) {
	if q.SelectFields == nil {
		q.SelectFields = make([]string, 0)
	}
	q.SelectFields = append(q.SelectFields, fieldNames...)
}

func (q *QueryPaginatorList) AddCondition(key string, val string) {
	if q.Conditions == nil {
		q.Conditions = make(map[string]string)
	}
	q.Conditions[key] = val
}

type OrderDirection int

const (
	OrderByASC OrderDirection = iota
	OrderByDESC
)

func NewOrderFields(orderFields ...*OrderField) []*OrderField {
	return orderFields
}

func NewOrderField(key string, d OrderDirection) *OrderField {
	return &OrderField{
		Key:       key,
		Direction: d,
	}
}

type OrderField struct {
	Key       string
	Direction OrderDirection
}

type OrderFieldFunc func(string) string
