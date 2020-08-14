# Go in Semaphore

```console
ll /usr/local/bin/
...
 lrwxrwxrwx  1 root   root         34 Aug 12 07:35 go -> /usr/local/golang/1.10.8/go/bin/go*
...
```

```console
ll /usr/local/golang/

 drwxr-xr-x  8 root root 4096 Aug 13 21:19 ./
 drwxr-xr-x 14 root root 4096 Aug 12 07:34 ../
 drwxr-xr-x  5 root root 4096 Aug 12 07:34 1.10.8/
 drwxr-xr-x  3 root root 4096 Aug 12 07:34 1.11.13/
 drwxr-xr-x  3 root root 4096 Aug 12 07:35 1.12.17/
 drwxr-xr-x  3 root root 4096 Aug 12 07:35 1.13.14/
 drwxr-xr-x  3 root root 4096 Aug 12 07:35 1.14.6/
 drwxr-xr-x  3 root root 4096 Aug 13 21:19 1.15/
```

```bash
# sem-version command
# https://github.com/semaphoreci/classic-toolbox
ll /home/runner/.toolbox/

# change-go-version function (change-go-version.sh)
# no GitHub repository
ll /opt/
```

`~/.bash_profile`:

```bash
# ...
source /opt/change-go-version.sh
# ...
```

`/opt/change-go-version.sh`:

```bash
#!/bin/bash

function switch_go() {
  local root_path=$1

  # remove other Go version from path
  export PATH=`echo $PATH | sed -e 's|:/usr/local/golang/[1-9.]*/go/bin||'`

  sudo ln -fs $root_path/bin/go /usr/local/bin/go

  # setup GOROOT
  export GOROOT="$root_path"

  # add new go installation to PATH
  export PATH="$PATH:$root_path/bin"
}

function change-go-version() {
  typeset NEW_GO_VERSION
  NEW_GO_VERSION="$1"

  case "$NEW_GO_VERSION" in
    "1.10" )
      switch_go /usr/local/golang/1.10.8/go
      ;;
    "1.11" )
      switch_go /usr/local/golang/1.11.13/go
      ;;
    "1.12" )
      switch_go /usr/local/golang/1.12.17/go
      ;;
    "1.13" )
      switch_go /usr/local/golang/1.13.14/go
      ;;
    "1.14" )
      switch_go /usr/local/golang/1.14.6/go
      ;;
    * )
      echo "Version not found. Please try another one."
      return 1
      ;;
  esac

  echo "Currently active Go version is:"
  go version

  return $?
}
```
