package connectivity

import "testing"

func TestMediaPortPlanValidate(t *testing.T) {
	tests := []struct {
		name    string
		plan    MediaPortPlan
		wantErr bool
	}{
		{
			name: "direct UDP mux",
			plan: MediaPortPlan{Mode: ModeDirect, UDPMuxPort: 59000},
		},
		{
			name: "direct TCP and UDP mux",
			plan: MediaPortPlan{Mode: ModeDirect, UDPMuxPort: 59000, TCPMuxPort: 59000, NAT1To1IP: "203.0.113.10"},
		},
		{
			name: "FRP same port",
			plan: MediaPortPlan{Mode: ModeFRP, UDPMuxPort: 40000, TCPMuxPort: 40000, NAT1To1IP: "198.51.100.20"},
		},
		{
			name:    "missing mode",
			plan:    MediaPortPlan{UDPMuxPort: 59000},
			wantErr: true,
		},
		{
			name:    "no mux",
			plan:    MediaPortPlan{Mode: ModeDirect},
			wantErr: true,
		},
		{
			name:    "out of range port",
			plan:    MediaPortPlan{Mode: ModeDirect, UDPMuxPort: 65536},
			wantErr: true,
		},
		{
			name:    "invalid NAT IP",
			plan:    MediaPortPlan{Mode: ModeDirect, UDPMuxPort: 59000, NAT1To1IP: "relay.example.com"},
			wantErr: true,
		},
		{
			name:    "FRP missing TCP",
			plan:    MediaPortPlan{Mode: ModeFRP, UDPMuxPort: 40000, NAT1To1IP: "198.51.100.20"},
			wantErr: true,
		},
		{
			name:    "FRP mismatched ports",
			plan:    MediaPortPlan{Mode: ModeFRP, UDPMuxPort: 40000, TCPMuxPort: 40001, NAT1To1IP: "198.51.100.20"},
			wantErr: true,
		},
		{
			name:    "FRP missing public IP",
			plan:    MediaPortPlan{Mode: ModeFRP, UDPMuxPort: 40000, TCPMuxPort: 40000},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.plan.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
