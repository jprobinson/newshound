package fetch

import (
	"context"
	"database/sql"
	"time"

	"github.com/jprobinson/newshound"
)

type DB interface {
	PutAlert(context.Context, *newshound.NewsAlert) error
	FindAlertsByTimeframe(context.Context, time.Time, time.Time) ([]*newshound.NewsAlert, error)
	FindPossibleLikeAlerts(context.Context, *newshound.NewsAlert) ([]*newshound.NewsAlert, error)
	FindEventsByAlertIDs(context.Context, []int64) ([]*newshound.NewsEvent, error)
	GetAlertsByID(context.Context, []int64) ([]*newshound.NewsAlert, error)
	UpsertEvent(context.Context, *newshound.NewsEvent) error
	DeleteEvents(context.Context, []int64) error
	GetAllAlerts(context.Context) (<-chan *newshound.NewsAlert, error)
}

type pq struct {
	db *sql.DB
}

func NewDB(db *sql.DB) DB {
	return &pq{db}
}

func (p *pq) PutAlert(context.Context, *newshound.NewsAlert) error {
	panic("not implemented")
	return nil
}

func (p *pq) FindAlertsByTimeframe(context.Context, time.Time, time.Time) ([]*newshound.NewsAlert, error) {
	panic("not implemented")
	return nil, nil
}

func (p *pq) FindPossibleLikeAlerts(context.Context, *newshound.NewsAlert) ([]*newshound.NewsAlert, error) {
	//	start := a.Timestamp.Add(-eventTimeframe)
	//	end := a.Timestamp.Add(eventTimeframe)
	//	query := bson.M{"timestamp": bson.M{"$gte": start, "$lte": end}, "_id": bson.M{"$ne": a.ID}, "tags": bson.M{"$in": a.Tags}}

	panic("not implemented")
	return nil, nil
}

func (p *pq) FindEventsByAlertIDs(context.Context, []int64) ([]*newshound.NewsEvent, error) {
	//	var existingEvents []newshound.NewsEvent
	//	query := bson.M{"news_alerts.alert_id": bson.M{"$in": cluster}}
	//	err = ne.Find(query).All(&existingEvents)

	panic("not implemented")
	return nil, nil
}

func (p *pq) GetAlertsByID(context.Context, []int64) ([]*newshound.NewsAlert, error) {
	panic("not implemented")
	return nil, nil
}

func (p *pq) UpsertEvent(context.Context, *newshound.NewsEvent) error {
	panic("not implemented")
	return nil
}

func (p *pq) DeleteEvents(context.Context, []int64) error {
	panic("not implemented")
	return nil
}

func (p *pq) GetAllAlerts(context.Context) (<-chan *newshound.NewsAlert, error) {
	panic("not implemented")
	return nil, nil
}
