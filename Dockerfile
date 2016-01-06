FROM golang:1.5

# Add source of micro-analytics
ADD ./ $GOPATH/src/github.com/GitbookIO/micro-analytics

# Build server
RUN cd $GOPATH/src/github.com/GitbookIO/micro-analytics && go get && go build --ldflags='-s'

##
# Env
##
ENV PORT 7070
ENV MA_ROOT /opt/data/
ENV MA_CACHE_DIR /opt/cache/

# Set default command
CMD ["/go/bin/micro-analytics"]
