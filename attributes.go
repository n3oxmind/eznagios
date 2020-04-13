/*Nagios standard objects directives and BFG custom directives.
Any new custom directives/variables need to be added here otherwise eznagios will flag it as unknown*/
// Standard attributes for Nagios host object
package main
var (
    maxHostAttrLen      = maxObjAttrLength(&hostAttr)
    maxSvcAttrLen       = maxObjAttrLength(&serviceAttr)
    maxSvcGrpAttrLen    = maxObjAttrLength(&serviceGroupAttr)
    maxHgrpAttrLen      = maxObjAttrLength(&hostGroupAttr)
    maxContactAttrLen   = maxObjAttrLength(&contactAttr)
    maxCGrpAttrLen      = maxObjAttrLength(&contactGroupAttr)
    maxCmdAttrLen       = maxObjAttrLength(&commandAttr)
    maxSvcEsclAttrLen   = maxObjAttrLength(&serviceEscalationAttr)
    maxSvcDpndlAttrLen  = maxObjAttrLength(&serviceDependencyAttr)
    maxHostEsclAttrLen  = maxObjAttrLength(&hostEscalationAttr)
    maxHostDpndAttrLen  = maxObjAttrLength(&hostDependencyAttr)
    maxCustomAtrrLen    = maxObjAttrLength(&customAttr)
)
var (
    hostAttr = []string{
         "host_name", 
         "name",
         "use",
         "alias",
         "display_name",
         "address",
         "parents",
         "hostgroups",
         "check_command",
         "initial_state",
         "max_check_attempts",
         "check_interval",
         "retry_interval",
         "active_checks_enabled",
         "passive_checks_enabled",
         "check_period",
         "obsess_over_host",
         "check_freshness",
         "freshness_threshold",
         "event_handler",
         "event_handler_enabled",
         "low_flap_threshold",
         "high_flap_threshold",
         "flap_detection_enabled",
         "flap_detection_options",
         "process_perf_data",
         "retain_status_information",
         "retain_nonstatus_information",
         "contacts",
         "contact_groups",
         "notification_interval",
         "first_notification_delay",
         "notification_period",
         "notification_options",
         "notifications_enabled",
         "stalking_options",
         "notes",
         "notes_url",
         "action_url",
         "icon_image",
         "icon_image_alt",
         "vrml_image",
         "statusmap_image",
         "2d_coords",
         "3d_coords",
         "register"}
    serviceAttr = []string{
         "host_name", 
         "name",
         "use",
         "hostgroup_name",
         "service_description",
         "display_name",
         "servicegroups",
         "is_volatile",
         "check_command",
         "initial_state",
         "max_check_attempts",
         "check_interval",
         "retry_interval",
         "retry_check_interval",
         "normal_check_interval",
         "parallelize_check",
         "active_checks_enabled",
         "passive_checks_enabled",
         "check_period",
         "obsess_over_service",
         "check_freshness",
         "freshness_threshold",
         "event_handler",
         "event_handler_enabled",
         "low_flap_threshold",
         "high_flap_threshold",
         "flap_detection_enabled",
         "flap_detection_options",
         "process_perf_data",
         "retain_status_information",
         "retain_nonstatus_information",
         "notification_interval",
         "first_notification_delay",
         "notification_period",
         "notification_options",
         "notifications_enabled",
         "contacts",
         "contact_groups",
         "stalking_options",
         "notes",
         "notes_url",
         "action_url",
         "icon_image",
         "register",
         "icon_image_alt"}
    hostGroupAttr = []string{
         "hostgroup_name",
         "alias",
         "members",
         "hostgroup_members"  ,
         "notes",
         "notes_url",
         "action_url",
         "register"}
    contactAttr = []string{
        "name",
        "contact_name",
        "alias",
        "use",
        "host_notifications_enabled",
        "service_notifications_enabled",
        "service_notification_period",
        "host_notification_period",
        "service_notification_options",
        "host_notification_options",
        "service_notification_commands",
        "host_notification_commands",
        "email",
        "pager",
        "address1",
        "address2",
        "register",
        "can_submit_commands"}
    contactGroupAttr = []string{
        "name",
        "use",
        "alias",
        "contact_name",
        "contactgroup_name",
        "members",
        "contactgroup_members",
        "register" }
    timeperiodAttr = []string{
        "name",
        "timeperiod_name",
        "alias",
        "exclude" }
    commandAttr = []string{
        "command_name",
        "command_line"}
    serviceDependencyAttr = []string{
        "host_name",
        "hostgroup_name",
        "servicegroup_name",
        "service_description",
        "dependent_host_name",
        "dependent_hostgroup_name",
        "dependent_servicegroup_name",
        "dependent_service_description",
        "inherits_parent",
        "execution_failure_criteria",
        "notification_failure_criteria",
        "dependency_period"}
    serviceGroupAttr = []string{
        "servicegroup_name",
        "servicegroup_members",
        "alias",
        "members",
        "notes",
        "notes_url",
        "action_url"}
    serviceEscalationAttr = []string{
        "host_name",
        "hostgroup_name",
        "service_description",
        "contacts",
        "contactgroup_name",
        "first_notification",
        "last_notification",
        "notification_interval",
        "escalation_period",
        "escalation_options"}
    hostDependencyAttr = []string{
        "host_name",
        "hostgroup_name",
        "dependent_host_name",
        "dependent_hostgroup_name",
        "inherits_parent",
        "execution_failure_criteria",
        "notification_failure_criteria",
        "dependency_period"}
    hostEscalationAttr = []string{
        "host_name",
        "hostgroup_name",
        "contacts",
        "contact_groups",
        "first_notification",
        "last_notification",
        "notification_interval",
        "escalation_period",
        "escalation_options"}
    customAttr = []string{
        "_EVENT_HANDLER",
        "_EVENTHANDLER",
        "_IPINSIDE",
        "_IPOUTSIDE",
        "_IPPUBLIC",
        "_graphite",
        "_graphite_id",
        "_oob_address",
        "_healthpage",
        "_tags",
        "_comment",
        "_cacti",
        "_tags"}
)

func maxObjAttrLength(a *[]string) int{
    maxLength := len((*a)[0])
    for _,v := range *a {
        if len(v) > maxLength {
            maxLength = len(v)
        }
    }
    return maxLength
}

