FROM golang:1.5

# Add source of micro-analytics
ADD ./ $GOPATH/src/github.com/GitbookIO/micro-analytics

# Build server
RUN cd $GOPATH/src/github.com/GitbookIO/micro-analytics && go get && go build --ldflags='-s'

##
# Env
##
ENV MA_PORT 7070
ENV MA_ROOT /opt/data/

# Set default command
CMD /go/bin/micro-analytics
