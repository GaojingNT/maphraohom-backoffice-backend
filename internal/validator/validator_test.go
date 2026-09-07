package validator

import (
	"reflect"
	"testing"

	"maphraohom.app/maphraohom-backoffice/internal/exception"
)

func TestValidate(t *testing.T) {
	type TestStruct struct {
		Name string `validate:"required"`
	}
	type args struct {
		data TestStruct
	}
	tests := []struct {
		name string
		args args
		want []exception.ParameterError
	}{
		{
			name: "Test 1",
			args: args{
				data: TestStruct{
					Name: "",
				},
			},
			want: []exception.ParameterError{
				{
					FailedField: "TestStruct.Name",
					Tag:         "required",
					Value:       "",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Validate(tt.args.data); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Validate() = %v, want %v", got, tt.want)
			}
		})
	}
}
