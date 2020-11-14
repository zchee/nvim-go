FROM golang:1.15-buster AS nvim-go
LABEL maintainer "Koichi Shiraishi <zchee.io@gmail.com>"

RUN set -ux \
	&& wget -q -O - https://github.com/neovim/neovim/releases/download/nightly/nvim-linux64.tar.gz | tar xzf - --strip-components=1 -C "/usr/local" \
	&& nvim --version \
	\
	&& go get github.com/constabulary/gb/...

COPY . /go/src/github.com/zchee/nvim-go
WORKDIR /go/src/github.com/zchee/nvim-go

CMD ["make", "test"]
