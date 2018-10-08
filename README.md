# SSMS

SSMS, Simple SWIM-like failure detector and Membership Service, for CS 425 MP2, implemented by Group#29 BaoheZhang&KechenLu, which includes infection-style dissemination, piggyback on membership messages, ping-ack failure detection with suspicion mechanism.

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

User can do front-end interaction in terminal when SSMS running and all the info/debug/error level logs would be stored in **ssms.log** file. We have four interaction command in the console. 

1. `join`,  join in the group
2. `leave`, voluntarily leave the group 
3. `showlist`, show membership list 
4. `showid`, show the id of this process itself

### Usage

For example, after we `join`  the group, we can show the list by `showlist` command, and show own id(including join timestamp and IP) by `showid` , and of course after `leave` command, we can show the list which is empty.

But one thing to be noted, we need to start our **introducer** first, otherwise other nodes cannot join in the group. 

```shell
[kechenl3@fa18-cs425-g29-10 ssms]$ ./ssms
[INFO]: Start service
join
showlist
------------------------------------------
Size: 10
idx: 0, TS: 1538885500836420235, IP: 172.22.156.95, ST: 1101
idx: 1, TS: 1538885509673591217, IP: 172.22.158.97, ST: 1
idx: 2, TS: 1538885539910014668, IP: 172.22.154.97, ST: 1
idx: 3, TS: 1538885547023251876, IP: 172.22.154.96, ST: 1
idx: 4, TS: 1538885556887217629, IP: 172.22.156.97, ST: 1
idx: 5, TS: 1538885782528985904, IP: 172.22.156.96, ST: 1
idx: 6, TS: 1538885794072175076, IP: 172.22.158.96, ST: 1
idx: 7, TS: 1538885805420036457, IP: 172.22.156.98, ST: 1
idx: 8, TS: 1538885809915011457, IP: 172.22.158.95, ST: 1
idx: 9, TS: 1538885815295645868, IP: 172.22.154.98, ST: 1
------------------------------------------
showid
Member (1538885805420036457, 172.22.156.98)
leave
showlist
------------------------------------------
Size: 0
------------------------------------------
```

### Log Debug

The distributed grep we implemented before in MP1 can be pretty helpful for our MP2 debug. We have a log file for membership service named `ssms.log` on each machine and first config the log file path in the configuration of our MP1 project dist-grep. Then start all of the grep servers.

Now we can do our distributed grep to get logs of different level(INFO/DEBUG/ERROR) or any pattern we want.

For example, we run dist-grep to query certain pattern "Failure" and then cut some other field in the terminal output.

```shell
Colearos-MacBook-Pro:client colearolu$ ./client -E "Failure" 
...
/home/kechenl3/go/src/ssms/ssms.log:179715:[INFO]: **172.22.156.95** 2018/10/06 22:53:57.505187 [Failure Detected](10.193.185.82, 1538884370872997000) Failed, detected by others
/home/kechenl3/go/src/ssms/ssms.log:181777:[INFO]: **172.22.156.95** 2018/10/06 22:55:49.535837 [Failure Detected](172.22.158.96, 1538884255618504178) Failed, detected by others
/home/kechenl3/go/src/ssms/ssms.log:183354:[INFO]: **172.22.156.95** 2018/10/06 22:57:12.381011 [Failure Detected](10.193.185.82, 1538884370872997000) Failed, detected by others
/home/kechenl3/go/src/ssms/ssms.log:183362:[INFO]: **172.22.156.95** 2018/10/06 22:57:12.748437 [Failure Detected](172.22.156.98, 1538884212443213196) Failed, detected by others
/home/kechenl3/go/src/ssms/ssms.log:183369:[INFO]: **172.22.156.95** 2018/10/06 22:57:12.963459 [Failure Detected](172.22.158.96, 1538884255618504178) Failed, detected by others
/home/kechenl3/go/src/ssms/ssms.log:183940:[INFO]: **172.22.156.95** 2018/10/06 22:57:39.253977 [Failure Detected](10.193.185.82, 1538884370872997000) Failed, detected by others
/home/kechenl3/go/src/ssms/ssms.log:183962:[INFO]: **172.22.156.95** 2018/10/06 22:57:40.209842 [Failure Detected](172.22.156.98, 1538884212443213196) Failed, detected by others
/home/kechenl3/go/src/ssms/ssms.log:184368:[INFO]: **172.22.156.95** 2018/10/06 22:57:57.235161 [Failure Detected](10.193.185.82, 1538884370872997000) Failed, detected by self
/home/kechenl3/go/src/ssms/ssms.log:184394:[INFO]: **172.22.156.95** 2018/10/06 22:57:57.961417 [Failure Detected](10.193.185.82, 1538884370872997000) Failed, detected by others
/home/kechenl3/go/src/ssms/ssms.log:184401:[INFO]: **172.22.156.95** 2018/10/06 22:57:58.160136 [Failure Detected](10.193.185.82, 1538884370872997000) Failed, detected by others
/home/kechenl3/go/src/ssms/ssms.log:184405:[INFO]: **172.22.156.95** 2018/10/06 22:57:58.237876 [Failure Detected](10.193.185.82, 1538884370872997000) Failed, detected by others
/home/kechenl3/go/src/ssms/ssms.log:184412:[INFO]: **172.22.156.95** 2018/10/06 22:57:58.281611 [Failure Detected](10.193.185.82, 1538884370872997000) Failed, detected by others
/home/kechenl3/go/src/ssms/ssms.log:184417:[INFO]: **172.22.156.95** 2018/10/06 22:57:58.352931 [Failure Detected](10.193.185.82, 1538884370872997000) Failed, detected by others
/home/kechenl3/go/src/ssms/ssms.log:184466:[INFO]: **172.22.156.95** 2018/10/06 22:58:00.628350 [Failure Detected](10.193.185.82, 0) Failed, detected by self
/home/kechenl3/go/src/ssms/ssms.log:184570:[INFO]: **172.22.156.95** 2018/10/06 22:58:05.628733 [Failure Detected](10.193.185.82, 0) Failed, detected by self
/home/kechenl3/go/src/ssms/ssms.log:184663:[INFO]: **172.22.156.95** 2018/10/06 22:58:10.629083 [Failure Detected](10.193.185.82, 0) Failed, detected by self
/home/kechenl3/go/src/ssms/ssms.log:184770:[INFO]: **172.22.156.95** 2018/10/06 22:58:15.629470 [Failure Detected](10.193.185.82, 0) Failed, detected by self
/home/kechenl3/go/src/ssms/ssms.log:190053:[INFO]: **172.22.156.95** 2018/10/06 23:17:15.579240 [Failure Detected](172.22.156.98, 1538885805420036457) Failed, detected by others
/home/kechenl3/go/src/ssms/ssms.log:198863:[INFO]: **172.22.156.95** 2018/10/06 23:25:27.151470 [Failure Detected](172.22.154.98, 1538885815295645868) Failed, detected by self
/home/kechenl3/go/src/ssms/ssms.log:1330
Total Connected VMs: 10
Total Counts: 13767
Total Time: 0.242 seconds


```

