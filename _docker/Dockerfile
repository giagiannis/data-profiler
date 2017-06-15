FROM debian

MAINTAINER Giannis Giannakopoulos

ENV GOPATH="/opt" GIMME_GO_VERSION="1.7"

RUN apt-get update && \
	apt-get install -y \
		curl \ 
		gcc \
		git \
		sqlite3 && \
	curl -sL -o gimme https://raw.githubusercontent.com/travis-ci/gimme/master/gimme && \
	chmod +x ./gimme && \
	eval $(./gimme) && \
	go get -v github.com/giagiannis/data-profiler/data-profiler-server

ENV R_LIBS="/opt/rlibs"
RUN apt-get update && apt-get install -y r-base && \
echo "install.packages('GenSA', repos='http://cran.us.r-project.org')" | R --vanilla && \
echo "install.packages('e1071', repos='http://cran.us.r-project.org')" | R --vanilla

ADD entrypoint.sh /
ADD data-profiler /etc/data-profiler
EXPOSE 8080

ENTRYPOINT "/entrypoint.sh"
