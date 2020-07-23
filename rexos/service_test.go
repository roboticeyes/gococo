package rexos

import "testing"

func TestUrlParse(t *testing.T) {
	if GetGUIDFromRexTagURL("hugo") != "" {
		t.Fatal("Wrong response")
	}
	if GetGUIDFromRexTagURL("http://hugo") != "" {
		t.Fatal("Wrong response")
	}
}

func TestUrlParseWithoutType(t *testing.T) {
	if GetGUIDFromRexTagURL("https://rex.codes/v1/xxx") != "xxx" {
		t.Fatal("Wrong response")
	}
}

func TestUrlParseWithType(t *testing.T) {
	if GetGUIDFromRexTagURL("https://rex.codes/v1/xxx?type=stub") != "xxx" {
		t.Fatal("Wrong response")
	}
}

func TestUrlParseDev(t *testing.T) {
	link := "https://dev-01.rex.codes/v1/62b34cec-47de-cd3c-4bff-b599166e8a04"
	res := "62b34cec-47de-cd3c-4bff-b599166e8a04"
	if GetGUIDFromRexTagURL(link) != res {
		t.Fatal("Wrong response")
	}
}

func TestUrlParseProduction(t *testing.T) {
	link := "https://rex.codes/v1/62b34cec-47de-cd3c-4bff-b599166e8a04"
	res := "62b34cec-47de-cd3c-4bff-b599166e8a04"
	if GetGUIDFromRexTagURL(link) != res {
		t.Fatal("Wrong response")
	}
}

func TestUrlParseProductionWithQuery(t *testing.T) {
	link := "https://rex.codes/v1/62b34cec-47de-cd3c-4bff-b599166e8a04?type=stub"
	res := "62b34cec-47de-cd3c-4bff-b599166e8a04"
	if GetGUIDFromRexTagURL(link) != res {
		t.Fatal("Wrong response")
	}
}
