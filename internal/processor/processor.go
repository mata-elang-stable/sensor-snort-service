package processor

import (
	"crypto/sha256"
	"encoding/hex"
	"gitlab.com/mata-elang/v2/mes-snort/internal/logger"
	"gitlab.com/mata-elang/v2/mes-snort/internal/pb"
	"gitlab.com/mata-elang/v2/mes-snort/internal/types"
)

var (
	log = logger.GetLogger()
)

// GetRawDataFromMetrics converts the protobuf message to the internal data structure.
func GetRawDataFromMetrics(data *pb.SensorEvent, metric *pb.Metric) *types.SnortAlert {
	if data == nil || metric == nil {
		return nil
	}

	return &types.SnortAlert{
		Metadata: types.Metadata{
			SensorID:      data.SensorId,
			SensorVersion: data.SensorVersion,
			HashSHA256:    data.EventHashSha256,
			SentAt:        data.EventSentAt,
			ReadAt:        data.EventReadAt,
			ReceivedAt:    data.EventReceivedAt,
		},
		Action:         data.SnortAction,
		Base64Data:     metric.SnortBase64Data,
		Classification: data.SnortClassification,
		ClientBytes:    metric.SnortClientBytes,
		ClientPkts:     metric.SnortClientPkts,
		Direction:      data.SnortDirection,
		DstAddr:        metric.SnortDstAddress,
		DstAp:          metric.SnortDstAp,
		DstPort:        metric.SnortDstPort,
		EthDst:         metric.SnortEthDst,
		EthLen:         metric.SnortEthLen,
		EthSrc:         metric.SnortEthSrc,
		EthType:        metric.SnortEthType,
		FlowStartTime:  metric.SnortFlowstartTime,
		GeneveVNI:      metric.SnortGeneveVni,
		GID:            data.SnortRuleGid,
		ICMPCode:       metric.SnortIcmpCode,
		ICMPID:         metric.SnortIcmpId,
		ICMPSeq:        metric.SnortIcmpSeq,
		ICMPType:       metric.SnortIcmpType,
		Interface:      data.SnortInterface,
		IPID:           metric.SnortIpId,
		IPLen:          metric.SnortIpLength,
		MPLS:           metric.SnortMpls,
		Message:        data.SnortMessage,
		PktGen:         metric.SnortPktGen,
		PktLen:         metric.SnortPktLength,
		PktNum:         metric.SnortPktNumber,
		Priority:       data.SnortPriority,
		Protocol:       data.SnortProtocol,
		Revision:       data.SnortRuleRev,
		RuleID:         data.SnortRule,
		Seconds:        data.SnortSeconds,
		ServerBytes:    metric.SnortServerBytes,
		ServerPkts:     metric.SnortServerPkts,
		Service:        data.SnortService,
		SGT:            metric.SnortSgt,
		SID:            data.SnortRuleSid,
		SrcAddr:        metric.SnortSrcAddress,
		SrcAp:          metric.SnortSrcAp,
		SrcPort:        metric.SnortSrcPort,
		Target:         metric.SnortTarget,
		TCPAck:         metric.SnortTcpAck,
		TCPFlags:       metric.SnortTcpFlags,
		TCPLen:         metric.SnortTcpLen,
		TCPSeq:         metric.SnortTcpSeq,
		TCPWin:         metric.SnortTcpWin,
		Timestamp:      metric.SnortTimestamp,
		TOS:            data.SnortTypeOfService,
		TTL:            metric.SnortTimeToLive,
		UDPLen:         metric.SnortUdpLength,
		VLAN:           metric.SnortVlan,
	}
}

// ConvertSnortAlertToSensorEvent converts SnortAlert to SensorEvent.
// It is used to convert the internal data structure to the protobuf message.
func ConvertSnortAlertToSensorEvent(data *types.SnortAlert) (*pb.SensorEvent, *pb.Metric) {
	if data == nil {
		return nil, nil
	}

	sensorEvent := &pb.SensorEvent{
		Metrics:             nil,
		EventHashSha256:     "",
		EventMetricsCount:   1,
		EventSeconds:        data.Seconds,
		SensorId:            data.Metadata.SensorID,
		SensorVersion:       data.Metadata.SensorVersion,
		SnortAction:         data.Action,
		SnortClassification: data.Classification,
		SnortDirection:      data.Direction,
		SnortInterface:      data.Interface,
		SnortMessage:        data.Message,
		SnortPriority:       data.Priority,
		SnortProtocol:       data.Protocol,
		SnortRuleGid:        data.GID,
		SnortRuleRev:        data.Revision,
		SnortRuleSid:        data.SID,
		SnortRule:           data.RuleID,
		SnortSeconds:        data.Seconds,
		SnortService:        data.Service,
		SnortTypeOfService:  data.TOS,
	}

	// Generate SHA256 hash
	sensorEvent.EventHashSha256 = generateHashSHA256(sensorEvent)

	sensorEvent.EventReadAt = data.Metadata.ReadAt
	sensorEvent.EventSentAt = data.Metadata.SentAt
	sensorEvent.EventReceivedAt = data.Metadata.ReceivedAt

	sensorMetric := &pb.Metric{
		SnortTimestamp:     data.Timestamp,
		SnortBase64Data:    data.Base64Data,
		SnortClientBytes:   data.ClientBytes,
		SnortClientPkts:    data.ClientPkts,
		SnortDstAddress:    data.DstAddr,
		SnortDstPort:       data.DstPort,
		SnortDstAp:         data.DstAp,
		SnortEthDst:        data.EthDst,
		SnortEthLen:        data.EthLen,
		SnortEthSrc:        data.EthSrc,
		SnortEthType:       data.EthType,
		SnortFlowstartTime: data.FlowStartTime,
		SnortGeneveVni:     data.GeneveVNI,
		SnortIcmpCode:      data.ICMPCode,
		SnortIcmpId:        data.ICMPID,
		SnortIcmpSeq:       data.ICMPSeq,
		SnortIcmpType:      data.ICMPType,
		SnortIpId:          data.IPID,
		SnortIpLength:      data.IPLen,
		SnortMpls:          data.MPLS,
		SnortPktGen:        data.PktGen,
		SnortPktLength:     data.PktLen,
		SnortPktNumber:     data.PktNum,
		SnortServerBytes:   data.ServerBytes,
		SnortServerPkts:    data.ServerPkts,
		SnortSgt:           data.SGT,
		SnortSrcAddress:    data.SrcAddr,
		SnortSrcPort:       data.SrcPort,
		SnortSrcAp:         data.SrcAp,
		SnortTarget:        data.Target,
		SnortTcpAck:        data.TCPAck,
		SnortTcpFlags:      data.TCPFlags,
		SnortTcpLen:        data.TCPLen,
		SnortTcpSeq:        data.TCPSeq,
		SnortTcpWin:        data.TCPWin,
		SnortTimeToLive:    data.TTL,
		SnortUdpLength:     data.UDPLen,
		SnortVlan:          data.VLAN,
	}

	return sensorEvent, sensorMetric
}

// generateHashSHA256 generates a SHA256 hash from the attributes.
// It is used to identify the sensor event record.
func generateHashSHA256(payload *pb.SensorEvent) string {
	hash := sha256.Sum256([]byte(payload.String()))
	return hex.EncodeToString(hash[:])
}
