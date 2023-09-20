# Generated by the gRPC Python protocol compiler plugin. DO NOT EDIT!
"""Client and server classes corresponding to protobuf-defined services."""
import grpc

from load_balancer import LoadBalancer_pb2 as load__balancer_dot_LoadBalancer__pb2


class BalancingServiceStub(object):
    """Missing associated documentation comment in .proto file."""

    def __init__(self, channel):
        """Constructor.

        Args:
            channel: A grpc.Channel.
        """
        self.GetEdge = channel.unary_unary(
                '/load_balancer.BalancingService/GetEdge',
                request_serializer=load__balancer_dot_LoadBalancer__pb2.User.SerializeToString,
                response_deserializer=load__balancer_dot_LoadBalancer__pb2.BalancerResponse.FromString,
                )
        self.LogClient = channel.unary_unary(
                '/load_balancer.BalancingService/LogClient',
                request_serializer=load__balancer_dot_LoadBalancer__pb2.User.SerializeToString,
                response_deserializer=load__balancer_dot_LoadBalancer__pb2.LoginResponse.FromString,
                )


class BalancingServiceServicer(object):
    """Missing associated documentation comment in .proto file."""

    def GetEdge(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def LogClient(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')


def add_BalancingServiceServicer_to_server(servicer, server):
    rpc_method_handlers = {
            'GetEdge': grpc.unary_unary_rpc_method_handler(
                    servicer.GetEdge,
                    request_deserializer=load__balancer_dot_LoadBalancer__pb2.User.FromString,
                    response_serializer=load__balancer_dot_LoadBalancer__pb2.BalancerResponse.SerializeToString,
            ),
            'LogClient': grpc.unary_unary_rpc_method_handler(
                    servicer.LogClient,
                    request_deserializer=load__balancer_dot_LoadBalancer__pb2.User.FromString,
                    response_serializer=load__balancer_dot_LoadBalancer__pb2.LoginResponse.SerializeToString,
            ),
    }
    generic_handler = grpc.method_handlers_generic_handler(
            'load_balancer.BalancingService', rpc_method_handlers)
    server.add_generic_rpc_handlers((generic_handler,))


 # This class is part of an EXPERIMENTAL API.
class BalancingService(object):
    """Missing associated documentation comment in .proto file."""

    @staticmethod
    def GetEdge(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/load_balancer.BalancingService/GetEdge',
            load__balancer_dot_LoadBalancer__pb2.User.SerializeToString,
            load__balancer_dot_LoadBalancer__pb2.BalancerResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def LogClient(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/load_balancer.BalancingService/LogClient',
            load__balancer_dot_LoadBalancer__pb2.User.SerializeToString,
            load__balancer_dot_LoadBalancer__pb2.LoginResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)
