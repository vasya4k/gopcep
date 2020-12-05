package pcep

func getLSPS() []*SRLSP {
	return []*SRLSP{
		{
			Delegate: true,
			Sync:     false,
			Remove:   false,
			Admin:    true,
			Name:     "LSP-55",
			Src:      "10.10.10.10",
			Dst:      "14.14.14.14",
			EROList: []SREROSub{
				{
					LooseHop:   false,
					MBit:       true,
					NT:         3,
					IPv4NodeID: "",
					SID:        402011,
					NoSID:      false,
					IPv4Adjacency: []string{
						0: "10.1.0.1",
						1: "10.1.0.0",
					},
				},
				{
					LooseHop:   false,
					MBit:       true,
					NT:         1,
					IPv4NodeID: "15.15.15.15",
					SID:        402015,
					NoSID:      false,
				},
				{
					LooseHop:   false,
					MBit:       true,
					NT:         1,
					IPv4NodeID: "14.14.14.14",
					SID:        402014,
					NoSID:      false,
				},
			},
			SetupPrio:    7,
			HoldPrio:     7,
			LocalProtect: false,
			BW:           100,
		},
		{
			Delegate: true,
			Sync:     false,
			Remove:   false,
			Admin:    true,
			Name:     "LSP-66",
			Src:      "10.10.10.10",
			Dst:      "13.13.13.13",
			EROList: []SREROSub{
				{
					LooseHop:   false,
					MBit:       true,
					NT:         3,
					IPv4NodeID: "",
					SID:        402011,
					NoSID:      false,
					IPv4Adjacency: []string{
						0: "10.1.0.1",
						1: "10.1.0.0",
					},
				},
				{
					LooseHop:   false,
					MBit:       true,
					NT:         1,
					IPv4NodeID: "15.15.15.15",
					SID:        402015,
					NoSID:      false,
				},
				{
					LooseHop:   false,
					MBit:       true,
					NT:         1,
					IPv4NodeID: "14.14.14.14",
					SID:        402014,
					NoSID:      false,
				},
				{
					LooseHop:   false,
					MBit:       true,
					NT:         1,
					IPv4NodeID: "13.13.13.13",
					SID:        402013,
					NoSID:      false,
				},
			},
			SetupPrio:    7,
			HoldPrio:     7,
			LocalProtect: false,
			BW:           100,
		},
	}
}
