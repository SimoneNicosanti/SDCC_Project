# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: load_balancer/LoadBalancer.proto
"""Generated protocol buffer code."""
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import symbol_database as _symbol_database
from google.protobuf.internal import builder as _builder
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()




DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n load_balancer/LoadBalancer.proto\x12\rload_balancer\"\x1f\n\rLoginResponse\x12\x0e\n\x06logged\x18\x01 \x01(\x08\"(\n\x04User\x12\x10\n\x08username\x18\x01 \x01(\t\x12\x0e\n\x06passwd\x18\x02 \x01(\t\"9\n\x10\x42\x61lancerResponse\x12\x12\n\nedgeIpAddr\x18\x01 \x01(\t\x12\x11\n\trequestId\x18\x02 \x01(\t*9\n\x16\x45rrorCodesLoadBalancer\x12\x06\n\x02OK\x10\x00\x12\x17\n\x13NO_SERVER_AVAILABLE\x10\x14\x32\x95\x01\n\x10\x42\x61lancingService\x12?\n\x07GetEdge\x12\x13.load_balancer.User\x1a\x1f.load_balancer.BalancerResponse\x12@\n\x0bLoginClient\x12\x13.load_balancer.User\x1a\x1c.load_balancer.LoginResponseB\x18Z\x16../proto/load_balancerb\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'load_balancer.LoadBalancer_pb2', _globals)
if _descriptor._USE_C_DESCRIPTORS == False:
  DESCRIPTOR._options = None
  DESCRIPTOR._serialized_options = b'Z\026../proto/load_balancer'
  _globals['_ERRORCODESLOADBALANCER']._serialized_start=185
  _globals['_ERRORCODESLOADBALANCER']._serialized_end=242
  _globals['_LOGINRESPONSE']._serialized_start=51
  _globals['_LOGINRESPONSE']._serialized_end=82
  _globals['_USER']._serialized_start=84
  _globals['_USER']._serialized_end=124
  _globals['_BALANCERRESPONSE']._serialized_start=126
  _globals['_BALANCERRESPONSE']._serialized_end=183
  _globals['_BALANCINGSERVICE']._serialized_start=245
  _globals['_BALANCINGSERVICE']._serialized_end=394
# @@protoc_insertion_point(module_scope)
