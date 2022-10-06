#  Building from source / Ubuntu Installation Guide 

Although personally I feel that using the docker container is the best way of using and enjoying something like Podgrab, a lot of people in the community are still not comfortable with using Docker and wanted to host it natively on their Linux servers.

This guide has been written with Ubuntu in mind. If you are using any other flavour of Linux and are decently competent with using command line tools, it should be easy to figure out the steps for your specific distro. 

## Install Go

Podgrab is built using Go which would be needed to compile and build the source code. Podgrab is written with Go 1.15 so any version equal to or above this should be good to Go. 

If you already have Go installed on your machine, you can skip to the next step.

Get precise Go installation process at the official link here - https://golang.org/doc/install

Following steps will only work if Go is installed and configured properly.

## Install dependencies

``` bash
 sudo apt-get install -y git ca-certificates ufw gcc
```

## Clone from Git

``` bash
git clone --depth 1 https://github.com/akhilrex/podgrab
```

## Build and Copy dependencies

``` bash
cd podgrab
mkdir -p ./dist
cp .env ./dist
go build -o ./dist/podgrab ./main.go
```

## Create final destination and copy executable
``` bash
sudo mkdir -p /usr/local/bin/podgrab
mv -v dist/* /usr/local/bin/podgrab
mv -v dist/.* /usr/local/bin/podgrab
```

At this point theoretically the installation is complete. You can make the relevant changes in the ```.env``` file present at ```/usr/local/bin/podgrab``` path and run the following command 

``` bash
cd /usr/local/bin/podgrab && ./podgrab
```

Point your browser to http://localhost:8080 (if trying on the same machine) or http://server-ip:8080 from other machines.

If you are using ufw or some other firewall, you might have to make an exception for this port on that.

## Setup as service (Optional)

If you want to run Podgrab in the background as a service or auto-start whenever the server starts, follow the next steps.

Create new file named ```podgrab.service``` at ```/etc/systemd/system``` and add the following content. You will have to modify the content accordingly if you changed the installation path in the previous steps.


``` unit
[Unit]
Description=Podgrab

[Service]
ExecStart=/usr/local/bin/podgrab/podgrab
WorkingDirectory=/usr/local/bin/podgrab/
[Install]
WantedBy=multi-user.target
```

Run the following commands 
``` bash
sudo systemctl daemon-reload
sudo systemctl enable podgrab.service
sudo systemctl start podgrab.service
```

Run the following command to check the service status.

``` bash
sudo systemctl status podgrab.service
```

# Update Podgrab

In case you have installed Podgrab and want to update the latest version (another area where Docker really shines) you need to repeat the steps from cloning to building and copying.

Stop the running service (if using)
``` bash
sudo systemctl stop podgrab.service
```

## Clone from Git

``` bash
git clone --depth 1 https://github.com/akhilrex/podgrab
```

## Build and Copy dependencies

``` bash
cd podgrab
mkdir -p ./dist
cp -r client ./dist
cp -r webassets ./dist
cp .env ./dist
go build -o ./dist/podgrab ./main.go
```

## Create final destination and copy executable
``` bash
sudo mkdir -p /usr/local/bin/podgrab
mv -v dist/* /usr/local/bin/podgrab
```

Restart the service (if using)
``` bash
sudo systemctl start podgrab.service
```
