# Generated by the gRPC Python protocol compiler plugin. DO NOT EDIT!
"""Client and server classes corresponding to protobuf-defined services."""
import grpc

from file_transfer import ClientFileTransfer_pb2 as file__transfer_dot_ClientFileTransfer__pb2


class FileServiceStub(object):
    """Missing associated documentation comment in .proto file."""

    def __init__(self, channel):
        """Constructor.

        Args:
            channel: A grpc.Channel.
        """
        self.Download = channel.unary_stream(
                '/client.FileService/Download',
                request_serializer=file__transfer_dot_ClientFileTransfer__pb2.FileDownloadRequest.SerializeToString,
                response_deserializer=file__transfer_dot_ClientFileTransfer__pb2.FileChunk.FromString,
                )
        self.Upload = channel.stream_unary(
                '/client.FileService/Upload',
                request_serializer=file__transfer_dot_ClientFileTransfer__pb2.FileChunk.SerializeToString,
                response_deserializer=file__transfer_dot_ClientFileTransfer__pb2.Response.FromString,
                )


class FileServiceServicer(object):
    """Missing associated documentation comment in .proto file."""

    def Download(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def Upload(self, request_iterator, context):
        """rpc DownloadFromEdge(EdgeFileDownloadRequest) returns (stream EdgeFileChunk);
        rpc Delete(FileDeleteRequest) returns (FileDeleteResponse)
        """
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')


def add_FileServiceServicer_to_server(servicer, server):
    rpc_method_handlers = {
            'Download': grpc.unary_stream_rpc_method_handler(
                    servicer.Download,
                    request_deserializer=file__transfer_dot_ClientFileTransfer__pb2.FileDownloadRequest.FromString,
                    response_serializer=file__transfer_dot_ClientFileTransfer__pb2.FileChunk.SerializeToString,
            ),
            'Upload': grpc.stream_unary_rpc_method_handler(
                    servicer.Upload,
                    request_deserializer=file__transfer_dot_ClientFileTransfer__pb2.FileChunk.FromString,
                    response_serializer=file__transfer_dot_ClientFileTransfer__pb2.Response.SerializeToString,
            ),
    }
    generic_handler = grpc.method_handlers_generic_handler(
            'client.FileService', rpc_method_handlers)
    server.add_generic_rpc_handlers((generic_handler,))


 # This class is part of an EXPERIMENTAL API.
class FileService(object):
    """Missing associated documentation comment in .proto file."""

    @staticmethod
    def Download(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_stream(request, target, '/client.FileService/Download',
            file__transfer_dot_ClientFileTransfer__pb2.FileDownloadRequest.SerializeToString,
            file__transfer_dot_ClientFileTransfer__pb2.FileChunk.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def Upload(request_iterator,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.stream_unary(request_iterator, target, '/client.FileService/Upload',
            file__transfer_dot_ClientFileTransfer__pb2.FileChunk.SerializeToString,
            file__transfer_dot_ClientFileTransfer__pb2.Response.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)
