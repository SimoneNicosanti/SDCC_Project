# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: file_transfer/EdgeFileTransfer.proto
"""Generated protocol buffer code."""
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import symbol_database as _symbol_database
from google.protobuf.internal import builder as _builder
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()




DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n$file_transfer/EdgeFileTransfer.proto\x12\x04\x65\x64ge\",\n\x17\x45\x64geFileDownloadRequest\x12\x11\n\tfile_name\x18\x01 \x01(\t\"\x1e\n\rEdgeFileChunk\x12\r\n\x05\x63hunk\x18\x01 \x01(\x0c*\xad\x01\n\nErrorCodes\x12\x06\n\x02OK\x10\x00\x12\x0f\n\x0b\x43HUNK_ERROR\x10\n\x12\x12\n\x0eINVALID_TICKET\x10\x0b\x12\x15\n\x11\x46ILE_CREATE_ERROR\x10\x0c\x12\x18\n\x14\x46ILE_NOT_FOUND_ERROR\x10\r\x12\x14\n\x10\x46ILE_WRITE_ERROR\x10\x0f\x12\x13\n\x0f\x46ILE_READ_ERROR\x10\x10\x12\x16\n\x12STREAM_CLOSE_ERROR\x10\x11\x32[\n\x0f\x45\x64geFileService\x12H\n\x10\x44ownloadFromEdge\x12\x1d.edge.EdgeFileDownloadRequest\x1a\x13.edge.EdgeFileChunk0\x01\x42\x0fZ\r../proto/edgeb\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'file_transfer.EdgeFileTransfer_pb2', _globals)
if _descriptor._USE_C_DESCRIPTORS == False:

  DESCRIPTOR._options = None
  DESCRIPTOR._serialized_options = b'Z\r../proto/edge'
  _globals['_ERRORCODES']._serialized_start=125
  _globals['_ERRORCODES']._serialized_end=298
  _globals['_EDGEFILEDOWNLOADREQUEST']._serialized_start=46
  _globals['_EDGEFILEDOWNLOADREQUEST']._serialized_end=90
  _globals['_EDGEFILECHUNK']._serialized_start=92
  _globals['_EDGEFILECHUNK']._serialized_end=122
  _globals['_EDGEFILESERVICE']._serialized_start=300
  _globals['_EDGEFILESERVICE']._serialized_end=391
# @@protoc_insertion_point(module_scope)