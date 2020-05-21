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

// Nagios object not found error 
type objectNotFoundError struct {
    err error
}
// Nagios object definition not found
type DefNotFoundError struct {
    err error
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
func (e *objectNotFoundError) Error() string {
    return fmt.Sprintf("ObjectNotFound: %vFatal%v: %v",Fatal,RST,e.err)
}
// definition not found error format
func (e *DefNotFoundError) Error() string {
    return fmt.Sprintf("ObjectNotFound: %vWarn%v: %v",Warn,RST,e.err)
}

// parsing error format
func (e *parsingError) Error() string {
    return fmt.Sprintf("ArgsParsing: %vError%v: %v", Red, RST, e.err)
}
