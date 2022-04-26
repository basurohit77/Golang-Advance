package pep

import (
	"reflect"
	"testing"
)

func Test_cacheKeyInfo_getCacheKeyPattern(t *testing.T) {
	type fields struct {
		CacheKeyPattern CacheKeyPattern
	}
	tests := []struct {
		name        string
		fields      fields
		wantPattern CacheKeyPattern
	}{
		{
			name: "can return an empty pattern",
			fields: fields{
				CacheKeyPattern: CacheKeyPattern{},
			},
			wantPattern: CacheKeyPattern{},
		},
		{
			name: "can return Order with 1 entry",
			fields: fields{
				CacheKeyPattern: CacheKeyPattern{
					Order: []string{"subject"},
				},
			},
			wantPattern: CacheKeyPattern{
				Order: []string{"subject"},
			},
		},
		{
			name: "can return Order with more than 1 entries",
			fields: fields{
				CacheKeyPattern: CacheKeyPattern{
					Order: []string{"subject", "resource"},
				},
			},
			wantPattern: CacheKeyPattern{
				Order: []string{"subject", "resource"},
			},
		},
		{
			name: "can return Resource with 1 pattern",
			fields: fields{
				CacheKeyPattern: CacheKeyPattern{
					Order:    []string{"subject", "resource"},
					Resource: [][]string{{"serviceName"}},
				},
			},
			wantPattern: CacheKeyPattern{
				Order:    []string{"subject", "resource"},
				Resource: [][]string{{"serviceName"}},
			},
		},
		{
			name: "can return Resource with more than 1 patterns",
			fields: fields{
				CacheKeyPattern: CacheKeyPattern{
					Order: []string{"subject", "resource"},
					Resource: [][]string{
						{"serviceName"},
						{"serviceName", "accountID"},
					},
				},
			},
			wantPattern: CacheKeyPattern{
				Order: []string{"subject", "resource"},
				Resource: [][]string{
					{"serviceName"},
					{"serviceName", "accountID"},
				},
			},
		},
		{
			name: "can return Subject with 1 pattern",
			fields: fields{
				CacheKeyPattern: CacheKeyPattern{
					Order:   []string{"subject", "resource"},
					Subject: [][]string{{"id"}},
					Resource: [][]string{
						{"serviceName"},
						{"serviceName", "accountID"},
					},
				},
			},
			wantPattern: CacheKeyPattern{
				Order: []string{"subject", "resource"},
				Resource: [][]string{
					{"serviceName"},
					{"serviceName", "accountID"},
				},
				Subject: [][]string{{"id"}},
			},
		},
		{
			name: "can return Subject with more than 1 patterns",
			fields: fields{
				CacheKeyPattern: CacheKeyPattern{
					Order:   []string{"subject", "resource"},
					Subject: [][]string{{"id"}, {"id", "scope"}},
					Resource: [][]string{
						{"serviceName"},
						{"serviceName", "accountID"},
					},
				},
			},
			wantPattern: CacheKeyPattern{
				Order: []string{"subject", "resource"},
				Resource: [][]string{
					{"serviceName"},
					{"serviceName", "accountID"},
				},
				Subject: [][]string{{"id"}, {"id", "scope"}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &cacheKeyInfo{
				CacheKeyPattern: tt.fields.CacheKeyPattern,
			}
			if gotPattern := c.getCacheKeyPattern(); !reflect.DeepEqual(gotPattern, tt.wantPattern) {
				t.Errorf("cacheKeyInfo.getCacheKeyPattern() = %v, want %v", gotPattern, tt.wantPattern)
			}
		})
	}
}

func Test_cacheKeyInfo_storeCacheKeyPattern(t *testing.T) {
	type args struct {
		src CacheKeyPattern
	}
	tests := []struct {
		name        string
		args        args
		wantPattern CacheKeyPattern
	}{
		{
			name: "should store an empty pattern",
			args: args{
				src: CacheKeyPattern{},
			},
			wantPattern: CacheKeyPattern{},
		},
		{
			name: "should store Order with 1 entry",
			args: args{
				src: CacheKeyPattern{
					Order: []string{"subject"},
				},
			},
			wantPattern: CacheKeyPattern{
				Order: []string{"subject"},
			},
		},
		{
			name: "should store Order with more than 1 entries",
			args: args{
				src: CacheKeyPattern{
					Order: []string{"subject", "resource"},
				},
			},
			wantPattern: CacheKeyPattern{
				Order: []string{"subject", "resource"},
			},
		},
		{
			name: "should store Resource with 1 pattern",
			args: args{
				src: CacheKeyPattern{
					Order:    []string{"subject", "resource"},
					Resource: [][]string{{"serviceName"}},
				},
			},
			wantPattern: CacheKeyPattern{
				Order:    []string{"subject", "resource"},
				Resource: [][]string{{"serviceName"}},
			},
		},
		{
			name: "should store Resource with more than 1 patterns",
			args: args{
				src: CacheKeyPattern{
					Order: []string{"subject", "resource"},
					Resource: [][]string{
						{"serviceName"},
						{"serviceName", "accountID"},
					},
				},
			},
			wantPattern: CacheKeyPattern{
				Order: []string{"subject", "resource"},
				Resource: [][]string{
					{"serviceName"},
					{"serviceName", "accountID"},
				},
			},
		},
		{
			name: "should store Subject with 1 pattern",
			args: args{
				src: CacheKeyPattern{
					Order:   []string{"subject", "resource"},
					Subject: [][]string{{"id"}},
					Resource: [][]string{
						{"serviceName"},
						{"serviceName", "accountID"},
					},
				},
			},
			wantPattern: CacheKeyPattern{
				Order: []string{"subject", "resource"},
				Resource: [][]string{
					{"serviceName"},
					{"serviceName", "accountID"},
				},
				Subject: [][]string{{"id"}},
			},
		},
		{
			name: "should store Subject with more than 1 patterns",
			args: args{
				src: CacheKeyPattern{
					Order:   []string{"subject", "resource"},
					Subject: [][]string{{"id"}, {"id", "scope"}},
					Resource: [][]string{
						{"serviceName"},
						{"serviceName", "accountID"},
					},
				},
			},
			wantPattern: CacheKeyPattern{
				Order: []string{"subject", "resource"},
				Resource: [][]string{
					{"serviceName"},
					{"serviceName", "accountID"},
				},
				Subject: [][]string{{"id"}, {"id", "scope"}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &cacheKeyInfo{
				CacheKeyPattern: tt.args.src,
			}
			c.storeCacheKeyPattern(tt.args.src)
			if gotPattern := c.getCacheKeyPattern(); !reflect.DeepEqual(gotPattern, tt.wantPattern) {
				t.Errorf("cacheKeyInfo.getCacheKeyPattern() = %v, want %v", gotPattern, tt.wantPattern)
			}
		})
	}
}
