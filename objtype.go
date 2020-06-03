package main

import (
//	"fmt"
	"regexp"
	"strings"
	//    "fmt"
)
// named types that will be used through out the code
type rawDef [][]string                    // raw data read from .cfg files
type attrVal []string                     // nagois object attribute value
type def map[string]*attrVal              // nagios object definition
type defs map[string]def                  // nagios object definitons
type offset map[string][]string           // nagios object attribute offset

// nagios object definition struct
type obj struct {
    hostDefs                defs        // nagios host object definitions
    serviceDefs             defs        // nagios service object definitions
    hostgroupDefs           defs        // nagios hostgroup object definitions
    hostdependencyDefs      defs        // nagios hostdependency object definition
    servicedependencyDefs   defs        // nagios servicedependency object definition
    contactDefs             defs        // nagios contact object definition
    contactgroupDefs        defs        // nagios contactgroup object definition
    commandDefs             defs        // nagios command object definition
    hostTempDefs            defs        // nagios host template object definition
    serviceTempDefs         defs        // nagios service template object definition
    contactTempDefs         defs        // nagios contact template object definition
}

// nagios service obj struct
type serviceOffset struct {
    hostName                    Set      // offset of service that the host is associated with via host_name attr
    hostNameExcl                Set      // offset of excluded service that the host is not associated with via host_name attr
    hostgroupName               Set      // offset of service that the host is associated with via hostgroup_name attr
    hostgroupNameExcl           Set      // offset of excluded service that the host is not associated with via hostgroup_name attr
    use                         Set      // offset of service with only 'use' attr that have association via template attributes
    enabled                     Set      // offset of active services
    disabled                    Set      // offset of excluded services
    deleted                     attrVal  // deleted service
    others                      attrVal  // service definition that has name and service_description attributes ( useful for printHostInfo )
    tmpl                        templateOffset
}

type templateOffset struct {
    hostgroupName               offset      // offset of hostgroup delcared in service template via hostgroup_name attr
    hostgroupNameExcl           offset      // offset of excluded hostgroup declared in service template via hostgroup_name attr
    hostName                    offset      // offset of hostname declared in service template via host_name attr
    hostNameExcl                offset      // offset of excluded hostname declared in service template via host_name attr
    enabled                     attrVal     // active service template name
    disabled                    attrVal     // disabled service template name
    enabledDisabled             attrVal     // active and excluded service template name
    deleted                     attrVal     // deleted service template

}

// nagios hostgroup obj struct
//TODO: use set instead of offset map[string]{}
type hostgroupOffset struct{
    members                     Set      // offset of hostgroup that the host is a member of
    membersExcl                 Set      // offset of excluded hostgroup that the host is not a member of 
    hostgroupMembers            Set      // offset of hostgroup that the host is a member of 
    hostgroupMembersExcl        Set      // offset of excluded hostgroups that the host not member of
    templateHostgroups          Set      // offset of hostgroup (acvite and excluded) defined in host template
    enabled                     attrVal    // active hostgroups
    disabled                    attrVal    // excluded hostgroups
    enabledDisabled             attrVal     // active and excluded hgroups
    deleted                     attrVal     // deleted hostgroup object definition
}

// nagios host obj struct
type hostOffset struct{
    hostgroups                  offset      // hostgroups defined in host object definition
    hostgroupsExcl              offset      // hostgroups defined in host object definition
    templateHostgroups          offset      // hostgroups defined in host templates via 'use' attr
    templateHostgroupsExcl      offset      // hostgroups defined in host templates via 'use' attr
    host                        offset      // host location in hostDefs
    hostDef                     def         // host definition
    hostIndex                   string      // host index
    hostName                    string      // hostname
    hostAddr                    string      // host ip address
    hgrpEnabledName             []string    // 
    hgrpDisabledName            []string    //
    templateOrder               []string       // host template index order (so inheritance dont get messed up)
}

// objDict is used to hold a collection of hosts with their association
type objDict struct{
    hosts        hostOffset
    services     serviceOffset
    hostgroups   hostgroupOffset
}

// initialize objDict
func newObjDict() *objDict{
    o := &objDict{}
    o.hosts = *newHostOffset()
    o.services = *newServiceOffset()
    o.hostgroups = *newHostGroupOffset()
    return o
}

// obj constructor
func newObj() *obj {
    o := &obj{}
    o.hostDefs = make(defs)
    o.hostTempDefs = make(defs)
    o.hostgroupDefs = make(defs)
    o.hostdependencyDefs = make(defs)
    o.serviceDefs = make(defs)
    o.serviceTempDefs = make(defs)
    o.servicedependencyDefs = make(defs)
    o.commandDefs  = make(defs)
    o.contactDefs  = make(defs)
    o.contactTempDefs  = make(defs)
    o.contactgroupDefs  = make(defs)
    return o
}

// hostOffset constructor
func newHostOffset() *hostOffset{
	o := &hostOffset{}
	o.hostIndex = ""
	o.hostName = ""              // "" indicate host name does not exist
    o.templateHostgroups =        make(offset)
    o.templateHostgroupsExcl =    make(offset)
    o.hostgroups =                make(offset)
    o.hostgroupsExcl =            make(offset)
	o.host =                      make(offset)
    return o
}

// serviceOffset constructor
func newServiceOffset() *serviceOffset{
    o := &serviceOffset{
        hostName:                   *NewSet(),
        hostNameExcl:               *NewSet(),
        hostgroupName:              *NewSet(),
        hostgroupNameExcl:          *NewSet(),
        use:                        *NewSet(),
        tmpl: templateOffset{
            hostName:           make(offset),
            hostNameExcl:       make(offset),
            hostgroupName:      make(offset),
            hostgroupNameExcl:  make(offset),
        },
    }
    return o
}

// hostgroupOffset constructor
func newHostGroupOffset() *hostgroupOffset{
    o := &hostgroupOffset{}
    o.members =                *NewSet()
    o.membersExcl =            *NewSet() 
    o.hostgroupMembers =       *NewSet()     
    o.hostgroupMembersExcl =   *NewSet()     
    o.templateHostgroups =     *NewSet()         
    return o
}

func (o *obj) SetContactTempDefs(contactTempDef def) {
    ID := contactTempDef["name"].ToString()
    o.contactDefs[ID] = contactTempDef
}

func (o *obj) SetHostTempDefs(hostTempDef def) {
    ID := hostTempDef["name"].ToString()
    o.hostTempDefs[ID] = hostTempDef
}

func (o *obj) SetServiceTempDefs(serviceTempDef def) {
    ID := serviceTempDef["name"].ToString()
    o.serviceTempDefs[ID] = serviceTempDef
}

func (o *obj) SetHostDefs(hostDef def) {
    ID := hostDef["host_name"].ToString()
    o.hostDefs[ID] = hostDef
}

func (o *obj) SetHostGroupDefs(hostgroupDef def) {
    ID := hostgroupDef["hostgroup_name"].ToString()
    o.hostgroupDefs[ID] = hostgroupDef
}

func (o *obj) SetServiceDefs(serviceDef def) {
    ID := serviceDef["service_description"].ToString()
    o.serviceDefs[ID] = serviceDef
}

func (o *obj) SetContactDefs(contactDef def) {
    ID := contactDef["contact_name"].ToString()
    o.contactDefs[ID] = contactDef
}

func (o *obj) SetContactGroupDefs(contactgroupDef def) {
    ID := contactgroupDef["contactgroup_name"].ToString()
    o.contactgroupDefs[ID] = contactgroupDef
}

func (o *obj) SetcommandDefs(commandDef def) {
    ID := commandDef["command_name"].ToString()
    o.commandDefs[ID] = commandDef
}

func (o *obj) SetHostDependencyDefs(hostdependencyDef def, idx int) {
    o.hostdependencyDefs[string(idx)] = hostdependencyDef
}

// TODO:obj dependency does not have a unique ID need to hcange the data surct to []defs
func (o *obj) SetServiceDependencyDefs(servicedependencyDef def, idx int) {
    o.servicedependencyDefs[string(idx)] = servicedependencyDef
}

// hostgroupOffset Getters
//func (o *hostgroupOffset) GetMembersOffset() offset {
//    return o.members
//}
//
//func (o *hostgroupOffset) GetMembersExclOffset() offset {
//    return o.membersExcl
//}
//
//func (o *hostgroupOffset) GetHostgroupMembersOffset() offset {
//    return o.hostgroupMembers
//}
//
//func (o *hostgroupOffset) GetHostgroupMembersExclOffset() offset {
//    return o.hostgroupMembersExcl
//}

//func (o *hostgroupOffset) GetEnabledHostgroup() *attrVal {
//    return o.hgrpEnabled
//}
//
//func (o *hostgroupOffset) GetDisabledHostgroup() *attrVal {
//    return o.hgrpDisabled
//}
//
//func (o *hostgroupOffset) GetEnabledHostgroupName() attrVal {
//    return o.hgrpEnabledName
//}
//
//func (o *hostgroupOffset) GetDisabledHostgroupName() attrVal {
//    return o.hgrpDisabledName
//}
//func (o *hostgroupOffset) GetTemplateHostgroupsOffset() offset {
//    return o.templateHostgroups
//}
//// hostgroupOffset Setters
//func (o *hostgroupOffset) SetMembersOffset(id string, member string) {
//    o.members[id] = append(o.members[id], member)
//}
//
//func (o *hostgroupOffset) SetMembersExclOffset(id string, member string) {
//    o.membersExcl[id] = append(o.membersExcl[id], member)
//}
//
//func (o *hostgroupOffset) SetHostgroupMembersOffset(id string, member string) {
//    o.hostgroupMembers[id] = append(o.hostgroupMembers[id], member)
//}
//
//func (o *hostgroupOffset) SetHostgroupMembersExclOffset(id string, member string) {
//    o.hostgroupMembersExcl[id] = append(o.hostgroupMembersExcl[id], member)
//}
//
//func (o *hostgroupOffset) SetTemplateHostgroupsOffset(id string, hostgroup string) {
//    o.templateHostgroups[id] = append(o.templateHostgroups[id], hostgroup)
//}

//func (o *hostgroupOffset) SetDeletedHostgroup(hostgroup string) {
//    o.hgrpDeleted = append(o.hgrpDeleted, hostgroup)
//}
//func (o *hostgroupOffset) GetEnabledDisabledHostgroup() attrVal {
//    return o.hgrpEnabledDisabledName
//}
//
//func (o *hostgroupOffset) GetDeletedHostgroup() attrVal {
//    return o.hgrpDeleted
//}

func (o *hostgroupOffset) SetEnabledDisabledHostgroups() {
    hgrpEnabled := Union(&o.members, &o.hostgroupMembers)
    hgrpDisabled := Union(&o.membersExcl, &o.hostgroupMembersExcl)
    // Remove excluded hostgroups
    for item := range hgrpDisabled.m {
        if hgrpEnabled.Has(item) {
            hgrpEnabled.Remove(item)
        }
    }
    // convert into slice
    for item := range hgrpEnabled.m{
        o.enabled.Add(item)
    }
    for item := range hgrpDisabled.m{
        o.disabled.Add(item)
    }
    hgrpEnabledDisabled := Union(hgrpEnabled, hgrpDisabled)
    for item := range hgrpEnabledDisabled.m{
        o.enabledDisabled.Add(item)
    }

}

//func (o *hostgroupOffset) SetEnabledDisabledHostgroup() {
//    o.hgrpEnabledDisabledName = append(o.hgrpEnabledDisabledName, o.hgrpEnabledName...)
//    o.hgrpEnabledDisabledName = append(o.hgrpEnabledDisabledName, o.hgrpDisabledName...)
//}

