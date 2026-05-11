package mqtt

import "testing"

func TestParseAckPayload_LegacyCommand(t *testing.T) {
	payload := []byte(`{"command_id":123,"ack_code":"OK","ack_message":"ok","ack_payload":{"k":"v"}}`)
	got, err := ParseAckPayload(payload)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if got.Kind != "legacy_command" {
		t.Fatalf("expected legacy_command, got %s", got.Kind)
	}
	if got.Legacy == nil || got.Legacy.CommandID != 123 {
		t.Fatalf("unexpected legacy payload")
	}
}

func TestParseAckPayload_V1Config(t *testing.T) {
	payload := []byte(`{"schema_version":1,"ack_type":"config","msg_id":"m1","trace_id":"t1","result":"ACKED","error_code":"OK","error_message":"","device_ts_ms":1710000000000,"payload":{"fw_version":"v1"}}`)
	got, err := ParseAckPayload(payload)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if got.Kind != "v1_config" {
		t.Fatalf("expected v1_config, got %s", got.Kind)
	}
	if got.V1 == nil || got.V1.MsgID != "m1" {
		t.Fatalf("unexpected v1 payload")
	}
}

func TestParseAckPayload_V1Command(t *testing.T) {
	payload := []byte(`{"schema_version":1,"ack_type":"command","msg_id":"m1","trace_id":"t1","result":"ACKED","payload":{"command_id":10,"ack_code":"OK","ack_message":"ok","ack_payload":{}}}`)
	got, err := ParseAckPayload(payload)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if got.Kind != "v1_command" {
		t.Fatalf("expected v1_command, got %s", got.Kind)
	}
	if got.V1 == nil || got.V1.AckType != "command" {
		t.Fatalf("unexpected v1 command payload")
	}
}
