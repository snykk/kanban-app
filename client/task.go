package client

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/snykk/kanban-app/config"
	"github.com/snykk/kanban-app/entity"
)

type TaskClient interface {
	CreateTask(title, description, category, userID string) (respCode int, err error)
	GetTaskById(id, userID string) (entity.Task, error)
	UpdateTask(id, title, description, userID string) (respCode int, err error)
	UpdateCategoryTask(id, catId, userID string) (respCode int, err error)
	DeleteTask(id, userID string) (respCode int, err error)
}

type taskClient struct {
}

func NewTaskClient() *taskClient {
	return &taskClient{}
}

func (t *taskClient) CreateTask(title, description, category, userID string) (respCode int, err error) {
	client, err := GetClientWithCookie(userID)
	if err != nil {
		return -1, err
	}

	catId, err := strconv.Atoi(category)
	if err != nil {
		return -1, err
	}

	datajson := map[string]interface{}{
		"title":       title,
		"description": description,
		"category_id": int(catId),
	}

	b, err := json.Marshal(datajson)
	if err != nil {
		return -1, err
	}

	req, err := http.NewRequest("POST", config.SetUrl("/api/v1/tasks/create"), bytes.NewBuffer(b))
	if err != nil {
		return -1, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return -1, err
	}

	defer resp.Body.Close()

	return resp.StatusCode, nil
}

func (t *taskClient) GetTaskById(id, userID string) (entity.Task, error) {
	client, err := GetClientWithCookie(userID)
	if err != nil {
		return entity.Task{}, err
	}

	req, err := http.NewRequest("GET", config.SetUrl("/api/v1/tasks/get?task_id="+id), nil)
	if err != nil {
		return entity.Task{}, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return entity.Task{}, err
	}

	defer resp.Body.Close()

	var task entity.Task
	err = json.NewDecoder(resp.Body).Decode(&task)
	if err != nil {
		return entity.Task{}, err
	}

	return task, nil
}

func (t *taskClient) UpdateTask(id, title, description, userID string) (respCode int, err error) {
	client, err := GetClientWithCookie(userID)
	if err != nil {
		return -1, err
	}

	taskId, err := strconv.Atoi(id)
	if err != nil {
		return -1, err
	}

	datajson := map[string]interface{}{
		"id":          int(taskId),
		"title":       title,
		"description": description,
	}

	b, err := json.Marshal(datajson)
	if err != nil {
		return -1, err
	}

	req, err := http.NewRequest("PUT", config.SetUrl("/api/v1/tasks/update?task_id="+id), bytes.NewBuffer(b))
	if err != nil {
		return -1, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return -1, err
	}

	defer resp.Body.Close()

	return resp.StatusCode, nil
}

func (t *taskClient) UpdateCategoryTask(id, catId, userID string) (respCode int, err error) {
	client, err := GetClientWithCookie(userID)
	if err != nil {
		return -1, err
	}

	taskId, err := strconv.Atoi(id)
	if err != nil {
		return -1, err
	}

	categoryId, err := strconv.Atoi(catId)
	if err != nil {
		return -1, err
	}

	datajson := map[string]interface{}{
		"id":          int(taskId),
		"category_id": int(categoryId),
	}

	b, err := json.Marshal(datajson)
	if err != nil {
		return -1, err
	}

	req, err := http.NewRequest("PUT", config.SetUrl("/api/v1/tasks/update/category?task_id="+id), bytes.NewBuffer(b))
	if err != nil {
		return -1, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return -1, err
	}

	defer resp.Body.Close()

	return resp.StatusCode, nil
}

func (t *taskClient) DeleteTask(id, userID string) (respCode int, err error) {
	client, err := GetClientWithCookie(userID)
	if err != nil {
		return -1, err
	}

	req, err := http.NewRequest("DELETE", config.SetUrl("/api/v1/tasks/delete?task_id="+id), nil)
	if err != nil {
		return -1, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return -1, err
	}

	defer resp.Body.Close()

	return resp.StatusCode, nil
}
