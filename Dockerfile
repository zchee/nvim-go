FROM golang:1.7.0
MAINTAINER Koichi Shiraishi <k@zchee.io>

ENV VIM=/usr/local/share/nvim

RUN set -ux \
	&& wget -q -O - https://github.com/neovim/neovim/releases/download/nightly/neovim-linux64.tar.gz | tar xzf - --strip-components=1 -C "/usr/local" \
	&& nvim --version \
	\
	&& go get -u -v -x github.com/constabulary/gb/...

COPY . /nvim-go
WORKDIR /nvim-go
ENV CI=true

CMD ["gb", "test", "-v", "-race", "-bench=.", "-benchmem"]
