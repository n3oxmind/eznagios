package main

import (
    "strings"
)
// named types that will be used through out the code
type rawDef [][]string                  // raw data read from .cfg files
type def map[string]*set                // nagios object definition
type defs []def                         // nagios object definitons
type offset map[int][]string              // nagios object attribute offset

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
    hostName                    offset      // service that the host is associated with via host_name attr
    hostNameExcl                offset      // excluded service that the host is not associated with via host_name attr
    hostgroupName               offset      // service that the host is associated with via hostgroup_name attr
    hostgroupNameExcl           offset      // excluded service that the host is not associated with via hostgroup_name attr
    use                         offset      // service with only 'use' attr that have association via template attributes
    templateHostgroupName       offset      // hostgroup located in service template via hostgroup_name attr
    templateHostgroupNameExcl   offset      // excluded hostgroup located in service template via hostgroup_name attr
    templateHostName            offset      // hostname located in service template via host_name attr
    templateHostNameExcl        offset      // excluded hostname located in service template via host_name attr
    svcEnabled                  offset      // active services
    svcDisabled                 offset      // excluded services
    tmplEnabledName             []string    // active service template name
    svcEnabledName              []string    // active services names that belong to a specific host
    svcDisabledName             []string    // excluded services names that belong to a specific host
//    *serviceInheritance                     // inherited struct (embedding)
}

// nagios service template obj struct
// template inheritence will be represented as a slice data structure e.g.:
// [D,N,M] --> means D inherit from N, N inerhrit from M          (multi values mean inheritance exist)
// [F]  --> means template F does not inherit from any other template (single value mean inheritance does not exist)
//type serviceInheritance struct {
//    hostgroupName       [][]string      // hostgroup located in service template via hostgroup_name attr
//    hostgroupNameExcl   [][]string      // excluded hostgroup located in service template via hostgroup_name attr
//    hostName            [][]string      // hostname located in service template via host_name attr
//    hostNameExcl        [][]string      // excluded hostname located in service template via host_name attr
//}

// single service definition inheritance
type serviceTemplate struct {
  hostgroupName       []string               // hostgroup_names from  nested inheritance for a single def
  hostgroupNameExcl   []string               // excluded hostgroup_names from nested inhertitance for a single def
  hostName            []string               // host_names from nested inheritacne for a single def
  hostNameExcl        []string               // excluded host_names from nested inheritance for a single def
}

func (ih *serviceTemplate) Clear() {
    ih.hostgroupName = nil
    ih.hostgroupNameExcl = nil
    ih.hostName = nil
    ih.hostNameExcl= nil
}

// nagios hostgroup obj struct
type hostgroupOffset struct {
    members              offset   // hostgroup that the host member of via members attr
    membersExcl          offset   // excluded hostgroup that the host not member of via members attr
    hostgroupMembers     offset   // hostgroup_name that the host member of via hostgroup_name attr
    hostgroupMembersExcl offset   // excluded hostgroups_name that the host not member of via hostgroup_name attr
    templateHostgroups   offset   // hostgroup (acvite and excluded) defined in host template via hostgroups attr
    hgrpEnabled          *set     // active hostgroups
    hgrpDisabled         *set     // excluded hostgroups
    hgrpEnabledName      []string // hostgroups names that associated with a specific host
    hgrpDisabledName     []string // excluded hostgroups names htat associated with a specific host
}

// nagios host obj struct
type hostOffset struct {
    host               offset   // host location in hostDefs
    hostDef            def      // host definition
    hostIndex          int      // host index
    hostName           string   // hostname
    templateHostgroups      offset // hostgroups defined in host templates via 'use' attr
    templateHostgroupsExcl  offset // hostgroups defined in host templates via 'use' attr
    hostgroups              offset // hostgroups defined in host object definition
    hostgroupsExcl          offset // hostgroups defined in host object definition
    hgrpEnabledName         []string
}
// hostOffset constructor
func newHostOffset() *hostOffset {
	o := &hostOffset{}
	o.host = make(offset)
	o.hostIndex = -1                    // -1 indicate host index does not exist
	o.hostName = ""                     // "" indicate host name does not exist
    o.templateHostgroups = make(offset)
    o.templateHostgroupsExcl = make(offset)
    o.hostgroups = make(offset)
    o.hostgroupsExcl = make(offset)
    return o
}

