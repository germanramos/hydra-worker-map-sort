hydra-worker-map-sort
=====================

Worker for Hydra v3.  
Given an array of values it map and sort by an specific attribute.

# Installation

## Ubuntu/Debian

Add PPAs for:  
https://launchpad.net/~chris-lea/+archive/libpgm  
https://launchpad.net/~chris-lea/+archive/zeromq  
  
and run:  
```
sudo dpkg -i hydra-worker-map-sort-1-0.x86_64.deb
sudo apt-get install -f
```
## CentOS/RedHat/Fedora
```
sudo yum install libzmq3-3.2.2-13.1.x86_64.rpm hydra-worker-map-sort-1-0.x86_64.rpm
```

# Configuration

In apps.json:

- Name: "MapSort"
- Arguments:
  - mapAttr: The name of the attribute that we are going to map and sort with
  - mapSort: An array with the prefered order. If the order do not match the instance will be put in map at the end of the list

## Configuration example

"MapSort": {	
				"mapAttr": "cloud",
				"mapSort": ["google", "amazon", "azure"]
			}
			
This will map the instance by the attribute "cloud" placing first all the instances with cloud=google, after that all the instances with cloud=amazon, after that all the instances with cloud=azure, and at the end all the other instances. 

## Service configuration

No additional configuration is needed if running in the same machine that Hydra.  
Tune start file at /etc/init.d/hydra-worker-map-sort if you run in a separate machine

# Run
```
sudo /etc/init.d/hydra-worker-map-sort start
```

# License

(The MIT License)

Authors:  
Germán Ramos &lt;german.ramos@gmail.com&gt;  
Pascual de Juan &lt;pascual.dejuan@gmail.com&gt;  
Jonas da Cruz &lt;unlogic@gmail.com&gt;  
Luis Mesas &lt;luismesas@gmail.com&gt;  
Alejandro Penedo &lt;icedfiend@gmail.com&gt;  
Jose María San José &lt;josem.sanjose@gmail.com&gt;  

Permission is hereby granted, free of charge, to any person obtaining
a copy of this software and associated documentation files (the
'Software'), to deal in the Software without restriction, including
without limitation the rights to use, copy, modify, merge, publish,
distribute, sublicense, and/or sell copies of the Software, and to
permit persons to whom the Software is furnished to do so, subject to
the following conditions:

The above copyright notice and this permission notice shall be
included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED 'AS IS', WITHOUT WARRANTY OF ANY KIND,
EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY
CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,
TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
