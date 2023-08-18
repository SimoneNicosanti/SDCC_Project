FROM python

RUN apt-get update

#===========#
# RABBIT_MQ #
#===========#
RUN pip3 install pika
#========#
# PROTOC #
#========#
RUN wget https://github.com/protocolbuffers/protobuf/releases/download/v24.0/protoc-24.0-linux-x86_64.zip -O /ProtocolBuffer.zip
RUN unzip /ProtocolBuffer.zip -d /ProtocolBuffer
ENV PATH="$PATH:/ProtocolBuffer/bin"
RUN rm /ProtocolBuffer.zip
#======#
# GRPC #
#======#
RUN pip3 install grpcio
RUN pip3 install grpcio-tools

# To compile protocol buffer
# python -m grpc_tools.protoc -I../proto --python_out=./proto --pyi_out=./proto --grpc_python_out=./proto ../proto/Login.proto