// serviceOffset constructor
func newServiceOffset() *serviceOffset {
    o := &serviceOffset{
        hostName:           make(offset),
        hostNameExcl:       make(offset),
        hostgroupName:      make(offset),
        hostgroupNameExcl:  make(offset),
        use:                make(offset),
        templateHostName: make(offset),
        templateHostNameExcl: make(offset),
        templateHostgroupName: make(offset),
        templateHostgroupNameExcl: make(offset),
//        serviceInheritance: &serviceInheritance{},
    }
    return o
}

// hostgroupOffset constructor
func newHostGroupOffset() *hostgroupOffset {
    o := &hostgroupOffset{}
    o.members = make(offset)
    o.membersExcl = make(offset)
    o.hostgroupMembers = make(offset)
    o.hostgroupMembersExcl = make(offset)
    o.templateHostgroups = make(offset)
    o.hgrpEnabled =  NewSet()
    o.hgrpDisabled = NewSet()
    return o
}

// Getter and Setter for obj type
func (o *obj) ContactTempDefs() defs {
    return o.contactTempDefs
}

func (o *obj) ServiceTempDefs() defs {
    return o.serviceTempDefs
}

func (o *obj) HostTempDefs() defs {
    return o.hostTempDefs
}

func (o *obj) ServiceDefs() defs {
    return o.serviceDefs
}
func (o *obj) HostDefs() defs {
    return o.hostDefs
}

func (o *obj) SetContactTempDefs(contactTempDef def) {
    o.contactTempDefs = append(o.contactTempDefs,contactTempDef)
}

func (o *obj) SetHostTempDefs(hostTempDef def) {
    o.hostTempDefs = append(o.hostTempDefs,hostTempDef)
}

func (o *obj) SetServiceTempDefs(serviceTempDef def) {
    o.serviceTempDefs = append(o.serviceTempDefs,serviceTempDef)
}

func (o *obj) SetHostDefs(hostDef def) {
    o.hostDefs = append(o.hostDefs, hostDef)
}

func (o *obj) SetHostGroupDefs(hostgroupDef def) {
    o.hostgroupDefs = append(o.hostgroupDefs, hostgroupDef)
}

func (o *obj) SetServiceDefs(serviceDef def) {
    o.serviceDefs = append(o.serviceDefs, serviceDef)
}

func (o *obj) SetContactDefs(contactDef def) {
    o.contactDefs = append(o.contactDefs, contactDef)
}

func (o *obj) SetContactGroupDefs(contactgoupDef def) {
    o.contactgroupDefs = append(o.contactgroupDefs, contactgoupDef)
}

func (o *obj) SetcommandDefs(commandDef def) {
    o.commandDefs = append(o.commandDefs, commandDef)
}

func (o *obj) SetHostDependencyDefs(hostdependencyDef def) {
    o.hostdependencyDefs = append(o.hostdependencyDefs, hostdependencyDef)
}

func (o *obj) SetServiceDependencyDefs(servicedependencyDef def) {
    o.servicedependencyDefs = append(o.servicedependencyDefs, servicedependencyDef)
}

// hostgroupOffset Getters and Setters
func (o *hostgroupOffset) GetMembersOffset() offset {
    return o.members
}

func (o *hostgroupOffset) GetMembersExclOffset() offset {
    return o.membersExcl }
func (o *hostgroupOffset) GetHostgroupMembersOffset() offset {
    return o.hostgroupMembers
}

func (o *hostgroupOffset) GetHostgroupMembersExclOffset() offset {
    return o.hostgroupMembersExcl
}

func (o *hostgroupOffset) GetEnabledHostgroup() *set {
    return o.hgrpEnabled
}

func (o *hostgroupOffset) GetDisabledHostgroup() *set {
    return o.hgrpDisabled
}

func (o *hostgroupOffset) GetEnabledHostgroupName() []string {
    return o.hgrpEnabledName
}

func (o *hostgroupOffset) GetDisabledHostgroupName() []string {
    return o.hgrpDisabledName
}
func (o *hostgroupOffset) GetTemplateHostgroupsOffset() offset {
    return o.templateHostgroups
}

func (o *hostgroupOffset) SetMembersOffset(i int, member string) {
    o.members[i] = append(o.members[i], member)
}

func (o *hostgroupOffset) SetMembersExclOffset(i int, member string) {
    o.membersExcl[i] = append(o.membersExcl[i], member)
}

func (o *hostgroupOffset) SetHostgroupMembersOffset(i int, member string) {
    o.hostgroupMembers[i] = append(o.hostgroupMembers[i], member)
}

