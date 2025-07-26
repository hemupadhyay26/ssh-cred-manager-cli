FROM alpine:3.22
COPY ssh-cred-manager-cli /usr/bin/ssh-cred-manager-cli
ENTRYPOINT ["/usr/bin/ssh-cred-manager-cli"]