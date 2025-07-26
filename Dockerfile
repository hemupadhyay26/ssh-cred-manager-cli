FROM alpine:3.20
COPY ssh-cred-manager-cli /usr/bin/ssh-cred-manager-cli
ENTRYPOINT ["/usr/bin/ssh-cred-manager-cli"]