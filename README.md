## What is GoPCEP ?

GoPCEP is a Stateful Segment Routing Traffic Engineering Controller it discovers network topology using BGP-LS then uses an SPF algorithm to find the shortest path and finally it pushes LSPs onto the network using PCEP protocol. You can also create LSPs manually in this case you need to specify the ERO yourself.

GoPCEP implements Stateful Segment Routing PCE using Path Computation Element Communication Protocol (PCEP)
with support for PCE-Initiated LSP Setup in a Stateful PCE Model. 

As of now, GoPCEP has REST and gRPC APIs although gRPC is lagging behind REST in the number of methods it provides. There is also a Web UI so one can quickly start using GoPCEP.  



![UI screenshot](webui.png?raw=true "UI screenshot") 


## Why GoPCEP ?
There already exists a well-known controller called [OpenDaylight (ODL)](https://www.opendaylight.org/). However, GoPCEP is intended to be a more lightweight version that has fewer features but is much easier to install and use.     

## How to run GoPCEP ?

The easiest way to start using GoPCEP is to download a binary from a release page and then set a few config parameters.
At the moment only IS-IS is supported as an IGP for topology discovery using BGP-LS.   

The config variables you need to set are:

* restapi user and pass
* grpcapi token  

The SSL certificates will be automatically generated on every run but you can generate your own and set the path to files in the config.

You can also run from the source you need to have Go 1.16  installed then just clone the repo and run `go build` that will produce the executable which you can run `./gopcep` . 

GoPCEP cannot run on Windows due to limitation in [GoBGP](https://github.com/osrg/gobgp/issues/1978). Docker is also not yet supported due to the fact that at least on Mac docker does NAT when the router connect to the controller as a result GoBGP cannot identify the clients. The docker issue can potentially be solved but I have not looked into it deep enough yet.

Once you have started GoPCEP you can navigate to the address you set int the config under ResAPI setting in your browser. For example if you left the address as 0.0.0.0 you can go to https://127.0.0.1:1443/ you will see in your browser Your connection is not private
you need to click advanced and then proceed. This happens because the certificates are self-signed there is not way to make this works cleanly when running on a laptop. 

##PLEASE HEED
You need to allow self signed certificates before the GUI can work for Chrome or Brave you can do chrome://flags/#allow-insecure-localhost. 

## Contribution.
Any contributions are welcome just submit a pull request. 

## Network setup.

I have only tested GoPCEP against Juniper VMX 17.2R1.13 as I do not have access to other vendors images. It would be good if anyone could test in a multi-vendor or at least using Cisco XR or XRv. 

I have added an Ansible role which can be used to setup a network of five routers. You can see my GNS setup below:

![GNS screenshot](gns.png?raw=true "GNS screenshot") 

## Roadmap.

One of the features I would like to add is auto-bandwidth where the collection of the traffic stats is done separately and the results are then fed into the GoPCEP using gRPC API. The next feature after AutoBW to add would be routing based on latency where the latency values for the links again are continuously streamed into the controller using gRPC. 

I have also started refactoring the original code I wrote in 2018 so there is less mutex locking. I have to admit that the code quality as it is way below my bar mainly due to lack of time and the development spread across almost three years. 
