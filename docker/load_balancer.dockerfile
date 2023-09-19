FROM redis AS LoadBalancer

RUN apt-get update -y

#========#
# PROTOC #
#========#
RUN apt-get install wget -y
RUN wget https://github.com/protocolbuffers/protobuf/releases/download/v24.0/protoc-24.0-linux-x86_64.zip -O /ProtocolBuffer.zip
RUN apt-get install unzip -y
RUN unzip /ProtocolBuffer.zip -d /ProtocolBuffer
ENV PATH="$PATH:/ProtocolBuffer/bin"
RUN rm /ProtocolBuffer.zip
#======#
# GRPC #
#======#
RUN apt-get install golang -y
ENV GOPATH="$HOME/go"
ENV PATH="$PATH:/usr/local/go/bin:$GOPATH/bin"
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
