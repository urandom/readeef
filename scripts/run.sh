#/bin/sh

set -x 
set -e

case "$1" in

  'server')
  	exec readeef server $@ $ARGS
	;;

  'client')
  	exec readeef-client $@ $ARGS
	;;

  'readeef-static-locator')
  	exec readeef-static-locator $@ $ARGS
  	;;

  *)
  	exec readeef server $@
	;;
esac