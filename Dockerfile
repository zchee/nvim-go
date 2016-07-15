FROM golang:1.7rc1
MAINTAINER Koichi Shiraishi <k@zchee.io>

RUN set -ux \
	&& wget -q -O - https://github.com/neovim/neovim/releases/download/nightly/neovim-linux64.tar.gz | tar xzf - --strip-components=1 -C "/usr/local" \
	&& nvim --version \
	\
	&& go get -u -v -x github.com/constabulary/gb/...

COPY . /nvim-go
WORKDIR /nvim-go

CMD ["gb", "test", "-v", "-race", "-bench=.", "-benchmem"]
