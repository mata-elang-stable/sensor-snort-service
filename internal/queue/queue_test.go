package queue

import (
	"reflect"
	"testing"

	"github.com/fadhilyori/mesentinel/internal/pb"
	"github.com/fadhilyori/mesentinel/internal/types"
)

func TestNewEventBatchQueue(t *testing.T) {
	tests := []struct {
		name string
		want *EventBatchQueue
	}{
		{
			name: "Must return a valid EventBatchQueue instance",
			want: &EventBatchQueue{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewEventBatchQueue(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewEventBatchQueue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQueue_AddRecordToQueue(t *testing.T) {
	type fields struct {
		queue map[int64]*BatchRecord
	}
	type args struct {
		record *types.SnortAlert
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   map[int64]*BatchRecord
	}{
		{
			name: "Must add a record to the queue",
			fields: fields{
				queue: make(map[int64]*BatchRecord, 86400),
			},
			args: args{
				record: &types.SnortAlert{
					Metadata: types.Metadata{
						SensorID:   "sensor1",
						Index:      0,
						HashSHA256: "cb1b7a6851fd5ab0a643aee2cabd10c1c4d409f4662cd0f4d3918cd5374ca762",
					},
					Action:         "allow",
					Base64Data:     "dGVzdGluZyBvbmx5",
					Classification: "none",
					ClientBytes:    3012,
					ClientPkts:     2,
					Direction:      "C2S",
					DstAddr:        "192.168.50.1",
					EthDst:         "90:B1:1C:A2:C0:D3",
					EthLen:         1506,
					EthSrc:         "70:F3:5A:42:73:E8",
					EthType:        "0x800",
					FlowStartTime:  1543675989,
					GID:            123,
					Interface:      "eth0",
					IPID:           3620,
					IPLen:          1472,
					MPLS:           0,
					Message:        "(stream_ip) fragmentation overlap",
					PktGen:         "raw",
					PktLen:         1492,
					PktNum:         140309,
					Priority:       3,
					Protocol:       "IP",
					Revision:       1,
					Seconds:        1543675989,
					ServerBytes:    0,
					ServerPkts:     0,
					Service:        "unknown",
					SID:            8,
					SrcAddr:        "172.16.0.5",
					Timestamp:      "18/12/01-14:53:09.797526",
					TOS:            0,
					TTL:            111,
					VLAN:           0,
				},
			},
		},
		{
			name: "Must return a single record in a batch",
			fields: fields{
				queue: map[int64]*BatchRecord{
					1543675989: {
						RecordMap: map[string]*SensorEventRecord{
							"cb1b7a6851fd5ab0a643aee2cabd10c1c4d409f4662cd0f4d3918cd5374ca762": {
								Payload: &pb.SensorEvent{
									Metadata: &pb.Metadata{
										SensorId:      "",
										SensorVersion: "",
										Index:         0,
										SentAt:        0,
										HashSha256:    "",
									},
									Metrics: []*pb.Metric{
										{
											Base64Data:    "dGVzdGluZyBvbmx5",
											ClientBytes:   3012,
											ClientPkts:    2,
											EthLen:        1506,
											FlowstartTime: 1543675989,
											IpId:          3620,
											IpLength:      1472,
											Mpls:          0,
											PktLength:     1492,
											PktNumber:     140309,
											ServerBytes:   0,
											ServerPkts:    0,
											Timestamp:     "18/12/01-14:53:09.797526",
											TypeOfService: 0,
											TimeToLive:    111,
											Vlan:          0,
										},
									},
									Attributes: &pb.Attributes{
										Action:         "allow",
										Classification: "none",
										Direction:      "C2S",
										DstAddress:     "192.168.50.1",
										EthDst:         "90:B1:1C:A2:C0:D3",
										EthSrc:         "70:F3:5A:42:73:E8",
										EthType:        "0x800",
										Rule: &pb.Rule{
											Gid: 123,
											Rev: 1,
											Sid: 8,
										},
										Interface:  "eth0",
										Message:    "(stream_ip) fragmentation overlap",
										PktGen:     "raw",
										Priority:   3,
										Protocol:   "IP",
										Seconds:    1543675989,
										Service:    "unknown",
										SrcAddress: "172.16.0.5",
									},
								},
							},
						},
					},
				},
			},
			args: args{
				record: &types.SnortAlert{
					Metadata: types.Metadata{
						SensorID:   "sensor1",
						Index:      0,
						HashSHA256: "cb1b7a6851fd5ab0a643aee2cabd10c1c4d409f4662cd0f4d3918cd5374ca762",
					},
					Action:         "allow",
					Base64Data:     "dGVzdGluZyBvbmx5",
					Classification: "none",
					ClientBytes:    3012,
					ClientPkts:     2,
					Direction:      "C2S",
					DstAddr:        "192.168.50.1",
					EthDst:         "90:B1:1C:A2:C0:D3",
					EthLen:         1506,
					EthSrc:         "70:F3:5A:42:73:E8",
					EthType:        "0x800",
					FlowStartTime:  1543675989,
					GID:            123,
					Interface:      "eth0",
					IPID:           3620,
					IPLen:          1472,
					MPLS:           0,
					Message:        "(stream_ip) fragmentation overlap",
					PktGen:         "raw",
					PktLen:         1492,
					PktNum:         140309,
					Priority:       3,
					Protocol:       "IP",
					Revision:       1,
					Seconds:        1543675989,
					ServerBytes:    0,
					ServerPkts:     0,
					Service:        "unknown",
					SID:            8,
					SrcAddr:        "172.16.0.5",
					Timestamp:      "18/12/01-14:53:09.797527",
					TOS:            0,
					TTL:            111,
					VLAN:           0,
				},
			},
			want: map[int64]*BatchRecord{
				1543675989: {
					RecordMap: map[string]*SensorEventRecord{
						"cb1b7a6851fd5ab0a643aee2cabd10c1c4d409f4662cd0f4d3918cd5374ca762": {
							Payload: &pb.SensorEvent{
								Metadata: &pb.Metadata{
									SensorId:      "",
									SensorVersion: "",
									Index:         0,
									SentAt:        0,
									HashSha256:    "",
								},
								Metrics: []*pb.Metric{
									{
										Base64Data:    "dGVzdGluZyBvbmx5",
										ClientBytes:   3012,
										ClientPkts:    2,
										EthLen:        1506,
										FlowstartTime: 1543675989,
										IpId:          3620,
										IpLength:      1472,
										Mpls:          0,
										PktLength:     1492,
										PktNumber:     140309,
										ServerBytes:   0,
										ServerPkts:    0,
										Timestamp:     "18/12/01-14:53:09.797526",
										TypeOfService: 0,
										TimeToLive:    111,
										Vlan:          0,
									},
									{
										Base64Data:    "dGVzdGluZyBvbmx5",
										ClientBytes:   3012,
										ClientPkts:    2,
										EthLen:        1506,
										FlowstartTime: 1543675989,
										IpId:          3620,
										IpLength:      1472,
										Mpls:          0,
										PktLength:     1492,
										PktNumber:     140309,
										ServerBytes:   0,
										ServerPkts:    0,
										Timestamp:     "18/12/01-14:53:09.797527",
										TypeOfService: 0,
										TimeToLive:    111,
										Vlan:          0,
									},
								},
								Attributes: &pb.Attributes{
									Action:         "allow",
									Classification: "none",
									Direction:      "C2S",
									DstAddress:     "192.168.50.1",
									EthDst:         "90:B1:1C:A2:C0:D3",
									EthSrc:         "70:F3:5A:42:73:E8",
									EthType:        "0x800",
									Rule: &pb.Rule{
										Gid: 123,
										Rev: 1,
										Sid: 8,
									},
									Interface:  "eth0",
									Message:    "(stream_ip) fragmentation overlap",
									PktGen:     "raw",
									Priority:   3,
									Protocol:   "IP",
									Seconds:    1543675989,
									Service:    "unknown",
									SrcAddress: "172.16.0.5",
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &EventBatchQueue{
				queue: tt.fields.queue,
			}
			q.AddRecordToQueue(tt.args.record)

			// Check if the record is added to the queue
			if _, ok := q.queue[tt.args.record.Seconds]; !ok {
				t.Errorf("Record is not added to the queue")
			}

			// Check if the record is added to the record list
			if _, ok := q.queue[tt.args.record.Seconds].RecordList[tt.args.record.Metadata.HashSHA256]; !ok {
				t.Errorf("Record is not added to the record list")
			}

			// Check if the record is equal to the record in the record list
			if tt.want != nil && !reflect.DeepEqual(q.queue, tt.want) {
				t.Errorf("Record is not equal to the record in the record list\n got\t= %v\n want\t= %v", q.queue, tt.want)
			}
		})
	}
}

func TestQueue_DeleteRecordFromQueue(t *testing.T) {
	type fields struct {
		queue map[int64]*BatchRecord
	}
	type args struct {
		seconds int64
		hash    string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "Must delete a record from the queue",
			fields: fields{
				queue: map[int64]*BatchRecord{
					1543675989: {
						RecordList: map[string]*Record{
							"cb1b7a6851fd5ab0a643aee2cabd10c1c4d409f4662cd0f4d3918cd5374ca762": {
								Payload: &pb.SensorEvent{
									Metadata: &pb.Metadata{
										SensorId:      "",
										SensorVersion: "",
										Index:         0,
										SentAt:        0,
										HashSha256:    "",
									},
									Metrics: []*pb.Metric{
										{
											Base64Data:    "dGVzdGluZyBvbmx5",
											ClientBytes:   3012,
											ClientPkts:    2,
											EthLen:        1506,
											FlowstartTime: 1543675989,
											IpId:          3620,
											IpLength:      1472,
											Mpls:          0,
											PktLength:     1492,
											PktNumber:     140309,
											ServerBytes:   0,
											ServerPkts:    0,
											Timestamp:     "18/12/01-14:53:09.797526",
											TypeOfService: 0,
											TimeToLive:    111,
											Vlan:          0,
										},
									},
									Attributes: &pb.Attributes{
										Action:         "allow",
										Classification: "none",
										Direction:      "C2S",
										DstAddress:     "192.168.50.1",
										EthDst:         "90:B1:1C:A2:C0:D3",
										EthSrc:         "70:F3:5A:42:73:E8",
										EthType:        "0x800",
										Rule: &pb.Rule{
											Gid: 123,
											Rev: 1,
											Sid: 8,
										},
										Interface:  "eth0",
										Message:    "(stream_ip) fragmentation overlap",
										PktGen:     "raw",
										Priority:   3,
										Protocol:   "IP",
										Seconds:    1543675989,
										Service:    "unknown",
										SrcAddress: "172.16.0.5",
									},
								},
							},
						},
					},
				},
			},
			args: args{
				seconds: 1543675989,
				hash:    "cb1b7a6851fd5ab0a643aee2cabd10c1c4d409f4662cd0f4d3918cd5374ca762",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &EventBatchQueue{
				queue: tt.fields.queue,
			}
			q.DeleteRecordFromQueue(tt.args.seconds, tt.args.hash)

			// Check if the record is deleted from the record list
			if _, ok := q.queue[tt.args.seconds].RecordList[tt.args.hash]; ok {
				t.Errorf("Record is not deleted from the record list")
			}
		})
	}
}

func TestQueue_GetRecordFromQueue(t *testing.T) {
	type fields struct {
		queue map[int64]*BatchRecord
	}
	type args struct {
		seconds int64
		hash    string
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		want        *pb.SensorEvent
		wantSuccess bool
	}{
		{
			name: "Must return a record from the queue",
			fields: fields{
				queue: map[int64]*BatchRecord{
					1543675989: {
						RecordList: map[string]*Record{
							"cb1b7a6851fd5ab0a643aee2cabd10c1c4d409f4662cd0f4d3918cd5374ca762": {
								Payload: &pb.SensorEvent{
									Metadata: &pb.Metadata{
										SensorId:      "",
										SensorVersion: "",
										Index:         0,
										SentAt:        0,
										HashSha256:    "",
									},
									Metrics: []*pb.Metric{
										{
											Base64Data:    "dGVzdGluZyBvbmx5",
											ClientBytes:   3012,
											ClientPkts:    2,
											EthLen:        1506,
											FlowstartTime: 1543675989,
											IpId:          3620,
											IpLength:      1472,
											Mpls:          0,
											PktLength:     1492,
											PktNumber:     140309,
											ServerBytes:   0,
											ServerPkts:    0,
											Timestamp:     "18/12/01-14:53:09.797526",
											TypeOfService: 0,
											TimeToLive:    111,
											Vlan:          0,
										},
									},
									Attributes: &pb.Attributes{
										Action:         "allow",
										Classification: "none",
										Direction:      "C2S",
										DstAddress:     "192.168.50.1",
										EthDst:         "90:B1:1C:A2:C0:D3",
										EthSrc:         "70:F3:5A:42:73:E8",
										EthType:        "0x800",
										Rule: &pb.Rule{
											Gid: 123,
											Rev: 1,
											Sid: 8,
										},
										Interface:  "eth0",
										Message:    "(stream_ip) fragmentation overlap",
										PktGen:     "raw",
										Priority:   3,
										Protocol:   "IP",
										Seconds:    1543675989,
										Service:    "unknown",
										SrcAddress: "172.16.0.5",
									},
								},
							},
						},
					},
				},
			},
			args: args{
				seconds: 1543675989,
				hash:    "cb1b7a6851fd5ab0a643aee2cabd10c1c4d409f4662cd0f4d3918cd5374ca762",
			},
			want: &pb.SensorEvent{
				Metadata: &pb.Metadata{
					SensorId:      "",
					SensorVersion: "",
					Index:         0,
					SentAt:        0,
					HashSha256:    "",
				},
				Metrics: []*pb.Metric{
					{
						Base64Data:    "dGVzdGluZyBvbmx5",
						ClientBytes:   3012,
						ClientPkts:    2,
						EthLen:        1506,
						FlowstartTime: 1543675989,
						IpId:          3620,
						IpLength:      1472,
						Mpls:          0,
						PktLength:     1492,
						PktNumber:     140309,
						ServerBytes:   0,
						ServerPkts:    0,
						Timestamp:     "18/12/01-14:53:09.797526",
						TypeOfService: 0,
						TimeToLive:    111,
						Vlan:          0,
					},
				},
				Attributes: &pb.Attributes{
					Action:         "allow",
					Classification: "none",
					Direction:      "C2S",
					DstAddress:     "192.168.50.1",
					EthDst:         "90:B1:1C:A2:C0:D3",
					EthSrc:         "70:F3:5A:42:73:E8",
					EthType:        "0x800",
					Rule: &pb.Rule{
						Gid: 123,
						Rev: 1,
						Sid: 8,
					},
					Interface:  "eth0",
					Message:    "(stream_ip) fragmentation overlap",
					PktGen:     "raw",
					Priority:   3,
					Protocol:   "IP",
					Seconds:    1543675989,
					Service:    "unknown",
					SrcAddress: "172.16.0.5",
				},
			},
			wantSuccess: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &EventBatchQueue{
				queue: tt.fields.queue,
			}
			got, got1 := q.GetRecordFromQueue(tt.args.seconds, tt.args.hash)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetRecordFromQueue() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.wantSuccess {
				t.Errorf("GetRecordFromQueue() got1 = %v, want %v", got1, tt.wantSuccess)
			}
		})
	}
}
