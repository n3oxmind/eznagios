# eznagios

gonag is a generic nagios config manager which is unfrotunatley half completed but fully functional. 
I was able to reversed engineer the way nagios function and make it possible to write application that will automate the config and make it more user friendly.
This was suppose to have a GUI and hope someone will take this and continue this project.

### Current Features 
- Search services and hostgroups associated with specific hosts(s) (support bulk search, and regex)
- Search all hosts that are using the same service checks (support bulk search)
- Delete/Purge host(s) and its associated services and hostgroups (support bulk deletion)

### Install
1. download the repo
2. cd to gonag directory 
3. make && sudo make install


### Usage Examples
#### Search
```shell
$ eznagios search -h java26-game01.sea.bigfishgames.com -pretty
$ eznagios search -h mts-.* 
```

#### Delete 
```shell
$ eznagios delete -h mts-.* --verbose
```


Note:

This is still under development, but search and delete are function just fine and even better than eznagios :). 
You can upgrade this or added more functionality to it.
