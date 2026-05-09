package client

import "testing"

func TestParseSamplesParsesPrefixedFirstRecordAndEnumValues(t *testing.T) {
	raw := "3021,AC_1,ENP_AC_SRVII[COM]^2,Return air temperature measurement,28.600000,℃,1778314683,0,1,1,2,2;" +
		"3,Return air humidity measurement,30.400000,%,1778314683,0,1,1,2,2;" +
		"5,Air conditioning operation status,Running[0],,1778314683,0,1,1,0,5;" +
		"261,Compressor output,OutputMove[1],,1778314683,0,1,1,0,5;"

	samples, err := ParseSamples(raw)
	if err != nil {
		t.Fatalf("ParseSamples returned error: %v", err)
	}

	if got := samples[2].Value; got != 28.6 {
		t.Fatalf("field 2 value = %v, want 28.6", got)
	}
	if got := samples[2].Name; got != "Return air temperature measurement" {
		t.Fatalf("field 2 name = %q", got)
	}
	if got := samples[3].Value; got != 30.4 {
		t.Fatalf("field 3 value = %v, want 30.4", got)
	}
	if got := samples[5].Value; got != 0 {
		t.Fatalf("field 5 enum value = %v, want 0", got)
	}
	if got := samples[261].Value; got != 1 {
		t.Fatalf("field 261 enum value = %v, want 1", got)
	}
}

func TestParseSamplesRejectsEmptyPayload(t *testing.T) {
	if _, err := ParseSamples("   "); err == nil {
		t.Fatal("ParseSamples returned nil error for empty payload")
	}
}

func TestParseSamplesSupportsP101Format(t *testing.T) {
	raw := "Y|32~700~进/回风温差;700;0,2@5.3|33~700~回风温度;700;0,157@24.6|261~700~Compressor output;700;0,1@1|"

	samples, err := ParseSamples(raw)
	if err != nil {
		t.Fatalf("ParseSamples returned error: %v", err)
	}

	if got := samples[32].Value; got != 5.3 {
		t.Fatalf("field 32 value = %v, want 5.3", got)
	}
	if got := samples[33].Value; got != 24.6 {
		t.Fatalf("field 33 value = %v, want 24.6", got)
	}
	if got := samples[261].Value; got != 1 {
		t.Fatalf("field 261 value = %v, want 1", got)
	}
}
