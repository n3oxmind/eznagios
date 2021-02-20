# PROJECT IS UNDER DEVELOPMENT 

eznagios is a Go tool to manage and automate nagios config files. Unfrotunatley, the tool is half completed but fully functional. 
I was able to reversed engineer the way nagios function and make it possible to write application that will automate the config and make it more user friendly.

### Current Features 
- Search services and hostgroups associated with specific hosts(s) (support bulk search, and regex)
- Search all hosts that are using the same service checks (support bulk search)
- Delete/Purge host(s) and its associated services and hostgroups (support bulk deletion)

### Features still in Development
- Add host(s) based on an existing host.
- Add host(s) based on a template.
- Add service(s) check for an existing host/servicegroup/hostgroup
- Integrate config with Git

### Install
1. download the repo
2. cd to eznagios directory 
3. make && sudo make install


### Usage Examples
#### Search
```shell
$ eznagios search -h host_name -pretty
$ eznagios search -h part_of_hostname-.* 
```

#### Delete 
```shell
$ eznagios delete -h part_of_hostname-.* --verbose
```

Note:

- This tool is still under development. search and delete are function just fine. I will add the rest of the features mentioned above on my free time.
