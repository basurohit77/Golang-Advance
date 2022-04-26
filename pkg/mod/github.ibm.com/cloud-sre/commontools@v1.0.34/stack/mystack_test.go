package stack

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPush(t *testing.T) {
	assert := assert.New(t)
	stack := &Stack{}
	stack.Push("1")
	assert.Equal(1, stack.Size(), "stack length should be 1")
	stack.Push("2")
	assert.Equal(2, stack.Size(), "stack length should be 2")
	stack.Push("3")
	assert.Equal(3, stack.Size(), "stack length should be 3")
	three := stack.Peek()
	assert.Equal("3", three, "stack top should be 3")
}

func TestPop(t *testing.T)  {
	assert := assert.New(t)
	stack := &Stack{}
	stack.Push(1)
	stack.Push(2)
	stack.Push(3)
	stack.Push(4)
	assert.Equal(4, stack.Size(), "stack length should be 4")

	four := stack.Peek()
	assert.Equal(4, four, "stack top should be four")
	four = stack.Pop()
	assert.Equal(4, four, "popped value should be 4 ")
	three := stack.Pop()
	assert.Equal(3, three, "popped value should be 3 ")
	two := stack.Pop()
	assert.Equal(2, two, "popped value should be 2")
	isEmpty := stack.Empty()
	assert.True(!isEmpty, "stack should not be empty now")
	one := stack.Pop()
	assert.Equal(1, one, "popped value should be 1")
	empty := stack.Pop()
	isEmpty = stack.Empty()
	assert.True(isEmpty, "stack is empty now")
	assert.Equal(nil, empty, "stack is empty")
	assert.Equal(0, stack.Size(), "stack is empty")
	empty = stack.Pop()
	assert.Equal(nil, empty, "stack is empty")
}