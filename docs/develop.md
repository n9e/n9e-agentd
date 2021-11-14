## Primer agent Development

#### build with python support

At present, it only supports linux/amd64

```sh
# download source
git clone https://github.com/n9e/n9e-agentd.git
cd n9e-agentd

# download file https://github.com/n9e/n9e-agentd/releases/download/v5.1.0-alpha.0/cache.tar.gz
tar xzvf cache.tar.gz

# generate n9e-agentd-5.1.0-alpha.0.linux.amd64.tar.gz
make pkg
```

#### build mocker

Use mocker instead of n9e-server

```sh
# build
go build -o ./build/mocker ./cmd/mocker

# run
./build/mocker
```
