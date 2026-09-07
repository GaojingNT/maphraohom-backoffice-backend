package color

import "testing"

func Test_color(t *testing.T) {
	type args struct {
		c int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Test color",
			args: args{
				c: 0,
			},
			want: "\x1b[0m",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := color(tt.args.c); got != tt.want {
				t.Errorf("color() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormat(t *testing.T) {
	type args struct {
		c    int
		text string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Test format",
			args: args{
				c:    0,
				text: "test-message",
			},
			want: "\x1b[0mtest-message\x1b[0m",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Format(tt.args.c, tt.args.text); got != tt.want {
				t.Errorf("Format() = %v, want %v", got, tt.want)
			}
		})
	}
}
