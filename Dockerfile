FROM golang:1.7rc1
MAINTAINER Koichi Shiraishi <k@zchee.io>

RUN set -ux \
	&& curl -sSL "https://github.com/neovim/neovim/releases/download/nightly/neovim-linux64.tar.gz" -o /tmp/neovim-linux64.tar.gz \
	&& mkdir -p /tmp/neovim /usr/local/bin /usr/local/share \
	&& tar -xzf /tmp/neovim-linux64.tar.gz -C /tmp \
	&& mv /tmp/neovim-linux64/bin/nvim /usr/local/bin \
	&& mv /tmp/neovim-linux64/share/nvim /usr/local/share \
	&& rm -rf /tmp/neovim-linux64.tar.gz /tmp/neovim-linux64 \
	\
	&& go get -u -v -x github.com/constabulary/gb/...

COPY . /nvim-go
WORKDIR /nvim-go

CMD ["gb", "test", "-v", "-race", "-bench=.", "-benchmem"]
