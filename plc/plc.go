// plc/plc.go
package plc

import (
	"fmt"
	"log"

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
func (plc *PLC) ReadTrigger(tagName string) (bool, error) {
	var tagValue bool
	err := plc.client.Read(tagName, &tagValue)
	return tagValue, err
}

// WriteTagBool writes a boolean value to the specified tag.
func (plc *PLC) WriteResponse(tagName string, value bool) error {
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

// ReadTag reads a value from the specified tag.
func (plc *PLC) ReadTag(tagName string, tagType string, length int) (any, error) {
	var err error
	switch tagType {
	case "bool":
		var tagValue bool
		err := plc.client.Read(tagName, &tagValue)
		return tagValue, err
	case "int":
		var tagValue int16
		err := plc.client.Read(tagName, &tagValue)
		return tagValue, err
	case "dint":
		var tagValue int32
		err := plc.client.Read(tagName, &tagValue)
		return tagValue, err
	case "real":
		var tagValue float32
		err := plc.client.Read(tagName, &tagValue)
		return tagValue, err
	case "string":
		var tagValue string
		err := plc.client.Read(tagName, &tagValue)
		return tagValue, err
	case "[]dint":
		values := make([]int32, length)

		// Read each element individually
		for i := 0; i < length; i++ {
			elementName := fmt.Sprintf("%s[%d]", tagName, i)
			value, err := plc.ReadTag(elementName, "dint", 0)
			if err != nil {
				return nil, fmt.Errorf("problem reading element %d of %s: %v", i, tagName, err)
			}

			// Ensure the value is of type int32
			intValue, ok := value.(int32)
			if !ok {
				return nil, fmt.Errorf("element %d of %s has incorrect type: %T", i, tagName, value)
			}

			values[i] = intValue
		}

		return values, nil
	case "[]real":
		values := make([]float32, length)

		// Read each element individually
		for i := 0; i < length; i++ {
			elementName := fmt.Sprintf("%s[%d]", tagName, i)
			value, err := plc.ReadTag(elementName, "real", 0)
			if err != nil {
				return nil, fmt.Errorf("problem reading element %d of %s: %v", i, tagName, err)
			}

			// Ensure the value is of type int32
			realValue, ok := value.(float32)
			if !ok {
				return nil, fmt.Errorf("element %d of %s has incorrect type: %T", i, tagName, value)
			}

			values[i] = realValue
		}

		return values, nil
	default:
		log.Printf("Incorrect data type, %v", err)
	}
	return nil, err
}

func (plc *PLC) WriteTag(tagName string, tagType string, tagValue interface{}) (any, error) {
	var err error
	switch tagType {
	case "bool":
		err := plc.client.Write(tagName, tagValue.(bool))
		return tagValue, err
	case "int":
		err := plc.client.Write(tagName, tagValue.(int16))
		return tagValue, err
	case "dint":
		err := plc.client.Write(tagName, tagValue.(int32))
		return tagValue, err
	case "real":
		err := plc.client.Write(tagName, tagValue.(float32))
		return tagValue, err
	case "string":
		err := plc.client.Write(tagName, tagValue.(string))
		return tagValue, err
	default:
		log.Printf("Incorrect data type, %v", err)
	}
	return nil, err
}

// Custom function to read an array of int32 from the PLC
// func (plc *PLC) ReadInt32Array(tagName string, length int) ([]int32, error) {
// 	// Create a slice to hold the array values
// 	values := make([]int32, length)

// 	// Read each element individually
// 	for i := 0; i < length; i++ {
// 		elementName := fmt.Sprintf("%s[%d]", tagName, i)
// 		value, err := plc.ReadTag(elementName, "dint")
// 		if err != nil {
// 			return nil, fmt.Errorf("problem reading element %d of %s: %v", i, tagName, err)
// 		}

// 		// Ensure the value is of type int32
// 		intValue, ok := value.(int32)
// 		if !ok {
// 			return nil, fmt.Errorf("element %d of %s has incorrect type: %T", i, tagName, value)
// 		}

// 		values[i] = intValue
// 	}

// 	return values, nil
// }
