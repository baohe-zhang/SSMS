# SSMS

SSMS, Simple SWIM-like failure detector and Membership Service, for CS 425 MP2, implemented by Group#29 BaoheZhang&KechenLu.

## Project Info

- Language: Golang 
- Tested Platform: macOS 10.13.6, CentOS 7

## How-to

### Build and Deploy

Build this SSMS project is easy. Just run 

```shell
$ go build
```

To deploy on the each machine of the cluster, we have to git clone this repo like:

```shell
$ git clone git@gitlab.engr.illinois.edu:kechenl3/ssms.git
```

To simply the repo update and build in the VM cluster, we have a easy-to-use script, each time we push a new commit to remote repo then run

```shell
$ ./update_build_all.sh
```

### Run

To run our membership service, just execute the `./ssms` :

```shell
$ ./ssms
[INFO]: Start service

```

User can do front-end interaction in terminal when SSMS running and all the info/debug/error level logs would be stored in ssms.log file. We have four interaction command in the console. 

1. join,  join in the group
2. leave, voluntarily leave the group 
3. showlist, show membership list 
4. showid, show the id of this process itself

