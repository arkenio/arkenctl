


arkenctl
========

[![Build Status](https://travis-ci.org/arkenio/arkenctl.png?branch=master)](https://travis-ci.org/arkenio/arkenctl)

Arken is a simple tool to instrospect an Arken cluster and do some simple actions like start/stop a service/

## Usage

### Cluster watch

arkenctl can watch if the cluster is healthy. If something goes wrong, then it generates an error log. 

	# arkenctl watch
	
### Services introspection

	# arkenctl service list -status passivated
	Name			Index	Domain					Status		LastAccess
    ----			-----	------					------		----------
	nxio_000001		1	testenv-nuxeo.test.io.nuxeo.com		passivated	2014-12-09 07:51:01 +0000 UTC
	nxio_000002		1	test2-nuxeo.test.io.nuxeo.com		passivated	2014-12-08 18:10:01 +0000 UTC
	nxio_000003		1	test3-nuxeo.test.io.nuxeo.com		passivated	2014-12-08 18:09:51 +0000 UTC
	
	## Shows informations about nxio_00101 
	# arkenctl service cat nxio_00001
	===========================================
    Node index : 1
    Name : nxio_000001
    UnitName : nxio@000001.service
    Etcd key : /services/nxio_000001/1
    Domain name : testenv-nuxeo.test.io.nuxeo.com
    Location : 172.32.46.78:49160
    LastAccess: 2014-12-09 07:51:01 +0000 UTC
    Status: passivated
      * expected : passivated
      * current : stopped
      * alive :

### Command templating

The `service list` command may take a `--template` parameter that allows to specify the template used 
to render a service's line. It is based Go template and you can issue commande like this :

    arkenctl service list --template "Domain : {{.Domain}} Service: {{.Name}}"

The context used for the templating is the Service context, meaning that you have access to all the 
public property of the Service object :


    Service
        Index
        NodeKey
        UnitName
        Location
            Host
            Port
        Domain
        Name
        Status
            Expected
            Current
            Alive
        LastAccess


## Report & Contribute


We are glad to welcome new developers on this initiative, and even simple usage feedback is great.
- Ask your questions on [Nuxeo Answers](http://answers.nuxeo.com)
- Report issues on this github repository (see [issues link](http://github.com/arkenio/arkenctl/issues) on the right)
- Contribute: Send pull requests!


## About Nuxeo

Nuxeo provides a modular, extensible Java-based
[open source software platform for enterprise content management](http://www.nuxeo.com/en/products/ep),
and packaged applications for [document management](http://www.nuxeo.com/en/products/document-management),
[digital asset management](http://www.nuxeo.com/en/products/dam) and
[case management](http://www.nuxeo.com/en/products/case-management).

Designed by developers for developers, the Nuxeo platform offers a modern
architecture, a powerful plug-in model and extensive packaging
capabilities for building content applications.

More information on: <http://www.nuxeo.com/>
