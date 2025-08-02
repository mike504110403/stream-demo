package response

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSuccessResponse(t *testing.T) {
	data := map[string]string{"key": "value"}
	response := NewSuccessResponse(data)

	assert.Equal(t, 200, response.Code)
	assert.Equal(t, "success", response.Message)
	assert.Equal(t, data, response.Data)
}

func TestNewErrorResponse(t *testing.T) {
	response := NewErrorResponse(400, "Bad Request")

	assert.Equal(t, 400, response.Code)
	assert.Equal(t, "Bad Request", response.Message)
}

func TestNewPageResponse(t *testing.T) {
	data := []string{"item1", "item2", "item3"}
	response := NewPageResponse(100, 10, 20, data)

	assert.Equal(t, int64(100), response.Total)
	assert.Equal(t, 10, response.Offset)
	assert.Equal(t, 20, response.Limit)
	assert.Equal(t, data, response.Data)
}

func TestNewListResponse(t *testing.T) {
	items := []string{"item1", "item2", "item3"}
	response := NewListResponse(int64(100), items)

	assert.Equal(t, int64(100), response.Total)
	assert.Len(t, response.Items, 3)
	assert.Equal(t, "item1", response.Items[0])
	assert.Equal(t, "item2", response.Items[1])
	assert.Equal(t, "item3", response.Items[2])
}

func TestNewListResponse_EmptySlice(t *testing.T) {
	items := []string{}
	response := NewListResponse(int64(0), items)

	assert.Equal(t, int64(0), response.Total)
	assert.Len(t, response.Items, 0)
}

func TestNewListResponse_IntSlice(t *testing.T) {
	items := []int{1, 2, 3, 4, 5}
	response := NewListResponse(int64(5), items)

	assert.Equal(t, int64(5), response.Total)
	assert.Len(t, response.Items, 5)
	assert.Equal(t, 1, response.Items[0])
	assert.Equal(t, 5, response.Items[4])
} 