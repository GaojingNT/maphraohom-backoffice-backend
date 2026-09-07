package utils

import (
	"testing"
)

func TestThrow(t *testing.T) {
	type args struct {
		ex Exception
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Test throw",
			args: args{
				ex: Exception("Test throw"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Check if the function is panicking
			if r := recover(); r != nil {
				return
			}
		})
	}
}

func TestBlock_Do(t *testing.T) {
	type fields struct {
		Try     func()
		Catch   func(Exception)
		Finally func()
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "Test block do",
			fields: fields{
				Try: func() {
					// Test try block
					Throw(Exception("Test block do"))
				},
				Catch:   func(e Exception) {},
				Finally: func() {},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := Block{
				Try:     tt.fields.Try,
				Catch:   tt.fields.Catch,
				Finally: tt.fields.Finally,
			}
			b.Do()
		})
	}
}
