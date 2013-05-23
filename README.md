scrapemonster
-------------

### Extremely Cursory Build Notes

*Requires the development version of go.* See http://tip.golang.org/ for instructions on how to install from source. Short version:

    $ cd $HOME
    $ hg clone -u tip https://code.google.com/p/go
    $ cd go/src
    $ ./all.bash
    $ export PATH=$HOME/go/bin:$PATH

(Note: the above won't work if you already have go installed in your homedir.)

*Use a custom GOPATH for this project.* Basic, example instructions for setting one up and checking out the source code into it:

    $ export GOPATH=$HOME/go_workspaces/scrapemonster
    $ mkdir -p $GOPATH/src/github.com/launchtime
    $ cd $GOPATH/src/github.com/launchtime
    $ hg git@github.com:launchtime/scrapemonster.git

(Alternatively, you could check out the source tree somewhere else and symlink it to `$GOPATH/src/github.com/launchtime/scrapemonster`.)

If everything is set up correctly, you should be able to run `make deps` && `make` in the `scrapemonster` directory. You can test that the programs are functioning properly with something like this:

    $ $GOPATH/bin/getDealInfo -s=tmon -d=14562681 -o=true
