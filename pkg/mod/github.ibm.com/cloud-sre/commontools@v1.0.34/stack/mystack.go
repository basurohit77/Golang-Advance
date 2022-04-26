package stack

//Stack a customsize stack 
type Stack struct {
	data []interface{}
}

//Empty the stack is empty or not
func (s *Stack) Empty() bool{
	return len(s.data) == 0
}

//Peek check the top value of the stack
func (s *Stack) Peek() interface{} {
	if s.Empty(){
		return nil
	}
	length := len(s.data)
	return s.data[length - 1]
}

//Pop get the front data of the queue
func (s *Stack) Pop() interface{}{
	if s.Empty() {
		return nil
	}
	topData := s.data[len(s.data) - 1] 
	s.data = s.data[0: len(s.data) - 1]
	return topData
}

//Push add data to the front of the queue
func (s *Stack) Push(data interface{}){
	s.data = append(s.data, data)	
}

//Size return the queue size
func (s *Stack) Size() int{
	return len(s.data)
}
