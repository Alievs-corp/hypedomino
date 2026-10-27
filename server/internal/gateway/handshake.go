package gateway

import (
	"strconv"
	"strings"

	pb "github.com/Alievs-corp/hypedomino/server/internal/pb/hypedomino/v1"
)

func VersionAtLeast(version, minimum string) bool {
	have, ok := parseVersion(version)
	if !ok {
		return false
	}
	want, ok := parseVersion(minimum)
	if !ok {
		return true
	}
	for index := range have {
		if have[index] != want[index] {
			return have[index] > want[index]
		}
	}
	return true
}

func parseVersion(text string) ([3]int, bool) {
	var parts [3]int
	fields := strings.SplitN(strings.TrimPrefix(text, "v"), ".", 4)
	if len(fields) > 3 {
		return parts, false
	}
	for index, field := range fields {
		value, err := strconv.Atoi(field)
		if err != nil || value < 0 {
			return parts, false
		}
		parts[index] = value
	}
	return parts, true
}

func (c *Conn) serverHello(status pb.HelloStatus, acked uint64) *pb.Envelope {
	return &pb.Envelope{
		ProtocolVersion: ProtocolVersion,
		AckedSequence:   acked,
		Payload: &pb.Envelope_ServerHello{ServerHello: &pb.ServerHello{
			Status:            status,
			MinimumAppVersion: c.options.Config.MinimumAppVersion,
			ServerTimeUnixMs:  uint64(c.options.Clock.Now().UnixMilli()),
			CatalogVersion:    c.options.Config.CatalogVersion,
		}},
	}
}

func errorFrame(code pb.ErrorCode, rejected, acked uint64) *pb.Envelope {
	return &pb.Envelope{
		ProtocolVersion: ProtocolVersion,
		AckedSequence:   acked,
		Payload:         &pb.Envelope_Error{Error: &pb.Error{Code: code, RejectedSequence: rejected}},
	}
}
