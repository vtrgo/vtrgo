// plc/plc.go
package plc

import (
	"github.com/danomagnum/gologix"
)

type PLC struct {
	client *gologix.Client
}

// NewPLC creates a new PLC instance with the specified IP address.
func NewPLC(ip string) *PLC {
	client := gologix.NewClient(ip)
	return &PLC{client: client}
}

// Connect establishes a connection to the PLC.
func (plc *PLC) Connect() error {
	return plc.client.Connect()
}

// Disconnect closes the connection to the PLC.
func (plc *PLC) Disconnect() {
	plc.client.Disconnect()
}

// ReadTagInt32 reads an int32 value from the specified tag.
func (plc *PLC) ReadTagInt32(tagName string) (int32, error) {
	var tagValue int32
	err := plc.client.Read(tagName, &tagValue)
	return tagValue, err
}

// WriteTagInt32 writes an int32 value to the specified tag.
func (plc *PLC) WriteTagInt32(tagName string, value int32) error {
	return plc.client.Write(tagName, value)
}

// ReadTagFloat32 reads a float32 value from the specified tag.
func (plc *PLC) ReadTagFloat32(tagName string) (float32, error) {
	var tagValue float32
	err := plc.client.Read(tagName, &tagValue)
	return tagValue, err
}

// WriteTagFloat32 writes a float32 value to the specified tag.
func (plc *PLC) WriteTagFloat32(tagName string, value float32) error {
	return plc.client.Write(tagName, value)
}

// ReadTagBool reads a boolean value from the specified tag.
func (plc *PLC) ReadTagBool(tagName string) (bool, error) {
	var tagValue bool
	err := plc.client.Read(tagName, &tagValue)
	return tagValue, err
}

// WriteTagBool writes a boolean value to the specified tag.
func (plc *PLC) WriteTagBool(tagName string, value bool) error {
	return plc.client.Write(tagName, value)
}

// ReadTagString reads a boolean value from the specified tag.
func (plc *PLC) ReadTagString(tagName string) (string, error) {
	var tagValue string
	err := plc.client.Read(tagName, &tagValue)
	return tagValue, err
}

// WriteTagString writes a boolean value to the specified tag.
func (plc *PLC) WriteTagString(tagName string, value string) error {
	return plc.client.Write(tagName, value)
}
