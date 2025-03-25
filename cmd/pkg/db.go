package main

import (
	"net/http"
	"time"
)

type model struct {
	data  map[int]*DataModel
	limit int
}

func (m *model) add(id int) {
	m.data[id] = &DataModel{
		Count:     1,
		ExpiresAt: time.Now().Add(time.Minute),
	}
}

func (m *model) update(id int) {
	data, ok := m.data[id]
	if !ok {
		m.add(id)
		return
	}

	data.Count += data.Count
}

func (m *model) refresh(id int) {
	data, ok := m.data[id]
	if !ok {
		m.add(id)
		return
	}

	data.Count = 0
	data.ExpiresAt = time.Now()
}

// 0 - limit reached before expire time - not ok
// 1 - limit not reached and count needs to be updated - ok
// 2 - limit reached but request made at expire time - not ok but count needs to be refreshed
// 3 - limit not reached but count needs to be refreshed - ok
func (m *model) check(data *DataModel) int {
	timeDiff := time.Now().Unix() - data.ExpiresAt.Unix()

	if timeDiff < 0 {
		if data.Count > m.limit {
			return 0
		}
		return 1
	}

	if timeDiff == 0 {
		if data.Count > m.limit {
			return 2
		}

		return 3
	}

	if data.Count > m.limit {
		return 3
	}

	return 1

}

func (m *model) Get(id int) int {
	data, ok := m.data[id]
	if !ok {
		m.add(id)
		return http.StatusOK
	}

	result := m.check(data)
	switch result {
	case 0:
		return http.StatusTooManyRequests

	case 1:
		m.update(id)
		return http.StatusOK

	case 2:
		m.refresh(id)
		return http.StatusTooManyRequests

	case 3:
		m.refresh(id)
		return http.StatusOK

	default:
		return http.StatusInternalServerError
	}
}

func NewModel(limit int) *model {
	return &model{
		data:  make(map[int]*DataModel),
		limit: limit,
	}
}
