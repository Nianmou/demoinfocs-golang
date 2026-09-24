package demoinfocs

import (
	"math"
	"slices"
	"testing"

	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	msg "github.com/markus-wa/demoinfocs-golang/v6/pkg/demoinfocs/msg"
)

func TestUserCmdRecordFieldsListsNestedPaths(t *testing.T) {
	full := &msg.CSGOUserCmdPB{Base: &msg.CBaseUserCmdPB{
		Viewangles:  &msg.CMsgQAngle{X: proto.Float32(1.5), Y: proto.Float32(-2.5)},
		Forwardmove: proto.Float32(3),
	}}

	data, err := proto.Marshal(full)
	if err != nil {
		t.Fatal(err)
	}

	got := userCmdRecordFields(data)
	for _, want := range []string{"base", "base.viewangles", "base.viewangles.x",
		"base.viewangles.y", "base.forwardmove"} {
		if !slices.Contains(got, want) {
			t.Fatalf("缺 %q，实际 %v", want, got)
		}
	}
}

// delta 里认不出的字段号必须显式记成 "#号"，否则「没解析出来」会被读成「没写」；
// wire type 7 的复位标记要落在名字上（后缀 !）。
func TestUserCmdRecordFieldsMarksUnknownAndReset(t *testing.T) {
	desc := (&msg.CSGOUserCmdPB{}).ProtoReflect().Descriptor()
	baseFD := fieldByName(t, desc, "base")
	nested := baseFD.Message()
	forward := fieldByName(t, nested, "forwardmove")
	mousedx := fieldByName(t, nested, "mousedy")

	var inner []byte
	inner = protowire.AppendTag(inner, forward.Number(), protowire.Fixed32Type)
	inner = protowire.AppendFixed32(inner, math.Float32bits(4.5))
	inner = protowire.AppendTag(inner, mousedx.Number(), wireTypeReset)
	inner = protowire.AppendTag(inner, protowire.Number(999), protowire.VarintType)
	inner = protowire.AppendVarint(inner, 1)

	var data []byte
	data = protowire.AppendTag(data, baseFD.Number(), protowire.BytesType)
	data = protowire.AppendBytes(data, inner)

	got := userCmdRecordFields(data)
	want := []string{"base", "base.forwardmove", "base.mousedy!", "base.#999"}
	if !slices.Equal(got, want) {
		t.Fatalf("实际 %v，应为 %v", got, want)
	}
}

func fieldByName(t *testing.T, desc protoreflect.MessageDescriptor, name string) protoreflect.FieldDescriptor {
	t.Helper()

	fd := desc.Fields().ByName(protoreflect.Name(name))
	if fd == nil {
		t.Fatalf("%s 上没有字段 %q", desc.FullName(), name)
	}

	return fd
}
