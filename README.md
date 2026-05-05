## EskPad

> http://localhost:8080

### Compile

```bash
GOOS=linux GOARCH=amd64 go build -o bin/linux/eskpad eskpad.go

# for alpine - requires: sudo apt-get install musl-tools
CC=musl-gcc CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o bin/alpine/eskpad eskpad.go

```

### Install

Copy **eskpad** binary and the **templates** folder to a location, then run it *./eskpad* or create a service to run it:

#### Service Alpine

```bash
doas nano /etc/init.d/eskpad
doas chmod +x /etc/init.d/eskpad
doas rc-update add eskpad default
doas rc-service eskpad start
```

```bash
#!/sbin/openrc-run

name="EskPad"
description="Simple pad"

command="/home/eskiso/eskpad"
directory="/home/eskiso"

command_background="yes"
pidfile="/run/${RC_SVCNAME}.pid"
command_user="eskiso:eskiso"

output_log="/home/eskiso/eskpad.log"
error_log="/home/eskiso/eskpad.log"

supervisor="supervise-daemon"

depend() {
    need net
}

start_pre() {
    checkpath --directory --owner root:root --mode 0755 /run
}

```

#### Service Ubuntu

```bash 
sudo nano /etc/systemd/system/eskpad.service

## File contents

[Unit]
Description=EskPad service
After=network.target
StartLimitIntervalSec=0

[Service]
Type=simple
Restart=always
RestartSec=1
User=eskiso
WorkingDirectory=/home/eskiso
ExecStart=/home/eskiso/eskpad

[Install]
WantedBy=multi-user.target

## End of file contents

sudo systemctl enable eskpad
sudo systemctl start eskpad
```