func (o *hostOffset) SetEnabledHostgroups(hg *hostgroupOffset) {
    for _, v := range o.GetEnabledHostgroupsName(){
        if !hg.enabled.Has(v){
            hg.enabled.Add(v)
        }
    }
    for _, v := range o.GetDisabledHostgroupsName(){
        if hg.enabled.Has(strings.TrimLeft(v, "!")){
            hg.enabled.Remove(strings.TrimLeft(v, "!"))
        }
    }
}

// serviceOffset Setters
func (s *serviceOffset) Add(attr string, id string, val string) {
    switch attr {
    case "tmplHostName":
        s.tmpl.hostName[id] = append(s.tmpl.hostName[id], val)
    case "tmplHostNameExcl":
        s.tmpl.hostNameExcl[id] = append(s.tmpl.hostNameExcl[id], val)
    case "tmplHostgroupName":
        s.tmpl.hostgroupName[id] = append(s.tmpl.hostgroupName[id], val)
    case "tmplHostgroupNameExcl":
        s.tmpl.hostgroupNameExcl[id] = append(s.tmpl.hostgroupNameExcl[id], val)
    }


}

//func (o *serviceOffset) SetHybridService(svc string) {
//    o.others = append(o.others, svc)
//}

//func (o *serviceOffset) SetDeletedService(svc string) {
//    o.svcDeleted = append(o.svcDeleted, svc)
//}

func (o *serviceOffset) SetEnabledServiceTemplate() {
    enabledTemplate := attrVal{}
    for _, v := range o.tmpl.hostgroupName {
        for _,t := range v{
            if !enabledTemplate.Has(t) {
            enabledTemplate.Add(t)
            }
        }
    }
    for _, v := range o.tmpl.hostName {
        for _,t := range v{
            if !enabledTemplate.Has(t) {
            enabledTemplate.Add(t)
            }
        }
    }
    o.tmpl.enabled = enabledTemplate
    o.tmpl.enabledDisabled = append(o.tmpl.enabledDisabled, enabledTemplate...)
}

func (o *serviceOffset) SetDisabledServiceTemplate() {
    disabledTemplate := attrVal{}
    for _, v := range o.tmpl.hostgroupNameExcl {
        for _,t := range v{
            if !disabledTemplate.Has(t){
            disabledTemplate.Add(t)
            }
        }
    }
    for _, v := range o.tmpl.hostNameExcl {
        for _,t := range v{
            if !disabledTemplate.Has(t){
            disabledTemplate.Add(t)
            }
        }
    }
    o.tmpl.disabled = disabledTemplate
    for _, t := range disabledTemplate {
        if !o.tmpl.enabledDisabled.Has(t){
            o.tmpl.enabledDisabled = append(o.tmpl.enabledDisabled, t)
        }
    }
}
// set enabled and disabled service check of a specific host
func (o *serviceOffset) SetEnabledDisabledServices() {
    // services from host_name and hostgroup_name
    svcEnabled := Union(&o.hostName, &o.hostgroupName, &o.use)
    svcDisabled := Union(&o.hostNameExcl, &o.hostgroupNameExcl)
    // Remove excluded services
    for k := range svcDisabled.m {
        svcEnabled.Remove(k)
    }
    o.enabled = *svcEnabled
    o.disabled = *svcEnabled
}

// hostOffset Getters
func (h *hostOffset) GetHostOffset() string {
    return h.hostIndex
}

func (h *hostOffset) GetHostDefinition() def {
    return h.hostDef
}

func (h *hostOffset) GetTemplateHostgroups() offset {
    return h.templateHostgroups
}

func (h *hostOffset) GetHostgroups() offset {
    return h.hostgroups
}

func (h *hostOffset) GetHostName() string {
    return h.hostName
}
func (h *hostOffset) GetEnabledHostgroupsName() attrVal {
    return h.hgrpEnabledName
}

func (h *hostOffset) GetDisabledHostgroupsName() attrVal {
    return h.hgrpDisabledName
}

