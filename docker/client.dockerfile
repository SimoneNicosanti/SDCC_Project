FROM python

RUN apt-get update -y

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
# To solve module import error
ENV PYTHONPATH="$PYTHONPATH:/src/proto:/src/proto/file_transer"
ENV AWS_SHARED_CREDENTIALS_FILE="/.aws"
#======#
# VIEW #
#======#
RUN pip3 install customtkinter

RUN pip3 install jproperties
RUN pip3 install boto3