func (o *hostgroupOffset) SetHostgroupMembersExclOffset(i int, member string) {
    o.hostgroupMembersExcl[i] = append(o.hostgroupMembersExcl[i], member)
}

func (o *hostgroupOffset) SetTemplateHostgroupsOffset(i int, hostgroup string) {
    o.templateHostgroups[i] = append(o.templateHostgroups[i], hostgroup)
}

func (o *serviceOffset) SetTemplateHostgroupNameOffset(i int, hostgroup string) {
    o.templateHostgroupName[i] = append(o.templateHostgroupName[i], hostgroup)
}
func (o *serviceOffset) SetTemplateHostgroupNameExclOffset(i int, hostgroup string) {
    o.templateHostgroupNameExcl[i] = append(o.templateHostgroupNameExcl[i], hostgroup)
}
func (o *serviceOffset) SetTemplateHostNameOffset(i int, hostname string) {
    o.templateHostName[i] = append(o.templateHostName[i], hostname)
}
func (o *serviceOffset) SetTemplateHostNameExclOffset(i int, hostname string) {
    o.templateHostNameExcl[i] = append(o.templateHostNameExcl[i], hostname)
}


func (o *hostgroupOffset) SetEnabledHostgroup() {
    for _, m := range o.GetMembersOffset() {
        for _, v := range m{
            o.hgrpEnabled.Add(v)
        }
    }
    for _, m := range o.GetHostgroupMembersOffset() {
        for _, v := range m{
            o.hgrpEnabled.Add(v)
        }
    }
    // Remove excluded hostgroups
    for _, m := range o.GetHostgroupMembersExclOffset() {
        for _, v := range m{
            o.hgrpEnabled.Remove(v)
        }
    }
    for _, m := range o.GetMembersExclOffset() {
        for _, v := range m{
            o.hgrpEnabled.Remove(v)
        }
    }
//    o.hgrpEnabledName = o.hgrpEnabled.StringSlice()
}
func (o *hostOffset) SetEnabledHostgroups(hg *hostgroupOffset) {
    for _, m := range o.templateHostgroups {
        for _, v := range m {
            hgrpName := strings.Split(v, ",")
            for _,item := range hgrpName {
                if strings.HasPrefix(item,"!"){
                    hg.hgrpEnabled.Remove(item)
                }
                hg.hgrpEnabled.Add(item)
            }
        }
    }
    hg.hgrpEnabledName = hg.hgrpEnabled.StringSlice()
}

func (o *hostgroupOffset) SetDisabledHostgroup() {
    for _,m := range o.GetMembersExclOffset(){
        for _, v := range m{
            o.hgrpDisabled.Add(v)
        }
    }
    for _,m := range o.GetHostgroupMembersExclOffset() {
        for _, v := range m{
            o.hgrpDisabled.Add(v)
        }
    }
    o.hgrpDisabledName = o.hgrpDisabled.StringSlice()
}

// serviceOffset Getters and Setters
func (o *serviceOffset) GetTemplateHostgroupNameOffset() offset {
    return o.templateHostgroupName
}
func (o *serviceOffset) GetTemplateHostgroupNameExclOffset() offset {
    return o.templateHostgroupNameExcl
}
func (o *serviceOffset) GetTemplateHostNameOffset() offset {
    return o.templateHostName
}
func (o *serviceOffset) GetTemplateHostNameExclOffset() offset {
    return o.templateHostNameExcl
}

func (o *serviceOffset) GetHostNameOffset() offset {
    return o.hostName
}

func (o *serviceOffset) GetHostNameExclOffset() offset {
    return o.hostNameExcl
}

func (o *serviceOffset) GetHostgroupNameOffset() offset {
    return o.hostgroupName
}

func (o *serviceOffset) GetHostgroupNameExclOffset() offset {
    return o.hostgroupNameExcl
}

func (o *serviceOffset) GetEnabledServiceOffset() offset {
    return o.svcEnabled
}

func (o *serviceOffset) GetDisabledServiceOffset() offset {
    return o.svcDisabled
}

func (o *serviceOffset) GetUseOffset() offset {
    return o.use
}

func (o *serviceOffset) GetEnabledServiceName() []string {
    return o.svcEnabledName
}
func (o *serviceOffset) GetEnabledTemplateName() []string {
    return o.tmplEnabledName
}

func (o *serviceOffset) GetDisabledServiceName() []string {
    return o.svcDisabledName
}

func (o *serviceOffset) SetHostNameOffset(i int, member string) {
    o.hostName[i] = append(o.hostName[i], member)
}

