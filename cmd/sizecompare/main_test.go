package main

import "testing"

func BenchmarkSerializeToJSON(b *testing.B) {
	for b.Loop() {
		serializeToJson(metadata)
	}
}

func BenchmarkSerializeToXML(b *testing.B) {
	for b.Loop() {
		serializeToXML(metadata)
	}
}

func BenchmarkSerializeToProto(b *testing.B) {
	for b.Loop() {
		serializeToProto(genMetadata)
	}
}
