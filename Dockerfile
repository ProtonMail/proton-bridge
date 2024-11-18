FROM ubuntu:24.04

WORKDIR /app

RUN apt-get update && apt-get install -y build-essential wget

RUN wget https://go.dev/dl/go1.21.9.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go1.21.9.linux-amd64.tar.gz && \
    rm go1.21.9.linux-amd64.tar.gz
ENV PATH=$PATH:/usr/local/go/bin

COPY . .

RUN apt-get install -y libglvnd-dev libsecret-1-dev pkg-config
ENV MSYSTEM=

RUN apt-get update && apt-get install -y \
    pass \
    gnupg2 \
    && rm -rf /var/lib/apt/lists/*

# Set up password store environment
ENV PASSWORD_STORE_DIR=/root/.password-store
ENV GNUPGHOME=/root/.gnupg

# Create required directories
RUN mkdir -p ${PASSWORD_STORE_DIR} ${GNUPGHOME} && \
    chmod 700 ${GNUPGHOME}

# Generate GPG key in batch mode
RUN gpg --batch --generate-key <<EOF
%no-protection
Key-Type: RSA
Key-Length: 2048
Subkey-Type: RSA
Subkey-Length: 2048
Name-Real: Proton Bridge
Name-Email: bridge@localhost
Expire-Date: 0
%commit
EOF

# Initialize pass with the GPG key
RUN key=$(gpg --list-keys --with-colons bridge@localhost | awk -F: '/^pub/ {print $5}') && \
    pass init $key

# Set up a basic password store entry for bridge
RUN echo "bridge-test-password" | pass insert -e bridge-test

RUN make build-nogui

EXPOSE 1025
EXPOSE 1143

CMD ["./bridge", "--noninteractive"]