// hostOffset Setters
func (h *hostOffset) SetHostOffset(hostIndex string) {
    h.hostIndex = hostIndex
}

func (h *hostOffset) SetHostName(hostName string) {
    h.hostName = hostName
}

func (h *hostOffset) SetTemplateHostgroupsOffset(id string, hgrp *attrVal) {
    h.templateHostgroups[id] = append(h.templateHostgroups[id], *hgrp...)
}

func (h *hostOffset) SetTemplateOrder(id string) {
    h.templateOrder = append(h.templateOrder, id)
}

func (h *hostOffset) SetTemplateHostgroupsExclOffset(id string, hgrp *attrVal) {
    h.templateHostgroupsExcl[id] = append(h.templateHostgroupsExcl[id], *hgrp...)
}

func (h *hostOffset) SetHostgroupsOffset(id string, hgrp *attrVal) {
    h.hostgroups[id] = append(h.hostgroups[id], *hgrp...)
}

func (h *hostOffset) SetHostgroupsExclOffset(id string, hgrp *attrVal) {
    h.hostgroupsExcl[id] = append(h.hostgroupsExcl[id], *hgrp...)
}

func (h *hostOffset) SetHostDefinition(hDef def) {
    h.hostDef = hDef
}

func (h *hostOffset) GetHostIndex() offset {
    return h.host
}

func (h *hostOffset) SetHostIndex(id string, hostName string) {
    h.host[id] = append(h.host[id], hostName)
}

func (o *hostOffset) SetEnabledHostgroupsName() {
    // hostgroups declared in host definition
    for _, m := range o.hostgroups {
        for _, v := range m {
            if !strings.HasPrefix(v, "!"){
                o.hgrpEnabledName = append(o.hgrpEnabledName, v)
            }else {
                o.hgrpDisabledName = append(o.hgrpDisabledName, v)
            }
        }
    }

    // hostgroups delcared in host template support multiple inheritance
    // check for addative iheritance start with '+' e.g. +hostgroup
    if (o.hgrpEnabledName == nil && o.hgrpDisabledName == nil) || strings.HasPrefix(o.hgrpEnabledName[0], "+"){
        for _,i := range o.templateOrder {
            if !strings.HasPrefix(o.templateHostgroups[i][0], "+"){
                for _, v := range o.templateHostgroups[i] {
                    if !strings.HasPrefix(v, "!"){
                        o.hgrpEnabledName = append(o.hgrpEnabledName, v)
                    }else {
                        o.hgrpDisabledName = append(o.hgrpDisabledName, v)
                    }
                }
                break
            }
            for _, v := range o.templateHostgroups[i] {
                if !strings.HasPrefix(v, "!"){
                    o.hgrpEnabledName = append(o.hgrpEnabledName, v)
                }else {
                    o.hgrpDisabledName = append(o.hgrpDisabledName, v)
                }
            }
        }
    }
}

// Add items to a slice
func (s *attrVal) Add(items ...string){
    for _, item := range items {
        *s = append(*s, item)
    }
}

// Remove items from slice if exist
func (s *attrVal) Remove(items ...string){
    for i, v := range *s {
        for _, item := range items{
            if v == item {
                (*s)[i] = (*s)[len(*s)-1]
                (*s) = (*s)[:len(*s)-1]
            }
        }
    }
}

// Check if item exist in a slice
func (s *attrVal) Has(item string) bool{
    for _, v := range *s {
        if item ==  v {
            return true
        }
    }
    return false
}

func (s *attrVal) length() int {
    return len(*s)
}

// Check if item exist using regex
func (s *attrVal) RegexHas(item string) bool{
    for _,v := range *s {
        if has, _ := regexp.MatchString(v, item); has {
            return true
        }
    }
    return false
}

// Check if an item exist in both slices
func (s *attrVal) HasAny(items ...string) bool{
    for _,v := range *s {
        for _, item := range items{
            if v == item{
                return true
            }
        }
    }
    return false
}
