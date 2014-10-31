package main

import (
	"github.com/flynn/flynn/Godeps/_workspace/src/github.com/flynn/go-docopt"
)

func main() {
	usage := `flynn-release generates Flynn releases.

Usage:
  flynn-release status <commit>
  flynn-release manifest [--output=<dest>] [--docker-registry=<domain>] [--docker-user=<user>] [--id-file=<file>] <template>
  flynn-release download [--driver=<name>] [--root=<path>] <manifest>
  flynn-release upload <manifest> [<tag>]
  flynn-release vagrant <url> <checksum> <version> <provider>
  flynn-release amis <version> <ids>

Options:
  -o --output=<dest>             output destination file ("-" for stdout) [default: -]
  -i --id-file=<file>            JSON file containing ID mappings
  -d --driver=<name>             image storage driver [default: aufs]
  -r --root=<path>               image storage root [default: /var/lib/docker]
  -s --docker-registry=<domain>  the Docker registry to use [default: registry.hub.docker.com]
  -u --docker-user=<user>        the Docker username to use [default: flynn]
`
	args, _ := docopt.Parse(usage, nil, true, "", false)

	switch {
	case args.Bool["status"]:
		status(args)
	case args.Bool["manifest"]:
		manifest(args)
	case args.Bool["download"]:
		download(args)
	case args.Bool["upload"]:
		upload(args)
	case args.Bool["vagrant"]:
		vagrant(args)
	case args.Bool["amis"]:
		amis(args)
	}
}