func (o *serviceOffset) SetHostNameExclOffset(i int, member string) {
    o.hostNameExcl[i] = append(o.hostNameExcl[i], member)
}

func (o *serviceOffset) SetHostgroupNameOffset(i int, member string) {
    o.hostgroupName[i] = append(o.hostgroupName[i], member)
}

func (o *serviceOffset) SetHostgroupNameExclOffset(i int, member string) {
    o.hostgroupNameExcl[i] = append(o.hostgroupNameExcl[i], member)
}
func (o *serviceOffset) SetUseOffset(i int, tmpl string) {
    o.use[i] = append(o.use[i], tmpl)
}

// set active service templates
func (o *serviceOffset) SetEnabledTemplate() {
    enabledTemplate := NewSet()
    for _, v := range o.templateHostgroupName {
        for _,t := range v{
            enabledTemplate.Add(t)
        }
    }
    for _, v := range o.templateHostName {
        for _,t := range v{
            enabledTemplate.Add(t)
        }
    }
    o.tmplEnabledName = enabledTemplate.StringSlice()
}
// set active service checks
func (o *serviceOffset) SetEnabledService() {
    // copy service assocaition via host_name
    svcEnabled := CopyMapInt(o.GetHostNameOffset())
    // add service association via hostgroup_name
    for k, v := range o.GetHostgroupNameOffset() {
        (*svcEnabled)[k] = v
    }
    // add service association via use directive (template)
    for k, v := range o.use {
        (*svcEnabled)[k] = v
    }
    // Remove excluded service declared in host_name
    for k, _ := range o.GetHostNameExclOffset() {
        if _,exist := (*svcEnabled)[k]; exist {
            delete(*svcEnabled, k)
        }
    }
    // Remove excluded service declared in hostgroup_name
    for k, _ := range o.GetHostgroupNameExclOffset() {
        if _,exist := (*svcEnabled)[k]; exist {
            delete(*svcEnabled, k)
        }
    }
    o.svcEnabled = *svcEnabled
    _,o.svcEnabledName = o.svcEnabled.ToSlice()
    
}

// set disabled service checks
func (o *serviceOffset) SetDisabledService() {
    svcDisabled := CopyMapInt(o.GetHostNameExclOffset())
    for k, v := range o.GetHostgroupNameExclOffset() {
        (*svcDisabled)[k] = v
    }
    o.svcDisabled = *svcDisabled
    _, o.svcDisabledName = o.svcDisabled.ToSlice()
}


func (h *hostOffset) GetHostOffset() int {
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

func (h *hostOffset) SetHostOffset(hostIndex int) {
    h.hostIndex = hostIndex
}

func (h *hostOffset) GetHostName() string {
    return h.hostName
}
func (h *hostOffset) GetEnabledHostgroupsName() []string {
    return h.hgrpEnabledName
}

func (h *hostOffset) SetHostName(hostName string) {
    h.hostName = hostName
}

func (h *hostOffset) SetTemplateHostgroupsOffset(idx int, hgrp *set) {
    h.templateHostgroups[idx] = append(h.templateHostgroups[idx], hgrp.StringSlice()...)
}

func (h *hostOffset) SetTemplateHostgroupsExclOffset(idx int, hgrp *set) {
    h.templateHostgroupsExcl[idx] = append(h.templateHostgroupsExcl[idx], hgrp.StringSlice()...)
}
func (h *hostOffset) SetHostgroupsOffset(idx int, hgrp *set) {
    h.hostgroups[idx] = append(h.hostgroups[idx], hgrp.StringSlice()...)
}
func (h *hostOffset) SetHostgroupsExclOffset(idx int, hgrp *set) {
    h.hostgroupsExcl[idx] = append(h.hostgroupsExcl[idx], hgrp.StringSlice()...)
}

func (h *hostOffset) SetHostDefinition(hDef def) {
    h.hostDef = hDef
}

func (h *hostOffset) GetHostIndex() offset {
    return h.host
}

func (h *hostOffset) SetHostIndex(idx int, hostName string) {
    h.host[idx] = append(h.host[idx], hostName)
}

func (o *hostOffset) SetEnabledHostgroupsName() {
    for _, slc := range o.templateHostgroups {
        for _, v := range slc {
            o.hgrpEnabledName = append(o.hgrpEnabledName, v)
        }
    }
    for _, slc := range o.hostgroups {
        for _, v := range slc {
            o.hgrpEnabledName = append(o.hgrpEnabledName, v)
        }
    }
}
