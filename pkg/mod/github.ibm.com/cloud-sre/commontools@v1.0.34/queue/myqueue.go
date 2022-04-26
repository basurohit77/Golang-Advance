package queue

//Queue a queue 
type Queue struct {
	data []interface{}
}

//Empty the queue is empty or not
func (q *Queue) Empty() bool{
	return len(q.data) == 0
}

//Front return the front data of the queue
func (q *Queue) Front() interface{} {
	if q.Empty() {
		return nil
	}
	return q.data[0]
}

//Rear return the rear data of the queue
func (q *Queue) Rear() interface{} {
	if q.Empty() {
		return nil
	}
	return q.data[len(q.data) - 1]
}

//Pop get the front data of the queue
func (q *Queue) Pop() interface{}{
	if q.Empty() {
		return nil
	}
	frontData := q.data[0] 
	q.data = q.data[1:]
	return frontData
}

//Push add data to the front of the queue
func (q *Queue) Push(data interface{}){
	q.data = append(q.data, data)	
}

//Size return the queue size
func (q *Queue) Size() int{
	return len(q.data)
}
