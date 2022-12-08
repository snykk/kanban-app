package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/snykk/kanban-app/config"
	"github.com/snykk/kanban-app/entity"
)

type CategoryClient interface {
	GetCategories(userID string) ([]entity.CategoryData, error)
	AddCategories(title string, userID string) (respCode int, err error)
	DeleteCategory(id, userID string) (respCode int, err error)
}

type categoryClient struct {
}

func NewCategoryClient() *categoryClient {
	return &categoryClient{}
}

func (c *categoryClient) DeleteCategory(id, userID string) (respCode int, err error) {
	client, err := GetClientWithCookie(userID)
	if err != nil {
		return -1, err
	}

	req, err := http.NewRequest("DELETE", config.SetUrl("/api/v1/categories/delete?category_id="+id), nil)
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

func (c *categoryClient) GetCategories(userID string) ([]entity.CategoryData, error) {
	client, err := GetClientWithCookie(userID)

	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", config.SetUrl("/api/v1/categories/dashboard"), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, errors.New("status code not 200")
	}

	var categories []entity.CategoryData
	err = json.Unmarshal(b, &categories)
	if err != nil {
		return nil, err
	}

	return categories, nil
}

func (c *categoryClient) AddCategories(title string, userID string) (respCode int, err error) {
	client, err := GetClientWithCookie(userID)

	if err != nil {
		return -1, err
	}

	jsonData := map[string]string{
		"type": title,
	}

	data, err := json.Marshal(jsonData)
	if err != nil {
		return -1, err
	}

	req, err := http.NewRequest("POST", config.SetUrl("/api/v1/categories/create"), bytes.NewBuffer(data))
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
