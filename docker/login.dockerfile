FROM golang

RUN apt-get update

#========#
# PROTOC #
#========#
RUN wget https://github.com/protocolbuffers/protobuf/releases/download/v24.0/protoc-24.0-linux-x86_64.zip -O /ProtocolBuffer.zip
RUN apt-get install unzip
RUN unzip /ProtocolBuffer.zip -d /ProtocolBuffer
ENV PATH="$PATH:/ProtocolBuffer/bin"
RUN rm /ProtocolBuffer.zip
#======#
# GRPC #
#======#
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest