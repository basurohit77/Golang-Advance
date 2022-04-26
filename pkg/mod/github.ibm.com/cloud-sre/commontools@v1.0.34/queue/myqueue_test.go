package queue

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEmpty(t *testing.T) {
	assert := assert.New(t)
	q := &Queue{}
	assert.True(true, q.Empty(), "queue should be empty")
}

func TestFront(t *testing.T) {
	assert := assert.New(t)
	q := &Queue{}
	assert.Equal(nil, q.Front(), "no data in queue")

	q.Push("123")
	assert.Equal("123", q.Front(), "the front of the queue should be 123")
}

func TestRear(t *testing.T) {
	assert := assert.New(t)
	q := &Queue{}
	assert.Equal(nil, q.Rear(), "no data in queue")
	q.Push("456")
	assert.Equal("456", q.Rear(), "rear is 456")
}

func TestPush(t *testing.T) {
	assert := assert.New(t)
	q := &Queue{}
	q.Push("123")
	q.Push("456")
	assert.Equal(2, q.Size(), "2 data in queue")
}

func TestPop(t *testing.T)  {
	assert := assert.New(t)
	q := &Queue{}
	q.Push("123")
	q.Push("456")
	assert.Equal(2, q.Size(), "2 data in queue")
	q.Pop()
	assert.Equal(1, q.Size(), "1 data in queue")
	q.Pop()
	assert.Equal(0, q.Size(), "no data in queue")
	q.Push("123")
	q.Push("456")
	assert.Equal("123", q.Front(), "font is 123")
	assert.Equal("456", q.Rear(), "rear is 456")
	q.Pop()
	assert.Equal("456", q.Front(), "font is 456")
	assert.Equal("456", q.Rear(), "rear is 456")

}

