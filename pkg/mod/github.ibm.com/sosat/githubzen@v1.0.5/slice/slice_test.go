package slice

import (
	"testing"
)

// func Test_Remove(t *testing.T) {
// 	tests := []struct {
// 		name   string
// 		slice  []string
// 		target string
// 		found  bool
// 		want   []string
// 	}{
// 		{
// 			name:   "remove C",
// 			slice:  []string{"A", "B", "C", "D", "E"},
// 			target: "C",
// 			want:   []string{"A", "B", "D", "E"},
// 		},
// 		{
// 			name:   "not found",
// 			slice:  []string{"A", "B", "C", "D", "E"},
// 			target: "Z",
// 			want:   []string{"A", "B", "C", "D", "E"},
// 		},
// 		{
// 			name:   "remove only first occurrence",
// 			slice:  []string{"A", "B", "C", "D", "E", "C"},
// 			target: "C",
// 			want:   []string{"A", "B", "D", "E", "C"},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			// if got := Remove(tt.slice, tt.target); !reflect.DeepEqual(got, tt.want) {
// 			got := Remove(tt.slice, tt.target)
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("Remove() = %v, want %v", got, tt.want)
// 			}
// 			// t.Logf("name=%s origLen=%d, newLen=%d\n", tt.name, len(tt.slice), len(Remove(tt.slice, tt.target)))
// 		})
// 	}
// }

func Test_Find(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		target   string
		location int
		found    bool
	}{
		{
			name:     "find only",
			slice:    []string{"A", "B", "C", "D", "E"},
			target:   "C",
			location: 2,
			found:    true,
		},
		{
			name:     "find first",
			slice:    []string{"A", "B", "C", "D", "E", "C"},
			target:   "C",
			location: 2,
			found:    true,
		},
		{
			name:     "not found",
			slice:    []string{"A", "B", "C", "D", "E"},
			target:   "Z",
			location: 5,
			found:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loc, found := Find(tt.slice, tt.target)
			if loc != tt.location {
				t.Errorf("location got=%d, wanted %d\n", loc, tt.location)
			}
			if found != tt.found {
				t.Errorf("found got=%t, wanted=%t", found, tt.found)
			}
		})
	}
}
