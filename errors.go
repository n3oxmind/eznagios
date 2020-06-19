package main

import (
    "fmt"
)

// parsing error
type parsingError struct {
    err error           // original error
}

// unknown Nagios object type error
type unknownObjectError struct {
    oDef def             // Nagios object definition
    oType string        // Nagios object type (host,service,...)
    err error           // original error

}

// Duplicate Nagios object attribute
type duplicateAttributeError struct {
    err      error          // original error
    objType  string         // Nagos object type (host,service,...)
    attrName string         // Nagios object attribute name
    oDef     def            // Nagios object definition
    dupAttr  def            // Duplicate attribute

}

// object not found error
type NotFoundError struct {
    err error           // what happen
    errType string      // error type warn,fatal,error,info
    value string        // object value 
}

// unknown object error format
func (e *unknownObjectError) Error() string {
    fDef := formatAttr(e.oDef)
    return fmt.Sprintf("UnknownObject: %vWarning%v: %v '%v'\n%v\n%v}",Yellow,RST,e.err,e.oType,e.oType,fDef)
}

// unknown object error format
func (e *duplicateAttributeError) Error() string {
    fDef := formatAttr(e.oDef)
    dAttr := formatAttr(e.dupAttr)
    return fmt.Sprintf("DuplicateAttribute: %vInfo%v: %v '%v'\n%v\n%v%v}",Info,RST,e.err,e.attrName,e.objType,dAttr,fDef)
}

// object not found error format
func (e *NotFoundError) Error() string {
    if e.errType == "Warn" {
        return fmt.Sprintf("NotFound: %vWarn%v: %v '%v'",Yellow, RST, e.err, e.value)
    } else {
        return fmt.Sprintf("NotFound: %v%v%v: '%v'",Red,RST, e.err, e.value)
    }
}

// parsing error format
func (e *parsingError) Error() string {
    return fmt.Sprintf("ArgsParsing: %vError%v: %v", Red, RST, e.err)
}
