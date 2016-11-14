FROM golang:1.7.3
MAINTAINER Koichi Shiraishi <zchee.io@gmail.com>

ENV COVERAGE_SERVICE=""

RUN set -ux \
	&& wget -q -O - https://github.com/neovim/neovim/releases/download/nightly/neovim-linux64.tar.gz | tar xzf - --strip-components=1 -C "/usr/local" \
	&& nvim --version \
	\
	&& go get github.com/constabulary/gb/...

COPY . /nvim-go
WORKDIR /nvim-go

CMD ["sh", "./scripts/coverage.sh"]
