# What is GoPCEP ?

GoPCEP is a Stateful Segment Routing Traffic Engineering Controller it discovers network topology using BGP-LS then uses an SPF algorithm to find the shortest path and finally it pushes LSPs onto the network using PCEP protocol. You can also create LSPs manually in this case you need to specify the ERO yourself.

GoPCEP implements Stateful Segment Routing PCE using Path Computation Element Communication Protocol (PCEP)
with support for PCE-Initiated LSP Setup in a Stateful PCE Model. 

As of now GoPCEP has REST and gRPC APIs although gRPC is lagging behind REST in completeness. There is also a Web UI so one can quickly start using GoPCEP. 

![UI screenshot](webui.png?raw=true "UI screenshot") 


## Why GoPCEP ?
There already exists a well-know controller called [OpenDaylight (ODL)](https://www.opendaylight.org/). However, GoPCEP is intended to be a more light-weight version which has less features but much easier to install and use.     







