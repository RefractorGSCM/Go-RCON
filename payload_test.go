package rcon

import (
	"regexp"
	"testing"
)

func Test_payload_isNotBroadcast(t *testing.T) {
	type fields struct {
		ID   int32
		Type int32
		Body []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "TestIsNotBroadcast-01",
			fields: fields{
				ID:   0,
				Type: 0,
				Body: []byte("Chat: 83A832G6428U3D2, Test, (ALL) test"),
			},
			want: false,
		},
		{
			name: "TestIsNotBroadcast-02",
			fields: fields{
				ID:   0,
				Type: 0,
				Body: []byte("Keeping client alive for another 0 seconds"),
			},
			want: true,
		},
		{
			name: "TestIsNotBroadcast-03",
			fields: fields{
				ID:   0,
				Type: 0,
				Body: []byte("Login: 2020.12.14-04.00.00: Test (83A832G6428U3D2) logged in"),
			},
			want: false,
		},
		{
			name: "TestIsNotBroadcast-04",
			fields: fields{
				ID:   0,
				Type: 0,
				Body: []byte("Killfeed: 2020.12.14-04.46.32: 83A832G6428U3D2 (Test) killed  (Test2)"),
			},
			want: false,
		},
		{
			name: "TestIsNotBroadcast-05",
			fields: fields{
				ID:   0,
				Type: 0,
				Body: []byte("Scorefeed: 2020.12.14-04.46.32: 83A832G6428U3D2 (Test)'s score changed by 100.0 points and is now 200.0 points"),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &payload{
				ID:   tt.fields.ID,
				Type: tt.fields.Type,
				Body: tt.fields.Body,
				NonBroadcastPatterns: []*regexp.Regexp{
					regexp.MustCompile("^Keeping client alive for another\\s[0-9]{1,6}\\sseconds$"),
				},
			}
			if got := p.isNotBroadcast(); got != tt.want {
				t.Errorf("payload.isNotBroadcast() = %v, want %v", got, tt.want)
			}
		})
	}
}
